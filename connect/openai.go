package openai_connect

import (
	"context"
	"errors"

	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/types"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
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

// TODO: Добавить в будущем возможность класть метадату в события (учет стоимости, айди генерации)

type OpenAIStream struct {
	queue []types.StreamEvent
	err   error
	cur   types.StreamEvent

	OpenAIClient *OpenAIClient
	SSEStream    *ssestream.Stream[openai.ChatCompletionChunk]
}

func (s *OpenAIStream) Next() bool {
	if s.err != nil {
		return false
	}

	// creating a queue if it's empty
	if len(s.queue) == 0 {
		if proceed := s.SSEStream.Next(); proceed {
			// parsing events
			queue, err := s.handleRawChunk(s.SSEStream.Current())
			// fall if error appears (or empty event list)
			if err != nil {
				s.err = err
				return false
			}

			// updating queue to new parsed events
			s.queue = queue
		} else {
			// put an error if we can't proceed
			s.err = s.SSEStream.Err()
			return false
		}
	}
	// processing queue
	s.cur = s.queue[0]
	s.queue = s.queue[1:]

	return true
}

func (s *OpenAIStream) Current() types.StreamEvent {
	return s.cur
}

func (s *OpenAIStream) Err() error {
	return s.err
}

func (s *OpenAIStream) Close() error {
	return s.SSEStream.Close()
}

func (s *OpenAIStream) handleRawChunk(chunk openai.ChatCompletionChunk) ([]types.StreamEvent, error) {
	events, err := s._handleRawChunk(chunk)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, errors.New("empty events")
	}
	return events, nil
}

func (s *OpenAIStream) _handleRawChunk(_ openai.ChatCompletionChunk) ([]types.StreamEvent, error) {
	return []types.StreamEvent{}, errors.New("not implemented")
}
