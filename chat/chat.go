package chat

import (
	"context"
	"strings"
)

func (c *Chat) Complete(ctx context.Context, client Client) <-chan StreamEvent {
	result := make(chan StreamEvent, 16)

	// sending events to the channel in background
	go func() {
		defer close(result)

		// text completion stream
		stream := client.NewStreaming(ctx)
		if stream == nil {
			result <- NewEventNewError(ErrNilStreaming)
			return
		}
		defer stream.Close()

		for stream.Next() {
			event := stream.Current()

			select {
			case result <- event:
			case <-ctx.Done():
				return
			}
		}

		if err := stream.Err(); err != nil {
			result <- NewEventNewError(err)
		}
	}()

	return result
}

func (c *Chat) Session(ctx context.Context, client Client) <-chan StreamEvent {
	// insert chat context into client input configuration
	client = client.SyncInput(c)

	// creating the channels
	result := make(chan StreamEvent, 16)
	events := c.Complete(ctx, client)

	// event handling
	go func() {
		defer close(result)

		// event collectors
		var stringBuilder strings.Builder
		toolCalls := make([]EventNewToolCall, 0)

		approval := NewApproveWaiter()

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					// adding collected events to the chat (assistant's tokens and tool calls)
					if stringBuilder.Len() != 0 {
						c.AppendEvent(NewEventNewToken(stringBuilder.String()))
					}
					for _, call := range toolCalls {
						c.AppendEvent(call)
					}

					callAmount := len(toolCalls)

					// ending current completion
					result <- NewEventCompletionEnded()

					if callAmount == 0 {
						return
					}

					// initializing approval waiter
					verdicts := approval.Wait(ctx, callAmount)

					// processing user verdicts
					for verdict := range verdicts {
						call := verdict.call

						if verdict.Accepted {
							callResult, success := c.Tools.Execute(call.Name, call.Content)
							c.AppendEvent(NewEventNewToolMessage(call.CallID, callResult, success))
						} else {
							c.AppendEvent(NewEventNewToolMessage(call.CallID, "User declined the tool call", false))
						}
					}

					// reset collectors
					stringBuilder.Reset()
					toolCalls = make([]EventNewToolCall, 0)

					// reset approval waiter
					approval = NewApproveWaiter()

					// resume text completion
					client = client.SyncInput(c)
					events = c.Complete(ctx, client)
					continue
				}

				// in case of ev changes inside "collecting events" block
				var modifiedEvent StreamEvent

				// collecting events
				switch event := ev.(type) {
				case EventNewToken:
					stringBuilder.WriteString(event.Content)
				case EventNewToolCall:
					approval.Attach(&event)
					modifiedEvent = event
					toolCalls = append(toolCalls, event)
				case EventNewRefusal:
					c.AppendEvent(event)
				}

				// modifying event
				if modifiedEvent != nil {
					ev = modifiedEvent
				}

				// sending events to the channel
				select {
				case result <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return result
}
