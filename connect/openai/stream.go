package openai_connect

import (
	"context"

	"github.com/openai/openai-go/v3"
	"github.com/x2d7/interlude/chat"
)

// sseStreamer is an interface for SSE streams, used to allow mocking in tests.
type sseStreamer interface {
	Next() bool
	Current() openai.ChatCompletionChunk
	Err() error
	Close() error
}

// OpenAIStream is a wrapper for OpenAI SSEStream
//
// Implements types.Stream interface
type OpenAIStream struct {
	queue []chat.StreamEvent
	err   error
	cur   chat.StreamEvent

	OpenAIClient *OpenAIClient
	SSEStream    sseStreamer
}

func (s *OpenAIStream) Next(ctx context.Context) bool {
	if s.err != nil {
		return false
	}

	// сheck context cancellation before trying to get next chunk
	select {
	case <-ctx.Done():
		s.err = ctx.Err()
		return false
	default:
	}

	// creating a queue if it's empty
	if len(s.queue) == 0 {
		if proceed := s.SSEStream.Next(); proceed {
			// parsing events
			queue, err := s.handleRawChunk(s.SSEStream.Current())
			if err != nil {
				s.err = err
				return false
			}

			// skip empty chunks and try next one
			if len(queue) == 0 {
				return s.Next(ctx)
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

func (s *OpenAIStream) Current() chat.StreamEvent {
	return s.cur
}

func (s *OpenAIStream) Err() error {
	return s.err
}

func (s *OpenAIStream) Close() error {
	return s.SSEStream.Close()
}

// embedded decorator for _handleRawChunk
func (s *OpenAIStream) handleRawChunk(chunk openai.ChatCompletionChunk) ([]chat.StreamEvent, error) {
	events, err := s._handleRawChunk(chunk)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// TODO: support more than 1 completion

// _handleRawChunk extracts list of events from raw openai chunk
//
// Should not return empty list. It would be considered as an error
func (s *OpenAIStream) _handleRawChunk(chunk openai.ChatCompletionChunk) ([]chat.StreamEvent, error) {
	result := make([]chat.StreamEvent, 0)
	if len(chunk.Choices) == 0 {
		return result, nil
	}
	choice := chunk.Choices[0]

	delta := choice.Delta

	content := delta.Content
	refusal := delta.Refusal
	tools := delta.ToolCalls

	// TODO: Использование FinishReason
	_ = choice.FinishReason

	if content != "" {
		result = append(result, chat.NewEventNewToken(content))
	}

	if refusal != "" {
		result = append(result, chat.NewEventNewRefusal(refusal))
	}

	for _, tool := range tools {
		name := tool.Function.Name
		arguments := tool.Function.Arguments
		result = append(result, chat.NewEventNewToolCall(tool.ID, name, arguments))
	}

	return result, nil
}
