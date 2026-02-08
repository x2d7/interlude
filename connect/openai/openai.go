package openai_connect

import (
	"context"

	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/chat/tools"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAIClient struct {
	Endpoint string
	APIKey   string

	Params openai.ChatCompletionNewParams

	RequestOptions []option.RequestOption
}

func (c *OpenAIClient) NewStreaming(ctx context.Context) chat.Stream[chat.StreamEvent] {
	stream := &OpenAIStream{
		OpenAIClient: c,
	}

	client := getClient(c)
	params := c.Params

	stream.SSEStream = client.Chat.Completions.NewStreaming(ctx, params)
	return stream
}

func (c *OpenAIClient) SyncInput(chat *chat.Chat) chat.Client {
	newClient := *c

	// copy messages
	messages := make(openAIMessages, 0)
	for _, m := range chat.Messages.Snapshot() {
		messages.Add(m)
	}

	newClient.Params.Messages = messages

	tools := ConvertTools(chat.Tools)
	newClient.Params.Tools = tools

	return &newClient
}

type openAIMessages []openai.ChatCompletionMessageParamUnion

func (m *openAIMessages) findLastAssistantMessage() *openai.ChatCompletionMessageParamUnion {
	for i := len(*m) - 1; i >= 0; i-- {
		if (*m)[i].OfAssistant != nil {
			return &(*m)[i]
		}
	}
	return nil
}

func (m *openAIMessages) Add(event chat.StreamEvent) {
	var message openai.ChatCompletionMessageParamUnion

	switch e := event.(type) {
	case chat.EventNewAssistantMessage:
		message = openai.AssistantMessage(e.Content)
	case chat.EventNewRefusal:
		refusal := openai.ChatCompletionContentPartRefusalParam{Refusal: e.Content}
		contentUnion := make([]openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion, 0)
		contentUnion = append(contentUnion,
			openai.ChatCompletionAssistantMessageParamContentArrayOfContentPartUnion{OfRefusal: &refusal},
		)
		message = openai.AssistantMessage(contentUnion)
	case chat.EventNewSystemMessage:
		message = openai.SystemMessage(e.Content)
	case chat.EventNewUserMessage:
		message = openai.UserMessage(e.Content)
	case chat.EventNewToolCall:
		messagePtr := m.findLastAssistantMessage()
		if messagePtr == nil {
			message = openai.AssistantMessage(" ")
			messagePtr = &message
		}

		toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0)
		functionCall := openai.ChatCompletionMessageFunctionToolCallParam{
			ID: e.CallID,
			Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
				Name:      e.Name,
				Arguments: e.Content,
			},
		}
		toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
			OfFunction: &functionCall,
		})
		if messagePtr.OfAssistant.ToolCalls == nil {
			messagePtr.OfAssistant.ToolCalls = make([]openai.ChatCompletionMessageToolCallUnionParam, 0)
		}
		messagePtr.OfAssistant.ToolCalls = append(messagePtr.OfAssistant.ToolCalls, toolCalls...)

	case chat.EventNewToolMessage:
		message = openai.ToolMessage(e.Content, e.CallID)
	}

	if !isEmpty(message) {
		*m = append(*m, message)
	}

}

func isEmpty(m openai.ChatCompletionMessageParamUnion) bool {
	return m.OfDeveloper == nil &&
		m.OfSystem == nil &&
		m.OfUser == nil &&
		m.OfAssistant == nil &&
		m.OfTool == nil &&
		m.OfFunction == nil
}

func ConvertTools(t tools.Tools) []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(t))
	for _, tool := range t {
		if tool.Schema == nil {
			out = append(out, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
			}))
			continue
		}
		out = append(out, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters:  openai.FunctionParameters(tool.Schema),
		}))
	}
	return out
}

func getClient(c *OpenAIClient) *openai.Client {
	requestOptions := make([]option.RequestOption, 0)

	if c.Endpoint != "" {
		requestOptions = append(requestOptions, option.WithBaseURL(c.Endpoint))
	}

	if c.APIKey != "" {
		requestOptions = append(requestOptions, option.WithAPIKey(c.APIKey))
	}

	if c.RequestOptions != nil {
		requestOptions = append(requestOptions, c.RequestOptions...)
	}

	client := openai.NewClient(requestOptions...)
	return &client
}
