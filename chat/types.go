package chat

import (
	"context"
	"sync"

	"github.com/x2d7/interlude/chat/tools"
)

// Chat is a struct that contains messages and tools for text completion
type Chat struct {
	Messages *Messages
	Tools    *tools.Tools
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

func (m *Messages) Snapshot() []StreamEvent {
	m.mu.Lock()
	defer m.mu.Unlock()

	cp := make([]StreamEvent, len(m.Events))
	copy(cp, m.Events)
	return cp
}

type Stream[T any] interface {
	Next() bool   // advance; returns false on EOF or error
	Current() T   // the current element; valid only if Last Next() returned true
	Err() error   // non-nil if the stream ended because of an error
	Close() error // release resources, ensure Next() returns false
}

type Verdict struct {
	Accepted bool
	call     EventNewToolCall
}

type ApproveWaiter struct {
	verdicts chan Verdict
}

func NewApproveWaiter() *ApproveWaiter {
	return &ApproveWaiter{
		verdicts: make(chan Verdict),
	}
}

// Attach wires the event to the waiter
// Call this on the local `event` value before appending it to toolCalls.
func (a *ApproveWaiter) Attach(e *EventNewToolCall) {
	e.approval = a
}

// Wait returns a channel that will deliver exactly `amount` verdicts (or close early on ctx cancel).
func (a *ApproveWaiter) Wait(ctx context.Context, amount int) chan Verdict {
	out := make(chan Verdict)
	if amount <= 0 {
		close(out)
		return out
	}

	go func() {
		defer func() {
			close(out)
		}()

		collected := 0
		for collected < amount {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-a.verdicts:
				if !ok {
					return
				}
				select {
				case out <- v:
					collected++
				case <-ctx.Done():
					return
				}
			}
		}

	}()

	return out
}

// Resolve allows programmatic submission of a verdict.
func (a *ApproveWaiter) Resolve(verdict Verdict) {
	select {
	case a.verdicts <- verdict:
	default:
		go func() { a.verdicts <- verdict }()
	}
}
