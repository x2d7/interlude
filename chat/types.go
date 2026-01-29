package chat

import (
	"context"
	"sync"

	"github.com/x2d7/interlude/chat/tools"
)

// Chat is a struct that contains messages and tools for text completion
type Chat struct {
	Messages *Messages
	Tools    tools.Tools
}

// Client interface represents the LLM connector client
type Client interface {
	// NewStreaming returns a new streaming client instance
	NewStreaming(ctx context.Context) Stream[StreamEvent]
	// SyncInput return a copy of the client with updated input configuration (messages, tools, etc.)
	SyncInput(chat *Chat) Client
}

// Messages is a slice of chat events that later get converted to client messages
type Messages struct {
	mu     sync.Mutex
	Events []StreamEvent
}

func NewMessages() *Messages {
	m := &Messages{
		Events: make([]StreamEvent, 0),
	}

	return m
}

func (m *Messages) AddEvent(event StreamEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
}

type Stream[T any] interface {
	Next() bool   // advance; returns false on EOF or error
	Current() T   // the current element; valid only if Last Next() returned true
	Err() error   // non-nil if the stream ended because of an error
	Close() error // release resources, ensure Next() returns false
}
