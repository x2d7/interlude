package chat

import (
	"context"
	"strings"
)

func (c *Chat) Complete(ctx context.Context, client Client) <-chan StreamEvent {
	result := make(chan StreamEvent, 16)

	// sending events to the channel in background
	go func() {
		defer close(result)

		// text completion stream
		stream := client.NewStreaming(ctx)
		if stream == nil {
			result <- NewEventNewError(ErrNilStreaming)
			return
		}
		defer stream.Close()

		for stream.Next(ctx) {
			event := stream.Current()

			select {
			case result <- event:
			case <-ctx.Done():
				return
			}
		}

		if err := stream.Err(); err != nil {
			result <- NewEventNewError(err)
		}
	}()

	return result
}

type sessionState struct {
	builder      strings.Builder
	toolCalls    []EventNewToolCall
	lastToolCall *EventNewToolCall
	approval     *ApproveWaiter
}

func (s *sessionState) reset() {
	s.builder.Reset()
	s.toolCalls = s.toolCalls[:0]
	s.lastToolCall = nil
	s.approval = NewApproveWaiter()
}

func (s *sessionState) flushLastToolCall(send func(StreamEvent) bool) bool {
    if s.lastToolCall == nil {
        return true
    }
    ok := send(*s.lastToolCall)
    s.lastToolCall = nil
    return ok
}

func (c *Chat) Session(ctx context.Context, client Client) <-chan StreamEvent {
	// insert chat context into client input configuration
	client = client.SyncInput(c)

	// creating the channels
	result := make(chan StreamEvent, 16)
	events := c.Complete(ctx, client)

	// delivers a StreamEvent to the result channel
	// skips nil events
	send := func(event StreamEvent) bool {
		if event == nil {
			if ctx.Err() != nil {
				return false
			}
			return true
		}
		select {
		case result <- event:
			return true
		case <-ctx.Done():
			return false
		}
	}

	// event handling
	go func() {
		defer close(result)

		// session state
		state := &sessionState{}
		state.reset()

		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					// adding collected events to the chat (assistant's tokens and tool calls)
					if state.builder.Len() != 0 {
						c.AppendEvent(NewEventNewToken(state.builder.String()))
					}
					for _, call := range state.toolCalls {
						c.AppendEvent(call)
					}

					callAmount := len(state.toolCalls)

					// send last tool call if it wasn't sent yet
					if !state.flushLastToolCall(send) {
						return
					}

					// ending current completion
					result <- NewEventCompletionEnded()

					if callAmount == 0 {
						return
					}

					// initializing approval waiter
					verdicts := state.approval.Wait(ctx, callAmount)

					// processing user verdicts
					for verdict := range verdicts {
						call := verdict.call

						if verdict.Accepted {
							callResult, success := c.Tools.Execute(call.Name, call.Content)
							c.AppendEvent(NewEventNewToolMessage(call.CallID, callResult, success))
						} else {
							c.AppendEvent(NewEventNewToolMessage(call.CallID, "User declined the tool call", false))
						}
					}

					// reset state
					state.reset()

					// resume text completion
					client = client.SyncInput(c)
					events = c.Complete(ctx, client)
					continue
				}

				// flush last tool call if event type switched away from tool call stream
				if _, isToolCall := ev.(EventNewToolCall); !isToolCall {
					if !state.flushLastToolCall(send) {
						return
					}
				}

				// in case if we need to skip event
				var skipEvent bool

				// collecting events
				switch event := ev.(type) {
				case EventNewToken:
					state.builder.WriteString(event.Content)
				case EventNewToolCall:
					// prevent adding tool call immediately — we need to wait until end of completion
					skipEvent = true
					if event.CallID != "" {
						// flush the previous tool call — it's now complete
						if !state.flushLastToolCall(send) {
							return
						}
						state.approval.Attach(&event)
						state.toolCalls = append(state.toolCalls, event)
						state.lastToolCall = &state.toolCalls[len(state.toolCalls)-1]
					} else {
						// add token to the last tool call
						state.lastToolCall.Content += event.Content
					}
				case EventNewRefusal:
					c.AppendEvent(event)
				}

				// skipping event
				if skipEvent {
					continue
				}

				// sending events to the channel
				if !send(ev) {
					return
				}
			}
		}
	}()

	return result
}
