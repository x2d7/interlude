package chat

import (
	"context"
	"strings"

	"github.com/x2d7/interlude/types"
)

func (c *Chat) Complete(ctx context.Context, client Client) chan types.StreamEvent {
	result := make(chan types.StreamEvent)

	// insert chat context into client input configuration
	client.SyncInput(c)

	// sending events to the channel in background
	go func() {
		// event collectors
		var stringBuilder strings.Builder
		toolCalls := make([]types.EventNewToolCall, 0)

		// text completion stream
		stream := client.NewStreaming(ctx)
		if stream == nil {
			result <- types.EventNewError{Error: ErrNilStreaming}
			close(result)
			return
		}
		defer stream.Close()

		for stream.Next() {
			chunk := stream.Current()

			// collecting events
			switch event := chunk.(type) {
			case types.EventNewToken:
				stringBuilder.WriteString(event.Content)
			case types.EventNewToolCall:
				toolCalls = append(toolCalls, event)
			}

			select {
			case result <- chunk:
			case <-ctx.Done():
				return
			}
		}

		// adding collected events to the chat
		c.AppendEvent(types.EventNewAssistantMessage{Content: stringBuilder.String()})
		for _, call := range toolCalls {
			c.AppendEvent(call)
		}

		if err := stream.Err(); err != nil {
			result <- types.EventNewError{Error: err}
		}

		close(result)
	}()

	return result
}
