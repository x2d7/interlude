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

		for stream.Next(ctx) {
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

	// delivers a StreamEvent to the result channel
	// skips nil events
	send := func(event StreamEvent) bool {
		if event == nil {
			if ctx.Err() != nil {
				return false
			}
			return true
		}
		select {
		case result <- event:
			return true
		case <-ctx.Done():
			return false
		}
	}

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

					// send every call from the queue
					for _, call := range toolCalls {
						if !send(call) {
							return
						}
					}

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

				// in case if we need to skip event
				var skipEvent bool

				// collecting events
				switch event := ev.(type) {
				case EventNewToken:
					stringBuilder.WriteString(event.Content)
				case EventNewToolCall:
					// prevent adding tool call immediately — we need to wait until end of completion
					skipEvent = true
					// if callid is present — it's the start of a new tool call
					if event.CallID != "" {
						approval.Attach(&event)
						toolCalls = append(toolCalls, event)
					} else {
						// add token to the last tool call
						toolCalls[len(toolCalls)-1].Content += event.Content
					}
				case EventNewRefusal:
					c.AppendEvent(event)
				}

				// skipping event
				if skipEvent {
					continue
				}

				// sending events to the channel
				if !send(ev) {
					return
				}
			}
		}
	}()

	return result
}
