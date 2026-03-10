package chat

import (
	"encoding/json"
	"fmt"
)

// envelope is used for serialization of types that require a type discriminator.
type envelope[T ~string] struct {
	Type    T               `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// MarshalEvent serializes a StreamEvent to JSON.
// The result contains a type discriminator and the event payload.
func MarshalEvent(e StreamEvent) ([]byte, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return json.Marshal(envelope[eventType]{
		Type:    e.getType(),
		Payload: payload,
	})
}

// UnmarshalEvent deserializes a StreamEvent from JSON.
func UnmarshalEvent(data []byte) (StreamEvent, error) {
	var env envelope[eventType]
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}

	switch env.Type {
	case eventToken:
		return unmarshalPayload[EventToken](env.Payload)
	case eventToolCall:
		return unmarshalPayload[EventToolCall](env.Payload)
	case eventToolCallToken:
		return unmarshalPayload[EventToolCallToken](env.Payload)
	case eventRefusal:
		return unmarshalPayload[EventRefusal](env.Payload)
	case eventCompletionStart:
		return unmarshalPayload[EventCompletionStart](env.Payload)
	case eventCompletionEnded:
		return unmarshalPayload[EventCompletionEnded](env.Payload)
	case eventUserMessage:
		return unmarshalPayload[EventUserMessage](env.Payload)
	case eventAssistantMessage:
		return unmarshalPayload[EventAssistantMessage](env.Payload)
	case eventSystemMessage:
		return unmarshalPayload[EventSystemMessage](env.Payload)
	case eventToolMessage:
		return unmarshalPayload[EventToolMessage](env.Payload)
	case eventError:
		return unmarshalPayload[EventError](env.Payload)
	default:
		return nil, fmt.Errorf("unknown event type: %s", env.Type)
	}
}

func unmarshalPayload[T any](data json.RawMessage) (T, error) {
	var v T
	err := json.Unmarshal(data, &v)
	return v, err
}
