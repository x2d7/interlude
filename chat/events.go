package chat

// eventType represents the type of event
type eventType string

const (
	// events produced by the text completion

	eventToken           eventType = "token"
	eventToolCall        eventType = "tool_call"
	eventRefusal         eventType = "refusal"
	eventCompletionEnded eventType = "completion_ended"

	// events produced by consumer

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
	Content string
}

type EventCompletionEnded struct{} // TODO: можно добавлять список вызовов инструментов и другую информацию о генерации

func (e EventCompletionEnded) getType() eventType { return eventCompletionEnded }

func NewEventCompletionEnded() EventCompletionEnded {
	return EventCompletionEnded{}
}

type EventToolCall struct {
	EventBase
	CallID string
	Name   string

	approval *ApproveWaiter
	answered bool
}

func (e *EventToolCall) Resolve(accept bool) {
	if e.answered {
		return
	}
	e.answered = true
	if e.approval == nil {
		return
	}
	verdict := Verdict{Accepted: accept, call: *e}
	e.approval.Resolve(verdict)
}

func (e EventToolCall) getType() eventType { return eventToolCall }

// NewEventToolCall creates a new EventToolCall
func NewEventToolCall(callID, name string, arguments string) EventToolCall {
	return EventToolCall{EventBase: EventBase{Content: arguments}, CallID: callID, Name: name}
}

// EventError represents a error event
type EventError struct {
	Error error
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

// EventAssistantMessage represents a assistant message event
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
	// CallID is the ID of the tool call request that was previously sent by assistant
	CallID  string
	Success bool
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
