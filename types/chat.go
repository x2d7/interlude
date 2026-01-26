package types

import (
	"sync"
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
