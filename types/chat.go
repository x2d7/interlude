package types

import "sync"

// SenderType represents the type of sender of a message in the chat
type SenderType uint

const (
	SenderTypeAssistant SenderType = iota
	SenderTypeSystem
	SenderTypeTool
	SenderTypeUser
)

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

// Chat is a struct that contains messages and tools for text completion
type Chat struct {
	Messages *Messages
	Tools    Tools
}
