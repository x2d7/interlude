package chat

import (
	"context"

	"github.com/x2d7/interlude/types"
)

func (c *Chat) AddMessage(sender types.Sender, content string) error {
	var newEvent types.StreamEvent

	switch s := sender.(type) {
	case types.SenderUser:
		newEvent = types.EventNewUserMessage{Content: content}
	case types.SenderAssistant:
		newEvent = types.EventNewAssistantMessage{Content: content}
	case types.SenderSystem:
		newEvent = types.EventNewSystemMessage{Content: content}
	case types.SenderTool:
		newEvent = types.EventNewToolMessage{Content: content}
	case types.SenderToolCaller:
		newEvent = types.EventNewToolCall{CallID: s.CallId, RawJSON: content}
	default:
		return ErrUnsupportedSender
	}

	c.AppendEvent(newEvent)

	return nil
}

func (c *Chat) AppendEvent(event types.StreamEvent) {
	c.Messages.Events = append(c.Messages.Events, event)
}

func (c *Chat) SendStream(ctx context.Context, client Client, sender types.Sender, content string) chan types.StreamEvent {
	err := c.AddMessage(sender, content)
	if err != nil {
		result := make(chan types.StreamEvent, 1)
		result <- types.EventNewError{Error: err}
		close(result)
		return result
	}

	return c.Complete(ctx, client)
}

func (c *Chat) SendUserStream(ctx context.Context, client Client, content string) chan types.StreamEvent {
	return c.SendStream(ctx, client, types.SenderUser{}, content)
}

func (c *Chat) SendAssistantStream(ctx context.Context, client Client, content string) chan types.StreamEvent {
	return c.SendStream(ctx, client, types.SenderAssistant{}, content)
}

func (c *Chat) SendSystemStream(ctx context.Context, client Client, content string) chan types.StreamEvent {
	return c.SendStream(ctx, client, types.SenderSystem{}, content)
}

func (c *Chat) SendToolStream(ctx context.Context, client Client, content string) chan types.StreamEvent {
	return c.SendStream(ctx, client, types.SenderTool{}, content)
}

func (c *Chat) SendToolCallStream(ctx context.Context, client Client, callId string, content string) chan types.StreamEvent {
	return c.SendStream(ctx, client, types.SenderToolCaller{CallId: callId}, content)
}
