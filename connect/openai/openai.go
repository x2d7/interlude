package openai_connect

import (
	"context"

	"github.com/x2d7/interlude/chat"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIClient struct {
	Endpoint string
	APIKey   string

	Params openai.ChatCompletionNewParams

	RequestOptions []option.RequestOption
}

func (c *OpenAIClient) NewStreaming(ctx context.Context) chat.Stream[chat.StreamEvent] {
	stream := &OpenAIStream{
		OpenAIClient: c,
	}

	client := getClient(c)
	params := c.Params

	stream.SSEStream = client.Chat.Completions.NewStreaming(ctx, params)
	return stream
}

// TODO: implement input configuration syncronization for OpenAI client
func (c *OpenAIClient) SyncInput(chat *chat.Chat) chat.Client {
	newClient := *c
	return &newClient
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
