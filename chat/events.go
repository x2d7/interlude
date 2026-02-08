package chat

// eventType represents the type of event
type eventType uint

const (
	// events produced by the text completion

	eventNewToken eventType = iota
	eventNewToolCall
	eventNewRefusal

	// events produced by consumer

	eventNewUserMessage
	eventNewAssistantMessage
	eventNewSystemMessage
	eventNewToolMessage

	// error event

	eventNewError
)

// TODO: Добавить в будущем возможность класть метадату в события (учет стоимости, айди генерации)
// TODO: Добавить возмжность добавлять Name к событиям

// EventBase is a base type for simple event types
type EventBase struct {
	Content string
}

// EventNewToolCall represents a new tool call event
// EventBase.Content contains raw JSON arguments of the call
type EventNewToolCall struct {
	EventBase
	// CallId is the ID of the tool call request
	CallID string
	// Name is the name of the tool that was called
	Name string
}

func (e EventNewToolCall) GetType() eventType { return eventNewToolCall }

// NewEventNewToolCall creates a new EventNewToolCall
func NewEventNewToolCall(callID, name string, arguments string) EventNewToolCall {
	return EventNewToolCall{EventBase: EventBase{Content: arguments}, CallID: callID, Name: name}
}

// EventNewError represents a new error event
type EventNewError struct {
	Error error
}

func (e EventNewError) GetType() eventType { return eventNewError }

// NewEventNewError creates a new EventNewError
func NewEventNewError(err error) EventNewError {
	return EventNewError{Error: err}
}

// EventNewToken represents a new token event
type EventNewToken struct {
	EventBase
}

func (e EventNewToken) GetType() eventType { return eventNewToken }

// NewEventNewToken creates a new EventNewToken
func NewEventNewToken(content string) EventNewToken {
	return EventNewToken{EventBase: EventBase{Content: content}}
}

// EventNewUserMessage represents a new user message event
type EventNewUserMessage struct {
	EventBase
}

func (e EventNewUserMessage) GetType() eventType { return eventNewUserMessage }

// NewEventNewUserMessage creates a new EventNewUserMessage
func NewEventNewUserMessage(content string) EventNewUserMessage {
	return EventNewUserMessage{EventBase: EventBase{Content: content}}
}

// EventNewAssistantMessage represents a new assistant message event
type EventNewAssistantMessage struct {
	EventBase
}

func (e EventNewAssistantMessage) GetType() eventType { return eventNewAssistantMessage }

// NewEventNewAssistantMessage creates a new EventNewAssistantMessage
func NewEventNewAssistantMessage(content string) EventNewAssistantMessage {
	return EventNewAssistantMessage{EventBase: EventBase{Content: content}}
}

// EventNewSystemMessage represents a new system message event
type EventNewSystemMessage struct {
	EventBase
}

func (e EventNewSystemMessage) GetType() eventType { return eventNewSystemMessage }

// NewEventNewSystemMessage creates a new EventNewSystemMessage
func NewEventNewSystemMessage(content string) EventNewSystemMessage {
	return EventNewSystemMessage{EventBase: EventBase{Content: content}}
}

// EventNewToolMessage represents a new tool message event
type EventNewToolMessage struct {
	EventBase
	// CallID is the ID of the tool call request that was previously sent by assistant
	CallID string
}

func (e EventNewToolMessage) GetType() eventType { return eventNewToolMessage }

// NewEventNewToolMessage creates a new EventNewToolMessage
func NewEventNewToolMessage(content string) EventNewToolMessage {
	return EventNewToolMessage{EventBase: EventBase{Content: content}}
}

// EventNewRefusal represents a new refusal event
type EventNewRefusal struct {
	EventBase
}

func (e EventNewRefusal) GetType() eventType { return eventNewRefusal }

// NewEventNewRefusal creates a new EventNewRefusal
func NewEventNewRefusal(content string) EventNewRefusal {
	return EventNewRefusal{EventBase: EventBase{Content: content}}
}

// StreamEvent represents a stream event
type StreamEvent interface {
	GetType() eventType
}
