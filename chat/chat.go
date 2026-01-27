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
		// text completion stream
		stream := client.NewStreaming(ctx)
		defer func() { _ = stream.Close() }()

		// event collectors
		var stringBuilder strings.Builder
		toolCalls := make([]types.EventNewToolCall, 0)

		ok := true
		for {
			ok = stream.Next()
			if !ok {
				break
			}
			chunk := stream.Current()

			// collecting events
			switch event := chunk.(type) {
			case types.EventNewToken:
				stringBuilder.WriteString(event.Content)
			case types.EventNewToolCall:
				toolCalls = append(toolCalls, event)
			}
			
			result <- chunk
		}

		// adding collected events to the chat
		c.AppendEvent(types.EventNewAssistantMessage{Content: stringBuilder.String()})
		for _, call := range toolCalls {
			c.AppendEvent(call)
		}

		if stream.Err() != nil {
			result <- types.EventNewError{Error: stream.Err()}
		}

		close(result)
	}()

	return result
}
