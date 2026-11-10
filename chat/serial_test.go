package chat

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalUnmarshalEvent_RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		event StreamEvent
		check func(t *testing.T, result StreamEvent)
	}{
		{
			name:  "EventToken",
			event: NewEventToken("hello"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToken)
				require.True(t, ok)
				assert.Equal(t, "hello", e.Content)
			},
		},
		{
			name:  "EventThinking",
			event: NewEventThinking("let me think..."),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventThinking)
				require.True(t, ok)
				assert.Equal(t, "let me think...", e.Content)
			},
		},
		{
			name:  "EventReasoningMessage",
			event: NewEventReasoningMessage("reasoning content"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventReasoningMessage)
				require.True(t, ok)
				assert.Equal(t, "reasoning content", e.Content)
			},
		},
		{
			name:  "EventToolCall",
			event: NewEventToolCall("call-123", "my_tool", `{"key":"value"}`),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolCall)
				require.True(t, ok)
				assert.Equal(t, "call-123", e.CallID)
				assert.Equal(t, "my_tool", e.Name)
				assert.Equal(t, `{"key":"value"}`, e.Content)
			},
		},
		{
			name:  "EventRefusal",
			event: NewEventRefusal("I cannot do that"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventRefusal)
				require.True(t, ok)
				assert.Equal(t, "I cannot do that", e.Content)
			},
		},
		{
			name:  "EventCompletionEnded",
			event: NewEventCompletionEnded(nil),
			check: func(t *testing.T, result StreamEvent) {
				_, ok := result.(EventCompletionEnded)
				require.True(t, ok)
			},
		},
		{
			name:  "EventUserMessage",
			event: NewEventUserMessage("hello from user"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventUserMessage)
				require.True(t, ok)
				assert.Equal(t, "hello from user", e.Content)
			},
		},
		{
			name:  "EventAssistantMessage",
			event: NewEventAssistantMessage("hello from assistant"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventAssistantMessage)
				require.True(t, ok)
				assert.Equal(t, "hello from assistant", e.Content)
			},
		},
		{
			name:  "EventSystemMessage",
			event: NewEventSystemMessage("system prompt"),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventSystemMessage)
				require.True(t, ok)
				assert.Equal(t, "system prompt", e.Content)
			},
		},
		{
			name:  "EventToolMessage",
			event: NewEventToolMessage("call-123", "tool result", true),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolMessage)
				require.True(t, ok)
				assert.Equal(t, "call-123", e.CallID)
				assert.Equal(t, "tool result", e.Content)
				assert.True(t, e.Success)
			},
		},
		{
			name:  "EventToolMessage_Failed",
			event: NewEventToolMessage("call-456", "tool failed", false),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolMessage)
				require.True(t, ok)
				assert.False(t, e.Success)
			},
		},
		{
			name:  "EventError",
			event: NewEventError(errors.New("something went wrong")),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventError)
				require.True(t, ok)
				assert.Equal(t, "something went wrong", e.Error.Error())
			},
		},
		{
			name:  "EventCompletionStart",
			event: NewEventCompletionStart(),
			check: func(t *testing.T, result StreamEvent) {
				_, ok := result.(EventCompletionStart)
				require.True(t, ok)
			},
		},
		{
			name:  "EventToolCallToken",
			event: NewEventToolCallToken("call-789", "stream_tool", `{"par`),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolCallToken)
				require.True(t, ok)
				assert.Equal(t, "call-789", e.CallID)
				assert.Equal(t, "stream_tool", e.Name)
				assert.Equal(t, `{"par`, e.Content)
			},
		},
		{
			name:  "EventToolCallResolved_Accepted",
			event: NewEventToolCallResolved("call-123", true),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolCallResolved)
				require.True(t, ok)
				assert.Equal(t, "call-123", e.CallID)
				assert.True(t, e.Accepted)
			},
		},
		{
			name:  "EventToolCallResolved_Rejected",
			event: NewEventToolCallResolved("call-456", false),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventToolCallResolved)
				require.True(t, ok)
				assert.Equal(t, "call-456", e.CallID)
				assert.False(t, e.Accepted)
			},
		},
		{
			name: "EventCompletionEnded_WithMeta",
			event: func() EventCompletionEnded {
				meta := CompletionMeta{
					PromptTokens:     38,
					CompletionTokens: 8,
					TotalTokens:      46,
					CostUSD:          0.00021,
					ListCostUSD:      0.000021,
					BillingSource:    "go",
					FinishReason:     "stop",
					Model:            "gpt-4o",
					Metadata:         map[string]string{"opencode.tokens": `{"input":38}`},
				}
				return NewEventCompletionEndedWithUsage(meta, []EventToolCall{
					NewEventToolCall("call-1", "tool_a", `{"x":1}`),
				})
			}(),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventCompletionEnded)
				require.True(t, ok)
				assert.Equal(t, int64(38), e.PromptTokens)
				assert.Equal(t, int64(8), e.CompletionTokens)
				assert.Equal(t, int64(46), e.TotalTokens)
				assert.Equal(t, 0.00021, e.CostUSD)
				assert.Equal(t, 0.000021, e.ListCostUSD)
				assert.Equal(t, "go", e.BillingSource)
				assert.Equal(t, "stop", e.FinishReason)
				assert.Equal(t, "gpt-4o", e.Model)
				assert.Equal(t, map[string]string{"opencode.tokens": `{"input":38}`}, e.Metadata)
				require.Len(t, e.ToolCalls, 1)
				assert.Equal(t, "call-1", e.ToolCalls[0].CallID)
			},
		},
		{
			name: "EventCompletionEnded_WithToolCalls",
			event: NewEventCompletionEnded([]EventToolCall{
				NewEventToolCall("call-1", "tool_a", `{"x":1}`),
				NewEventToolCall("call-2", "tool_b", `{"y":2}`),
			}),
			check: func(t *testing.T, result StreamEvent) {
				e, ok := result.(EventCompletionEnded)
				require.True(t, ok)
				require.Len(t, e.ToolCalls, 2)
				assert.Equal(t, "call-1", e.ToolCalls[0].CallID)
				assert.Equal(t, "tool_a", e.ToolCalls[0].Name)
				assert.Equal(t, `{"x":1}`, e.ToolCalls[0].Content)
				assert.Equal(t, "call-2", e.ToolCalls[1].CallID)
				assert.Equal(t, "tool_b", e.ToolCalls[1].Name)
				assert.Equal(t, `{"y":2}`, e.ToolCalls[1].Content)
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalEvent(tt.event)
			require.NoError(t, err)
			require.NotEmpty(t, data)

			result, err := UnmarshalEvent(data)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Equal(t, tt.event.getType(), result.getType())
			tt.check(t, result)
		})
	}
}

func TestMarshalEvent_ContainsTypeDiscriminator(t *testing.T) {
	data, err := MarshalEvent(NewEventToken("hello"))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"token"`)
	assert.Contains(t, string(data), `"payload"`)
}

func TestMarshalEvent_CompletionEndedType(t *testing.T) {
	data, err := MarshalEvent(NewEventCompletionEnded(nil))
	require.NoError(t, err)
	assert.Contains(t, string(data), `"type":"completion_ended"`)
}

func TestMarshalEvent_CompletionEndedOmitsEmptyMeta(t *testing.T) {
	data, err := MarshalEvent(NewEventCompletionEnded(nil))
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"prompt_tokens"`)
	assert.NotContains(t, string(data), `"cost_usd"`)
	assert.NotContains(t, string(data), `"billing_source"`)
}

func TestUnmarshalEvent_UnknownType(t *testing.T) {
	_, err := UnmarshalEvent([]byte(`{"type":"unknown","payload":{}}`))
	assert.Error(t, err)
}

func TestUnmarshalEvent_InvalidJSON(t *testing.T) {
	_, err := UnmarshalEvent([]byte(`not json`))
	assert.Error(t, err)
}

func TestUnmarshalEvent_InvalidPayload(t *testing.T) {
	_, err := UnmarshalEvent([]byte(`{"type":"token","payload":"not an object"}`))
	assert.Error(t, err)
}
