package interlude

import "context"

func (c Chat) SendUserStream(ctx context.Context, userMessage string, events chan<- EventData) error {
	return c.SendStream(ctx, userMessage, events, SenderTypeUser)
}

func (c Chat) SendToolStream(ctx context.Context, userMessage string, events chan<- EventData) error {
	return c.SendStream(ctx, userMessage, events, SenderTypeTool)
}

func pipe[T any](src <-chan T, dst chan<- T, handler func(T)) {
	for v := range src {
		dst <- v
		if handler != nil {
			go handler(v)
		}
	}
}