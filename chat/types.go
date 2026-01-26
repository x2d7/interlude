package chat

import (
	"context"

	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/x2d7/interlude/types"
)

// Chat is a struct that contains messages and tools for text completion
type Chat struct {
	Messages *types.Messages
	Tools    types.Tools
}

// Client interface represents the LLM connector client
type Client interface {
	NewStreaming(ctx context.Context) *ssestream.Stream[types.StreamEvent]
	SyncInput(chat *Chat)
}
