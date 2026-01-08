package interlude

import (
	"reflect"
	"sync"

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

// Messages is a slice of events that later converts to messages for text completion
type Messages struct {
	mu     sync.Mutex
	Events []EventData
}

type EventData struct {
	EventType   EventType
	Data        any
	RawJSON     string
	ToolSuccess bool
}

type Tools []Tool

type Tool struct {
	Name        string
	Description string
	Func        ToolFunction

	InputType reflect.Type
	Schema    map[string]any
}

type ToolFunction func(input any) (string, error)

type OpenAIMessages []openai.ChatCompletionMessageParamUnion

type SendFunction func(content string) openai.ChatCompletionMessageParamUnion
