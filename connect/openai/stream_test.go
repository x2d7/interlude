package openai_connect

import (
	"context"
	"errors"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/x2d7/interlude/chat"
)

// ==================== Mock SSE Stream ====================

// mockSSEStream implements sseStreamer for testing purposes.
type mockSSEStream struct {
	chunks []openai.ChatCompletionChunk
	index  int
	err    error
	closed bool
}

func newMockSSEStream(chunks []openai.ChatCompletionChunk, err error) *mockSSEStream {
	return &mockSSEStream{
		chunks: chunks,
		index:  -1,
		err:    err,
	}
}

func (m *mockSSEStream) Next() bool {
	if m.index >= len(m.chunks)-1 {
		return false
	}
	m.index++
	return true
}

func (m *mockSSEStream) Current() openai.ChatCompletionChunk {
	if m.index < 0 || m.index >= len(m.chunks) {
		return openai.ChatCompletionChunk{}
	}
	return m.chunks[m.index]
}

func (m *mockSSEStream) Err() error {
	return m.err
}

func (m *mockSSEStream) Close() error {
	m.closed = true
	return nil
}

// ==================== Helper constructors ====================

// makeToolCall builds a ToolCall entry for use in a chunk delta.
func makeToolCall(id, name, arguments string) openai.ChatCompletionChunkChoiceDeltaToolCall {
	return openai.ChatCompletionChunkChoiceDeltaToolCall{
		ID: id,
		Function: openai.ChatCompletionChunkChoiceDeltaToolCallFunction{
			Name:      name,
			Arguments: arguments,
		},
	}
}

// makeChunk builds a minimal ChatCompletionChunk with the given delta content.
func makeChunk(content, refusal string, toolCalls []openai.ChatCompletionChunkChoiceDeltaToolCall) openai.ChatCompletionChunk {
	return openai.ChatCompletionChunk{
		Choices: []openai.ChatCompletionChunkChoice{
			{
				Delta: openai.ChatCompletionChunkChoiceDelta{
					Content:   content,
					Refusal:   refusal,
					ToolCalls: toolCalls,
				},
			},
		},
	}
}

// newStream creates an OpenAIStream backed by the provided mock.
func newStream(mock sseStreamer) *OpenAIStream {
	return &OpenAIStream{SSEStream: mock}
}

// ==================== _handleRawChunk Tests ====================

func TestHandleRawChunk_ContentOnly(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("Hello", "", nil)

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	token, ok := events[0].(chat.EventNewToken)
	if !ok {
		t.Fatalf("Expected EventNewToken, got %T", events[0])
	}
	if token.Content != "Hello" {
		t.Errorf("Expected content 'Hello', got '%s'", token.Content)
	}
}

func TestHandleRawChunk_RefusalOnly(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("", "I cannot help", nil)

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	refusal, ok := events[0].(chat.EventNewRefusal)
	if !ok {
		t.Fatalf("Expected EventNewRefusal, got %T", events[0])
	}
	if refusal.Content != "I cannot help" {
		t.Errorf("Expected refusal 'I cannot help', got '%s'", refusal.Content)
	}
}

func TestHandleRawChunk_SingleToolCall(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("", "", []openai.ChatCompletionChunkChoiceDeltaToolCall{
		makeToolCall("call-1", "weather", `{"city":"Moscow"}`),
	})

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	tc, ok := events[0].(chat.EventNewToolCall)
	if !ok {
		t.Fatalf("Expected EventNewToolCall, got %T", events[0])
	}
	if tc.CallID != "call-1" {
		t.Errorf("Expected CallID 'call-1', got '%s'", tc.CallID)
	}
	if tc.Name != "weather" {
		t.Errorf("Expected Name 'weather', got '%s'", tc.Name)
	}
	if tc.Content != `{"city":"Moscow"}` {
		t.Errorf("Expected Content '{\"city\":\"Moscow\"}', got '%s'", tc.Content)
	}
}

func TestHandleRawChunk_MultipleToolCalls(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("", "", []openai.ChatCompletionChunkChoiceDeltaToolCall{
		makeToolCall("call-1", "tool_a", `{}`),
		makeToolCall("call-2", "tool_b", `{"key":"val"}`),
	})

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	for i, ev := range events {
		if _, ok := ev.(chat.EventNewToolCall); !ok {
			t.Errorf("Expected EventNewToolCall at index %d, got %T", i, ev)
		}
	}
}

func TestHandleRawChunk_AllTypesSimultaneously(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("Hello", "refused", []openai.ChatCompletionChunkChoiceDeltaToolCall{
		makeToolCall("call-1", "my_tool", `{}`),
	})

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// content + refusal + 1 tool call = 3 events
	if len(events) != 3 {
		t.Fatalf("Expected 3 events (content+refusal+toolcall), got %d", len(events))
	}
}

func TestHandleRawChunk_EmptyDelta_ReturnsEmptyList(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("", "", nil)

	events, err := s._handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("_handleRawChunk should not return error for empty delta, got %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected empty events list, got %d", len(events))
	}
}

// ==================== handleRawChunk (decorator) Tests ====================

func TestHandleRawChunkDecorator_NonEmptyEvents_PassesThrough(t *testing.T) {
	s := newStream(newMockSSEStream(nil, nil))
	chunk := makeChunk("token", "", nil)

	events, err := s.handleRawChunk(chunk)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
}

// ==================== OpenAIStream.Next() Tests ====================

func TestNext_SingleChunk_ReturnsTrueAndParsesEvent(t *testing.T) {
	chunk := makeChunk("Hello", "", nil)
	mock := newMockSSEStream([]openai.ChatCompletionChunk{chunk}, nil)
	s := newStream(mock)

	if !s.Next(context.Background()) {
		t.Fatal("Expected Next() = true for first chunk")
	}
	event := s.Current()
	if event == nil {
		t.Fatal("Expected non-nil Current() after Next()")
	}
	if _, ok := event.(chat.EventNewToken); !ok {
		t.Errorf("Expected EventNewToken, got %T", event)
	}
}

func TestNext_MultipleEventsInOneChunk_DrainsQueue(t *testing.T) {
	// Chunk with content + refusal → 2 events from one SSE chunk
	chunk := makeChunk("Hello", "refused", nil)
	mock := newMockSSEStream([]openai.ChatCompletionChunk{chunk}, nil)
	s := newStream(mock)

	// First Next() — fetches chunk, puts 2 events in queue, returns first
	if !s.Next(context.Background()) {
		t.Fatal("Expected Next() = true (1st)")
	}
	first := s.Current()
	if first == nil {
		t.Fatal("Expected non-nil Current() for first event")
	}

	// Second Next() — takes from queue without fetching new chunk
	if !s.Next(context.Background()) {
		t.Fatal("Expected Next() = true (2nd, from queue)")
	}
	second := s.Current()
	if second == nil {
		t.Fatal("Expected non-nil Current() for second event")
	}

	// Third Next() — queue empty, SSE exhausted
	if s.Next(context.Background()) {
		t.Fatal("Expected Next() = false after all events consumed")
	}
}

func TestNext_SSEExhausted_ReturnsFalse(t *testing.T) {
	mock := newMockSSEStream([]openai.ChatCompletionChunk{}, nil)
	s := newStream(mock)

	if s.Next(context.Background()) {
		t.Fatal("Expected Next() = false for empty SSE stream")
	}
}

func TestNext_SSEExhausted_SetsNilErr(t *testing.T) {
	// When SSE ends cleanly (Err() == nil), s.err should be nil
	mock := newMockSSEStream([]openai.ChatCompletionChunk{}, nil)
	s := newStream(mock)

	s.Next(context.Background())

	if s.Err() != nil {
		t.Errorf("Expected nil error after clean SSE end, got %v", s.Err())
	}
}

func TestNext_SSEError_StopsAndSetsErr(t *testing.T) {
	apiErr := errors.New("API error")
	mock := newMockSSEStream([]openai.ChatCompletionChunk{}, apiErr)
	s := newStream(mock)

	if s.Next(context.Background()) {
		t.Fatal("Expected Next() = false when SSE has error")
	}
	if !errors.Is(s.Err(), apiErr) {
		t.Errorf("Expected API error, got %v", s.Err())
	}
}

func TestNext_ErrAlreadySet_ReturnsFalseImmediately(t *testing.T) {
	chunk := makeChunk("Hello", "", nil)
	mock := newMockSSEStream([]openai.ChatCompletionChunk{chunk}, nil)
	s := newStream(mock)

	// Manually set error before calling Next()
	s.err = errors.New("pre-existing error")

	if s.Next(context.Background()) {
		t.Fatal("Expected Next() = false when err is already set")
	}
	// SSE should not have been called (mock index remains -1)
	if mock.index != -1 {
		t.Error("Expected SSE.Next() not to be called when err is already set")
	}
}

func TestNext_MultipleChunks_AllEventsReceived(t *testing.T) {
	chunks := []openai.ChatCompletionChunk{
		makeChunk("Hello", "", nil),
		makeChunk(" world", "", nil),
		makeChunk("!", "", nil),
	}
	mock := newMockSSEStream(chunks, nil)
	s := newStream(mock)

	var received []chat.StreamEvent
	for s.Next(context.Background()) {
		received = append(received, s.Current())
	}

	if len(received) != 3 {
		t.Fatalf("Expected 3 events from 3 chunks, got %d", len(received))
	}

	contents := []string{"Hello", " world", "!"}
	for i, ev := range received {
		token, ok := ev.(chat.EventNewToken)
		if !ok {
			t.Fatalf("Expected EventNewToken at index %d, got %T", i, ev)
		}
		if token.Content != contents[i] {
			t.Errorf("Expected content '%s' at index %d, got '%s'", contents[i], i, token.Content)
		}
	}
}

// ==================== OpenAIStream.Current() Tests ====================

func TestCurrent_BeforeNext_ReturnsNil(t *testing.T) {
	mock := newMockSSEStream(nil, nil)
	s := newStream(mock)

	if s.Current() != nil {
		t.Error("Expected Current() = nil before any Next() call")
	}
}

func TestCurrent_AfterNext_ReturnsEvent(t *testing.T) {
	chunk := makeChunk("token", "", nil)
	mock := newMockSSEStream([]openai.ChatCompletionChunk{chunk}, nil)
	s := newStream(mock)

	s.Next(context.Background())
	if s.Current() == nil {
		t.Error("Expected non-nil Current() after Next()")
	}
}

// ==================== OpenAIStream.Err() Tests ====================

func TestErr_InitiallyNil(t *testing.T) {
	mock := newMockSSEStream(nil, nil)
	s := newStream(mock)

	if s.Err() != nil {
		t.Errorf("Expected nil error initially, got %v", s.Err())
	}
}

func TestErr_AfterSSEError(t *testing.T) {
	expected := errors.New("stream error")
	mock := newMockSSEStream(nil, expected)
	s := newStream(mock)

	s.Next(context.Background()) // triggers SSE read → SSE.Next() returns false → s.err = SSE.Err()

	if !errors.Is(s.Err(), expected) {
		t.Errorf("Expected stream error, got %v", s.Err())
	}
}

// ==================== OpenAIStream.Close() Tests ====================

func TestClose_DelegatesToSSEStream(t *testing.T) {
	mock := newMockSSEStream(nil, nil)
	s := newStream(mock)

	err := s.Close()
	if err != nil {
		t.Errorf("Expected nil error from Close(), got %v", err)
	}
	if !mock.closed {
		t.Error("Expected SSEStream.Close() to be called")
	}
}
