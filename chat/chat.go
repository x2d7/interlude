package chat

import (
	"context"
	"strings"

	"github.com/x2d7/interlude/types"
)

func (c *Chat) Complete(ctx context.Context, client Client) chan types.StreamEvent {
	result := make(chan types.StreamEvent, 16)

	// sending events to the channel in background
	go func() {
		defer close(result)

		// text completion stream
		stream := client.NewStreaming(ctx)
		if stream == nil {
			result <- types.EventNewError{Error: ErrNilStreaming}
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
			result <- types.EventNewError{Error: err}
		}
	}()

	return result
}

func (c *Chat) Session(ctx context.Context, client Client) chan types.StreamEvent {
	// insert chat context into client input configuration
	client.SyncInput(c)

	// creating the channels
	result := make(chan types.StreamEvent, 16)
	events := c.Complete(ctx, client)

	// event handling
	go func() {
		defer close(result)

		// event collectors
		var stringBuilder strings.Builder
		toolCalls := make([]types.EventNewToolCall, 0)

		// adding collected events to the chat (assistant's tokens and tool calls)
		defer func() {
			if stringBuilder.Len() != 0 {
				c.AppendEvent(types.EventNewAssistantMessage{Content: stringBuilder.String()})
			}
			for _, call := range toolCalls {
				c.AppendEvent(call)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				// closing the result channel if the events channel is closed
				// TODO: in future expected behavior will different: result channel will be closed if all tool calls are processed and text completion is done
				if !ok { 
					return
				}

				// collecting events
				switch event := ev.(type) {
				case types.EventNewToken:
					stringBuilder.WriteString(event.Content)
				case types.EventNewToolCall:
					toolCalls = append(toolCalls, event)
				case types.EventNewRefusal:
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
