package chat

import (
	"context"

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

		status := true
		for {
			status = stream.Next()
			if !status {
				break
			}
			chunk := stream.Current()
			result <- chunk
		}

		if stream.Err() != nil {
			event := types.EventNewError{Error: stream.Err()}
			result <- event
		}

		close(result)
	}()

	return result
}
