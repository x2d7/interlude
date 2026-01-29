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

	RequestOptions []option.RequestOption
} // TODO: drastically improve config

func (c *OpenAIClient) NewStreaming(ctx context.Context) types.Stream[types.StreamEvent] {
	stream := &OpenAIStream{
		OpenAIClient: c,
	}

	client := getClient(c)
	params := getParams(c)

	stream.SSEStream = client.Chat.Completions.NewStreaming(ctx, params)
	return stream
}

// TODO: implement input configuration syncronization for OpenAI client
func (c *OpenAIClient) SyncInput(chat *chat.Chat) chat.Client {
	newClient := *c
	return &newClient
}

func getParams(c *OpenAIClient) openai.ChatCompletionNewParams {
	return openai.ChatCompletionNewParams{
		Model: c.Model,
	}
}

func getClient(c *OpenAIClient) *openai.Client {
	requestOptions := make([]option.RequestOption, 0)

	if c.Endpoint != "" {
		requestOptions = append(requestOptions, option.WithBaseURL(c.Endpoint))
	}

	if c.APIKey != "" {
		requestOptions = append(requestOptions, option.WithAPIKey(c.APIKey))
	}

	if c.RequestOptions != nil {
		requestOptions = append(requestOptions, c.RequestOptions...)
	}

	client := openai.NewClient(requestOptions...)
	return &client
}
