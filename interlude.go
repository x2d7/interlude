package interlude

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func Complete(ctx context.Context, client *openai.Client, connector *ModelConnector, result chan<- EventData) error {
	defer close(result)
	params := connector.Params
	baseurl := connector.BaseURL
	messages, err := ToOpenAI(connector.Messages)
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
