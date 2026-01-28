package openai_connect

import (
	"context"

	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/types"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIClient struct {
	Endpoint string
	APIKey   string
	Model    string
} // TODO: drastically improve config

func (c *OpenAIClient) NewStreaming(ctx context.Context) types.Stream[types.StreamEvent] {
	stream := &OpenAIStream{
		OpenAIClient: c,
	}

	// TODO: better SSEStream composition
	client := openai.NewClient(option.WithAPIKey(c.APIKey))
	params := openai.ChatCompletionNewParams{
		Model: c.Model,
	}

	stream.SSEStream = client.Chat.Completions.NewStreaming(ctx, params, option.WithBaseURL(c.Endpoint))
	return stream
}

func (c *OpenAIClient) SyncInput(chat *chat.Chat) {
}
