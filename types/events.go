package types

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

// eventNewContent is a base type for simple event types
type eventNewContent struct {
	Content string
}

// EventNewToolCall represents a new tool call event
type EventNewToolCall struct {
	CallID  string
	RawJSON string
}

func (e EventNewToolCall) GetType() eventType { return eventNewToolCall }

// EventNewError represents a new error event
type EventNewError struct {
	Error error
}

func (e EventNewError) GetType() eventType { return eventNewError }

// EventNewToken represents a new token event
type EventNewToken eventNewContent

func (e EventNewToken) GetType() eventType { return eventNewToken }

// EventNewUserMessage represents a new user message event
type EventNewUserMessage eventNewContent

func (e EventNewUserMessage) GetType() eventType { return eventNewUserMessage }

// EventNewAssistantMessage represents a new assistant message event
type EventNewAssistantMessage eventNewContent

func (e EventNewAssistantMessage) GetType() eventType { return eventNewAssistantMessage }

// EventNewSystemMessage represents a new system message event
type EventNewSystemMessage eventNewContent

func (e EventNewSystemMessage) GetType() eventType { return eventNewSystemMessage }

// EventNewToolMessage represents a new tool message event
type EventNewToolMessage eventNewContent

func (e EventNewToolMessage) GetType() eventType { return eventNewToolMessage }

// StreamEvent represents a stream event
type StreamEvent interface {
	GetType() eventType
}
