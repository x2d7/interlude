package interlude

import (
	"fmt"
	"strings"
	"sync"

	"github.com/openai/openai-go/v3"
)

// Messages is a slice of events that later converts to messages for text completion
type Messages struct {
	mu     sync.Mutex
	Events []EventData
}

func NewMessages() *Messages {
	m := &Messages{
		Events: make([]EventData, 0),
	}

	return m
}

func (m *Messages) AddEvent(event EventData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
}

// OpenAIMessages is a slice of openai.ChatCompletionMessageParamUnion (messages for text completion)
type OpenAIMessages []openai.ChatCompletionMessageParamUnion

// SendFunction is a function that generates a OpenAI message (openai.ChatCompletionMessageParamUnion)
type SendFunction func(content string) openai.ChatCompletionMessageParamUnion

// ToOpenAI converts Messages to OpenAIMessages
func ToOpenAI(m *Messages) (result OpenAIMessages, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		senderFunc SendFunction
		events     []EventData
	)

	events = m.Events
	result = make(OpenAIMessages, 0, len(events))

	message := strings.Builder{}
	for _, ev := range events {
		eventType := ev.EventType
		rawData := ev.Data

		// --- handle tool call events specially ---
		if eventType == EventNewToolCall {
			message := openai.ChatCompletionMessageParamUnion{}
			if ev.RawJSON == "" {
				continue
			}
			message.UnmarshalJSON([]byte(ev.RawJSON))
			result = append(result, message)
			continue
		}

		// IF block -> messages, ELSE block -> new assistant tokens
		data, ok := rawData.(string)
		if !ok {
			data = fmt.Sprintf("%v", rawData)
		}

		if eventType != EventNewToken {
			if message.Len() > 0 {
				result = append(result, openai.AssistantMessage(message.String()))
				message.Reset()
			}

			if len(data) == 0 {
				continue
			}
			switch eventType {
			case EventNewUserMessage:
				senderFunc = openai.UserMessage
			case EventNewAssistantMessage:
				senderFunc = openai.AssistantMessage
			case EventNewSystemMessage:
				senderFunc = openai.SystemMessage
			case EventNewToolMessage:
				senderFunc = EmbeddedIDToolMessage
			default:
				err = fmt.Errorf("unknown event type: %d", eventType)
				return
			}

			result = append(result, senderFunc(data))
		} else {
			message.WriteString(data)
		}
	}

	if message.Len() > 0 {
		result = append(result, openai.AssistantMessage(message.String()))
	}

	return
}

// EmbeddedIDToolMessage takes a string content and returns an openai.ChatCompletionMessageParamUnion
// object. The string content is split by ">" (1 time), the first part is ID, the second part is the message.
// The function returns an openai.ToolMessage object where the ID is the tool name and the message is the tool content.
func EmbeddedIDToolMessage(content string) openai.ChatCompletionMessageParamUnion {
	// split string by ">" (1 time), the first part is ID, the second part is the message
	parts := strings.SplitN(content, ">", 2)
	return openai.ToolMessage(parts[1], parts[0])
}
