package openai_connect

import (
	"errors"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/x2d7/interlude/types"
)

// TODO: Добавить в будущем возможность класть метадату в события (учет стоимости, айди генерации)

// OpenAIStream is a wrapper for OpenAI SSEStream
//
// Implements types.Stream interface
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

// embedded decorator for _handleRawChunk
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

// TODO: implement handleRawChunk

// _handleRawChunk extracts list of events from raw openai chunk
//
// Should not return empty list. It would be considered as an error
func (s *OpenAIStream) _handleRawChunk(_ openai.ChatCompletionChunk) ([]types.StreamEvent, error) {
	return []types.StreamEvent{}, errors.New("not implemented")
}
