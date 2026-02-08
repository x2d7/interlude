package chat

import (
	"context"
	"strings"
)

func (c *Chat) Complete(ctx context.Context, client Client) chan StreamEvent {
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

func (c *Chat) Session(ctx context.Context, client Client) chan StreamEvent {
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

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				// closing the result channel if the events channel is closed
				// TODO: in future expected behavior will different: result channel will be closed if all tool calls are processed and text completion is done
				if !ok {
					result <- NewEventCompletionEnded()

					// adding collected events to the chat (assistant's tokens and tool calls)
					if stringBuilder.Len() != 0 {
						c.AppendEvent(NewEventNewToken(stringBuilder.String()))
					}
					for _, call := range toolCalls {
						c.AppendEvent(call)
					}

					// reset collectors
					stringBuilder.Reset()
					toolCalls = make([]EventNewToolCall, 0)

					return
				}

				// collecting events
				switch event := ev.(type) {
				case EventNewToken:
					stringBuilder.WriteString(event.Content)
				case EventNewToolCall:
					toolCalls = append(toolCalls, event)
				case EventNewRefusal:
					c.AppendEvent(event)
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
