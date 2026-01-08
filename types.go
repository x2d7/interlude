package interlude

import (
	"github.com/openai/openai-go/v3"
)

type EventType uint

const (
	EventNewToken EventType = iota
	EventPause

	EventNewUserMessage
	EventNewAssistantMessage
	EventNewSystemMessage
	EventNewToolCall
	EventNewToolMessage
)

type SenderType uint

const (
	SenderNoSender SenderType = iota
	SenderTypeAssistant
	SenderTypeSystem
	SenderTypeTool
	SenderTypeUser
)

type Chat struct {
	Client       *openai.Client
	ModelOptions ModelOptions
}

type ModelOptions struct {
	MainModel ModelConnector
}

type ModelConnector struct {
	BaseURL  string
	Params   openai.ChatCompletionNewParams
	Messages *Messages
	Tools    Tools
}

type EventData struct {
	EventType   EventType
	Data        any
	RawJSON     string
	ToolSuccess bool
}
