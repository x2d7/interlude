package types

import (
	"context"

	"github.com/openai/openai-go/v3/packages/ssestream"
)

// Client interface represents the LLM connector client
type Client interface {
	NewStreaming(ctx context.Context) *ssestream.Stream[StreamEvent]
	SyncInput(chat *Chat)
}

// eventType represents the type of event
type eventType uint

const (
	// events produced by the text completion
	eventNewToken eventType = iota
	eventNewToolCall

	// events produced by consumer
	eventNewUserMessage
	eventNewAssistantMessage
	eventNewSystemMessage
	eventNewToolMessage

	// error event
	eventNewError
)

// StreamEvent represents a stream event
type StreamEvent interface {
	GetType() eventType
}

// EventNewToken represents a new token event
type EventNewToken struct {
	Token string
}

func (e EventNewToken) GetType() eventType { return eventNewToken }

// EventNewToolCall represents a new tool call event
type EventNewToolCall struct {
	CallID  string
	RawJSON string
}

func (e EventNewToolCall) GetType() eventType { return eventNewToolCall }

// EventNewUserMessage represents a new user message event
type EventNewUserMessage struct {
	Message string
}

func (e EventNewUserMessage) GetType() eventType { return eventNewUserMessage }

// EventNewAssistantMessage represents a new assistant message event
type EventNewAssistantMessage struct {
	Message string
}

func (e EventNewAssistantMessage) GetType() eventType { return eventNewAssistantMessage }

// EventNewSystemMessage represents a new system message event
type EventNewSystemMessage struct {
	Message string
}

func (e EventNewSystemMessage) GetType() eventType { return eventNewSystemMessage }

// EventNewToolMessage represents a new tool message event
type EventNewToolMessage struct {
	Message string
}

func (e EventNewToolMessage) GetType() eventType { return eventNewToolMessage }

// EventNewError represents a new error event
type EventNewError struct {
	Error string
}

func (e EventNewError) GetType() eventType { return eventNewError }
