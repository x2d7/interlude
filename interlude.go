package interlude

import (
	"context"
	"fmt"

	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)


func MessagesEmpty() *Messages {
	m := &Messages{
		Events: make([]EventData, 0),
	}

	return m
}

// EmbeddedIDToolMessage takes a string content and returns an openai.ChatCompletionMessageParamUnion
// object. The string content is split by ">" (1 time), the first part is ID, the second part is the message.
// The function returns an openai.ToolMessage object where the ID is the tool name and the message is the tool content.
func EmbeddedIDToolMessage(content string) openai.ChatCompletionMessageParamUnion {
	// split string by ">" (1 time), the first part is ID, the second part is the message
	parts := strings.SplitN(content, ">", 2)
	return openai.ToolMessage(parts[1], parts[0])
}

func (m *Messages) AddEvent(event EventData) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Events = append(m.Events, event)
}

// convert converts Messages to OpenAIMessages and supports tool calls.
func convert(m *Messages) (result OpenAIMessages, err error) {
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

func Complete(ctx context.Context, client *openai.Client, connector *ModelConnector, result chan<- EventData) error {
	defer close(result)
	params := connector.Params
	baseurl := connector.BaseURL
	messages, err := convert(connector.Messages)
	if err != nil {
		return err

	}

	params.Messages = messages
	tools := ConvertTools(connector.Tools)
	params.Tools = tools

	stream := client.Chat.Completions.NewStreaming(ctx, params, option.WithBaseURL(baseurl))
	defer func() { _ = stream.Close() }()

	var event EventData

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			event = EventData{EventType: EventNewToken, Data: chunk.Choices[0].Delta.Content}
			connector.Messages.AddEvent(event)
			result <- event
		}

		// TOOL CALLS
		// we need to divide rawjson and calls
		if len(chunk.Choices[0].Delta.ToolCalls) != 0 {
			event = EventData{EventType: EventNewToolCall, RawJSON: chunk.Choices[0].Delta.RawJSON()}
			connector.Messages.AddEvent(event)
		}

		for _, call := range chunk.Choices[0].Delta.ToolCalls {
			event = EventData{EventType: EventNewToolCall, Data: call}
			connector.Messages.AddEvent(event)
			result <- event
		}
	}

	if err := stream.Err(); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func pipe[T any](src <-chan T, dst chan<- T, handler func(T)) {
	for v := range src {
		dst <- v
		if handler != nil {
			go handler(v)
		}
	}
}

func (c Chat) SendStream(ctx context.Context, userMessage string, events chan<- EventData, senderType SenderType) error {
	defer close(events)

	// create new event (user/tool/system input)
	var newEvent EventType

	if senderType != SenderNoSender {
		switch senderType {
		case SenderTypeUser:
			newEvent = EventNewUserMessage
		case SenderTypeAssistant:
			newEvent = EventNewAssistantMessage
		case SenderTypeSystem:
			newEvent = EventNewSystemMessage
		case SenderTypeTool:
			newEvent = EventNewToolMessage
		default:
			return fmt.Errorf("unsupported sender type: %d", senderType)
		}

		c.ModelOptions.MainModel.Messages.AddEvent(EventData{EventType: newEvent, Data: userMessage})
	}

	// proxy channel start
	eventsProxy := make(chan EventData)
	toolCalls := make([]openai.ChatCompletionChunkChoiceDeltaToolCall, 0)
	handler := func(ev EventData) {
		if ev.EventType == EventNewToolCall {
			data := ev.Data.(openai.ChatCompletionChunkChoiceDeltaToolCall)
			toolCalls = append(toolCalls, data)
		}
	}

	go pipe(eventsProxy, events, handler)

	// start completion
	errchan := make(chan error, 1)
	go func() {
		if err := Complete(ctx, c.Client, &c.ModelOptions.MainModel, eventsProxy); err != nil {
			errchan <- err
		}
		if len(toolCalls) > 0 {
			tools := c.ModelOptions.MainModel.Tools
			for _, call := range toolCalls {
				result, ok := tools.Execute(call)

				// add to history
				c.ModelOptions.MainModel.Messages.AddEvent(EventData{EventType: EventNewToolMessage, Data: result, ToolSuccess: ok})
			}
			localEventProxy := make(chan EventData)
			go pipe(localEventProxy, events, nil)
			c.SendStream(ctx, "", localEventProxy, SenderNoSender)
		}
		close(errchan)
	}()

	// wait for an error or completion
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errchan:
			if err != nil {
				return err
			}
			return nil
		}
	}
}

func (c Chat) SendUserStream(ctx context.Context, userMessage string, events chan<- EventData) error {
	return c.SendStream(ctx, userMessage, events, SenderTypeUser)
}

func (c Chat) SendToolStream(ctx context.Context, userMessage string, events chan<- EventData) error {
	return c.SendStream(ctx, userMessage, events, SenderTypeTool)
}
