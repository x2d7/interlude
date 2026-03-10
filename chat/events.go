package chat

import (
	"encoding/json"
	"errors"
)

// eventType represents the type of event
type eventType string

const (
	// events produced by the text completion

	eventToken           eventType = "token"
	eventToolCall        eventType = "tool_call"
	eventToolCallToken   eventType = "tool_call_token"
	eventRefusal         eventType = "refusal"
	eventCompletionStart eventType = "completion_start"
	eventCompletionEnded eventType = "completion_ended"

	// events produced by consumer

	eventToolCallResolved eventType = "tool_call_resolved"
	eventUserMessage      eventType = "user_message"
	eventAssistantMessage eventType = "assistant_message"
	eventSystemMessage    eventType = "system_message"
	eventToolMessage      eventType = "tool_message"

	// error event

	eventError eventType = "error"
)

// TODO: Добавить в будущем возможность класть метадату в события (учет стоимости, айди генерации)
// TODO: Добавить возмжность добавлять Name к событиям сообщений (четкое разделение отправителей)

// EventBase is a base type for simple event types
type EventBase struct {
	Content string `json:"text"`
}

// EventCompletionStart represents a completion start event
type EventCompletionStart struct{}

func (e EventCompletionStart) getType() eventType { return eventCompletionStart }

func NewEventCompletionStart() EventCompletionStart {
	return EventCompletionStart{}
}

// EventCompletionEnded represents a completion ended event
// TODO: добавить finish_reason
// TODO: добавить список вызовов инструментов
type EventCompletionEnded struct{}

func (e EventCompletionEnded) getType() eventType { return eventCompletionEnded }

func NewEventCompletionEnded() EventCompletionEnded {
	return EventCompletionEnded{}
}

// EventToolCall represents a tool call event
type EventToolCall struct {
	EventBase
	CallID string `json:"call_id"`
	Name   string `json:"name"`

	approval *ApproveWaiter
	answered bool

	onResolved func(callID string, accepted bool)
}

func (e *EventToolCall) Resolve(accept bool) {
	if e.answered {
		return
	}
	e.answered = true
	if e.approval == nil {
		return
	}

	if e.onResolved != nil {
		e.onResolved(e.CallID, accept)
	}

	verdict := Verdict{Accepted: accept, call: *e}
	e.approval.Resolve(verdict)
}

func (e EventToolCall) getType() eventType { return eventToolCall }

// NewEventToolCall creates a new EventToolCall
func NewEventToolCall(callID, name string, arguments string) EventToolCall {
	return EventToolCall{EventBase: EventBase{Content: arguments}, CallID: callID, Name: name}
}

// EventToolCallResolved spawns when tool call is resolved by the user
type EventToolCallResolved struct {
	CallID   string `json:"call_id"`
	Accepted bool   `json:"accepted"`
}

func (e EventToolCallResolved) getType() eventType { return eventToolCallResolved }

func NewEventToolCallResolved(callID string, accepted bool) EventToolCallResolved {
	return EventToolCallResolved{CallID: callID, Accepted: accepted}
}

// EventToolCallToken represents a tool call token event
// Used for streaming. Not supported by some providers
type EventToolCallToken struct {
	EventBase
	CallID string `json:"call_id"`
	Name   string `json:"name"`
}

func (e EventToolCallToken) getType() eventType { return eventToolCallToken }

// NewEventToolCallToken creates a new EventToolCallToken
func NewEventToolCallToken(callID, name string, token string) EventToolCallToken {
	return EventToolCallToken{EventBase: EventBase{Content: token}, CallID: callID, Name: name}
}

// EventError represents an error event
type EventError struct {
	Error error
}

func (e EventError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Message string `json:"error"`
	}{
		Message: e.Error.Error(),
	})
}

func (e *EventError) UnmarshalJSON(data []byte) error {
	var v struct {
		Message string `json:"error"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	e.Error = errors.New(v.Message)
	return nil
}

func (e EventError) getType() eventType { return eventError }

// NewEventError creates a new EventError
func NewEventError(err error) EventError {
	return EventError{Error: err}
}

// EventToken represents a token event
type EventToken struct {
	EventBase
}

func (e EventToken) getType() eventType { return eventToken }

// NewEventToken creates a new EventToken
func NewEventToken(content string) EventToken {
	return EventToken{EventBase: EventBase{Content: content}}
}

// EventUserMessage represents a user message event
type EventUserMessage struct {
	EventBase
}

func (e EventUserMessage) getType() eventType { return eventUserMessage }

// NewEventUserMessage creates a new EventUserMessage
func NewEventUserMessage(content string) EventUserMessage {
	return EventUserMessage{EventBase: EventBase{Content: content}}
}

// EventAssistantMessage represents an assistant message event
type EventAssistantMessage struct {
	EventBase
}

func (e EventAssistantMessage) getType() eventType { return eventAssistantMessage }

// NewEventAssistantMessage creates a new EventAssistantMessage
func NewEventAssistantMessage(content string) EventAssistantMessage {
	return EventAssistantMessage{EventBase: EventBase{Content: content}}
}

// EventSystemMessage represents a system message event
type EventSystemMessage struct {
	EventBase
}

func (e EventSystemMessage) getType() eventType { return eventSystemMessage }

// NewEventSystemMessage creates a new EventSystemMessage
func NewEventSystemMessage(content string) EventSystemMessage {
	return EventSystemMessage{EventBase: EventBase{Content: content}}
}

// EventToolMessage represents a tool message event
type EventToolMessage struct {
	EventBase
	CallID  string `json:"call_id"`
	Success bool   `json:"success"`
}

func (e EventToolMessage) getType() eventType { return eventToolMessage }

// NewEventToolMessage creates a new EventToolMessage
func NewEventToolMessage(callID, content string, success bool) EventToolMessage {
	return EventToolMessage{EventBase: EventBase{Content: content}, CallID: callID, Success: success}
}

// EventRefusal represents a refusal event
type EventRefusal struct {
	EventBase
}

func (e EventRefusal) getType() eventType { return eventRefusal }

// NewEventRefusal creates a new EventRefusal
func NewEventRefusal(content string) EventRefusal {
	return EventRefusal{EventBase: EventBase{Content: content}}
}

// StreamEvent represents a stream event
type StreamEvent interface {
	getType() eventType
}

// TODO: Remove deprecated types in v0.4

// Deprecated: Use EventToken instead.
type EventNewToken = EventToken

// Deprecated: Use EventToolCall instead.
type EventNewToolCall = EventToolCall

// Deprecated: Use EventToolMessage instead.
type EventNewToolMessage = EventToolMessage

// Deprecated: Use EventRefusal instead.
type EventNewRefusal = EventRefusal

// Deprecated: Use EventCompletionEnded instead.
type EventNewCompletionEnded = EventCompletionEnded

// Deprecated: Use EventError instead.
type EventNewError = EventError

// Deprecated: Use EventUserMessage instead.
type EventNewUserMessage = EventUserMessage

// Deprecated: Use EventAssistantMessage instead.
type EventNewAssistantMessage = EventAssistantMessage

// Deprecated: Use EventSystemMessage instead.
type EventNewSystemMessage = EventSystemMessage
