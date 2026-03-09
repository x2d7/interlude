package chat

import (
	"context"
)

func (c *Chat) AddMessage(sender Sender, content string) error {
	var newEvent StreamEvent

	switch sender.(type) {
	case SenderUser:
		newEvent = NewEventUserMessage(content)
	case SenderAssistant:
		newEvent = NewEventAssistantMessage(content)
	case SenderSystem:
		newEvent = NewEventSystemMessage(content)
	default:
		return ErrUnsupportedSender
	}

	c.AppendEvent(newEvent)

	return nil
}

// AppendEvent adds a full message event to the chat.
// The event must represent a complete message (e.g., EventUserMessage,
// EventAssistantMessage, EventToolMessage), not intermediate streaming
// events like EventToken. Providers are not required to sync such events
// in their SyncInput implementation.
func (c *Chat) AppendEvent(event StreamEvent) {
	c.Messages.AddEvent(event)
}

func (c *Chat) SendStream(ctx context.Context, client Client, sender Sender, content string) <-chan StreamEvent {
	err := c.AddMessage(sender, content)
	if err != nil {
		result := make(chan StreamEvent, 1)
		result <- NewEventError(err)
		close(result)
		return result
	}

	return c.Session(ctx, client)
}

func (c *Chat) SendUserStream(ctx context.Context, client Client, content string) <-chan StreamEvent {
	return c.SendStream(ctx, client, SenderUser{}, content)
}

func (c *Chat) SendAssistantStream(ctx context.Context, client Client, content string) <-chan StreamEvent {
	return c.SendStream(ctx, client, SenderAssistant{}, content)
}

func (c *Chat) SendSystemStream(ctx context.Context, client Client, content string) <-chan StreamEvent {
	return c.SendStream(ctx, client, SenderSystem{}, content)
}
