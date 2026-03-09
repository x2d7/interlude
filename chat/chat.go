package chat

import (
	"context"
	"strings"

	"github.com/x2d7/interlude/chat/tools"
)

func (c *Chat) Complete(ctx context.Context, client Client) <-chan StreamEvent {
	result := make(chan StreamEvent, 16)

	// sending events to the channel in background
	go func() {
		defer close(result)

		// text completion stream
		stream := client.NewStreaming(ctx)
		if stream == nil {
			result <- NewEventError(ErrNilStreaming)
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
			result <- NewEventError(err)
		}
	}()

	return result
}

type sessionState struct {
	// session context

	client Client
	events <-chan StreamEvent
	send   func(StreamEvent) bool
	ctx    context.Context

	// session state variables

	builder      strings.Builder
	toolCalls    []EventToolCall
	lastToolCall *EventToolCall
	approval     *ApproveWaiter
}

func (s *sessionState) reset() {
	s.builder.Reset()
	s.toolCalls = s.toolCalls[:0]
	s.lastToolCall = nil
	s.approval = NewApproveWaiter(s.ctx)
}

func (s *sessionState) flushLastToolCall() bool {
	if s.lastToolCall == nil {
		return true
	}
	ok := s.send(*s.lastToolCall)
	s.lastToolCall = nil
	return ok
}

func (c *Chat) ensureDefaults() {
	if c.Messages == nil {
		c.Messages = NewMessages()
	}
	if c.Tools == nil {
		t := tools.NewTools()
		c.Tools = t
	}
}

func (c *Chat) Session(ctx context.Context, client Client) <-chan StreamEvent {
	// ensuring default values
	c.ensureDefaults()

	// creating the channels
	result := make(chan StreamEvent, 16)

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
		state := &sessionState{
			send: send,
			ctx:  ctx,
		}

		// flag to start completion this iteration
		restart := true

		for {
			if restart {
				// reset state
				state.reset()

				// insert chat context into client input configuration
				client := client.SyncInput(c)
				state.client = client

				// start completion
				state.events = c.Complete(ctx, client)
				restart = false
			}

			select {
			case <-ctx.Done():
				return
			case ev, ok := <-state.events:
				if !ok {
					if !c.handleCompletionEnd(ctx, state) {
						return
					}
					restart = true
					continue
				}

				// flush last tool call if event type switched away from tool call stream
				if _, isToolCall := ev.(EventToolCall); !isToolCall {
					if !state.flushLastToolCall() {
						return
					}
				}

				// in case if we need to skip event
				var skipEvent bool

				// collecting events
				switch event := ev.(type) {
				case EventToken:
					state.builder.WriteString(event.Content)
				case EventToolCall:
					// prevent adding tool call immediately — we need to wait until end of completion
					skipEvent = true
					if event.CallID != "" {
						// flush the previous tool call — it's now complete
						if !state.flushLastToolCall() {
							return
						}
						state.approval.Attach(&event)
						state.toolCalls = append(state.toolCalls, event)
						state.lastToolCall = &state.toolCalls[len(state.toolCalls)-1]
					} else {
						// add token to the last tool call
						state.lastToolCall.Content += event.Content
					}
				case EventRefusal:
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

func (c *Chat) handleCompletionEnd(ctx context.Context, state *sessionState) (proceed bool) {
	// adding collected events to the chat (assistant's tokens and tool calls)
	if state.builder.Len() != 0 {
		c.AppendEvent(NewEventAssistantMessage(state.builder.String()))
	}
	for _, call := range state.toolCalls {
		c.AppendEvent(call)
	}

	callAmount := len(state.toolCalls)

	// send last tool call if it wasn't sent yet
	if !state.flushLastToolCall() {
		return false
	}

	// ending current completion
	if !state.send(NewEventCompletionEnded()) {
		return false
	}

	if callAmount == 0 {
		return false
	}

	// initializing approval waiter
	verdicts := state.approval.Wait(ctx, callAmount)

	// processing user verdicts
	for verdict := range verdicts {
		call := verdict.call

		if verdict.Accepted {
			callResult, success := c.Tools.Execute(call.Name, call.Content)
			c.AppendEvent(NewEventToolMessage(call.CallID, callResult, success))
		} else {
			c.AppendEvent(NewEventToolMessage(call.CallID, "User declined the tool call", false))
		}
	}

	return true
}
