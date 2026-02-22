package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/x2d7/interlude/chat/tools"
)

// ==================== Mock Implementations ====================

// MockStream implements Stream[StreamEvent] for testing
type MockStream struct {
	events []StreamEvent
	index  int
	err    error
	mu     sync.Mutex
}

func NewMockStream(events []StreamEvent, err error) *MockStream {
	return &MockStream{
		events: events,
		index:  -1,
		err:    err,
	}
}

func (s *MockStream) Next(ctx context.Context) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.index >= len(s.events)-1 {
		return false
	}
	s.index++
	return true
}

func (s *MockStream) Current() StreamEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.index < 0 || s.index >= len(s.events) {
		return nil
	}
	return s.events[s.index]
}

func (s *MockStream) Err() error {
	return s.err
}

func (s *MockStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = len(s.events)
	return nil
}

// MockClient implements Client for testing
type MockClient struct {
	StreamingEvents []StreamEvent
	StreamingError  error
	SyncedChat      *Chat
	CallCount       int
	mu              sync.Mutex
}

func NewMockClient() *MockClient {
	return &MockClient{}
}

func (c *MockClient) NewStreaming(ctx context.Context) Stream[StreamEvent] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CallCount++
	// Return nil if no events are configured (simulates nil stream case)
	if c.StreamingEvents == nil {
		return nil
	}
	return NewMockStream(c.StreamingEvents, c.StreamingError)
}

func (c *MockClient) SyncInput(chat *Chat) Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CallCount++
	newClient := &MockClient{
		StreamingEvents: c.StreamingEvents,
		StreamingError:  c.StreamingError,
		SyncedChat:      chat,
	}
	return newClient
}

// SetStreamingEvents sets the events to be returned by NewStreaming
func (c *MockClient) SetStreamingEvents(events []StreamEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StreamingEvents = events
}

// SetStreamingError sets the error to be returned by stream.Err()
func (c *MockClient) SetStreamingError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.StreamingError = err
}

// MultiRoundMockClient returns different events on each round
type MultiRoundMockClient struct {
	Rounds     [][]StreamEvent
	roundIndex int
	SyncedChat *Chat
	mu         sync.Mutex
}

func NewMultiRoundMockClient(rounds [][]StreamEvent) *MultiRoundMockClient {
	return &MultiRoundMockClient{Rounds: rounds}
}

func (c *MultiRoundMockClient) NewStreaming(ctx context.Context) Stream[StreamEvent] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.roundIndex >= len(c.Rounds) {
		// Return empty stream - Session will handle completion signal itself
		return NewMockStream([]StreamEvent{}, nil)
	}
	events := c.Rounds[c.roundIndex]
	c.roundIndex++
	return NewMockStream(events, nil)
}

func (c *MultiRoundMockClient) SyncInput(chat *Chat) Client {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.SyncedChat = chat
	return c
}

// ==================== Complete Tests ====================

// TestComplete_SuccessTokens tests successful completion with tokens
func TestComplete_SuccessTokens(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Hello"),
		NewEventNewToken(" world"),
	})

	ctx := context.Background()
	events := chat.Complete(ctx, mockClient)

	var received []StreamEvent
	for event := range events {
		received = append(received, event)
	}

	// Should receive 2 token events
	if len(received) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(received))
	}

	// Verify token content
	if received[0].GetType() != eventNewToken {
		t.Errorf("Expected EventNewToken, got %v", received[0].GetType())
	}
	if received[1].GetType() != eventNewToken {
		t.Errorf("Expected EventNewToken, got %v", received[1].GetType())
	}
}

// TestComplete_SuccessToolCall tests completion with tool call
func TestComplete_SuccessToolCall(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Calling tool"),
		NewEventNewToolCall("call-123", "weather", `{"city": "Moscow"}`),
	})

	ctx := context.Background()
	events := chat.Complete(ctx, mockClient)

	var received []StreamEvent
	for event := range events {
		received = append(received, event)
	}

	// Should receive token + tool call
	if len(received) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(received))
	}

	// Second event should be tool call
	toolCall, ok := received[1].(EventNewToolCall)
	if !ok {
		t.Fatalf("Expected EventNewToolCall, got %T", received[1])
	}

	if toolCall.CallID != "call-123" {
		t.Errorf("Expected CallID 'call-123', got '%s'", toolCall.CallID)
	}
	if toolCall.Name != "weather" {
		t.Errorf("Expected Name 'weather', got '%s'", toolCall.Name)
	}
}

// TestComplete_NilStream tests when NewStreaming returns nil
func TestComplete_NilStream(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	// Custom mock that returns nil stream
	mockClient := &MockClient{
		StreamingEvents: nil, // This will cause nil stream in Complete
	}

	ctx := context.Background()
	events := chat.Complete(ctx, mockClient)

	var received []StreamEvent
	for event := range events {
		received = append(received, event)
	}

	// Should receive error event
	if len(received) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(received))
	}

	errEvent, ok := received[0].(EventNewError)
	if !ok {
		t.Fatalf("Expected EventNewError, got %T", received[0])
	}

	if !errors.Is(errEvent.Error, ErrNilStreaming) {
		t.Errorf("Expected ErrNilStreaming, got %v", errEvent.Error)
	}
}

// TestComplete_StreamError tests error during streaming
func TestComplete_StreamError(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Some text"),
	})
	mockClient.SetStreamingError(errors.New("API error"))

	ctx := context.Background()
	events := chat.Complete(ctx, mockClient)

	var received []StreamEvent
	for event := range events {
		received = append(received, event)
	}

	// Should receive token + error
	if len(received) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(received))
	}

	errEvent, ok := received[1].(EventNewError)
	if !ok {
		t.Fatalf("Expected EventNewError, got %T", received[1])
	}

	if errEvent.Error.Error() != "API error" {
		t.Errorf("Expected 'API error', got '%s'", errEvent.Error.Error())
	}
}

// ==================== SyncInput Tests ====================

// TestSyncInput_ReturnsNewClient tests that SyncInput returns a new client instance
func TestSyncInput_ReturnsNewClient(t *testing.T) {
	mockClient := NewMockClient()

	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	newClient := mockClient.SyncInput(chat)

	// Should be a different instance
	if newClient == mockClient {
		t.Error("SyncInput should return a new client instance")
	}
}

// TestSyncInput_WithNewMessages tests that new client contains updated messages
func TestSyncInput_WithNewMessages(t *testing.T) {
	mockClient := NewMockClient()

	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	// Add a message to the chat
	chat.AddMessage(SenderUser{}, "Hello")

	// Get synced client
	newClient := mockClient.SyncInput(chat)

	// Verify the chat was passed through
	syncedClient, ok := newClient.(*MockClient)
	if !ok {
		t.Fatalf("Expected *MockClient, got %T", newClient)
	}

	if syncedClient.SyncedChat == nil {
		t.Error("SyncedChat should not be nil")
	}

	// Check that message was added
	messages := syncedClient.SyncedChat.Messages.Snapshot()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewUserMessage {
		t.Errorf("Expected EventNewUserMessage, got %v", messages[0].GetType())
	}
}

func TestSyncInput_WithNewTools(t *testing.T) {
	mockClient := NewMockClient()

	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}
	// Add a tool using NewTool
	tool, err := tools.NewTool("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "result", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Get synced client
	newClient := mockClient.SyncInput(chat)

	// Verify tools were passed
	syncedClient, ok := newClient.(*MockClient)
	if !ok {
		t.Fatalf("Expected *MockClient, got %T", newClient)
	}

	if syncedClient.SyncedChat == nil {
		t.Error("SyncedChat should not be nil")
	}

	// Check that tool was added
	toolList := syncedClient.SyncedChat.Tools.Snapshot()
	if len(toolList) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(toolList))
	}

	if toolList[0].Id != "test-tool" {
		t.Errorf("Expected tool Id 'test-tool', got '%s'", toolList[0].Id)
	}
}

// ==================== Session Tests - Data Collection ====================

// TestSession_CollectsTokens tests that accumulated tokens are added to history
func TestSession_CollectsTokens(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	// Note: tokens are accumulated into a single message
	// EventCompletionEnded is generated by Session, not the mock
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Hello"),
		NewEventNewToken(" world"),
	})

	ctx := context.Background()
	events := chat.Session(ctx, mockClient)

	// Drain the events channel - this will close when stream ends
	for range events {
		// Just drain
	}

	// Check that tokens were added to history
	messages := chat.Messages.Snapshot()

	// Tokens are accumulated into ONE message (concatenated)
	var tokenCount int
	var tokenContent string
	for _, msg := range messages {
		if msg.GetType() == eventNewAssistantMessage {
			tokenCount++
			tokenContent = msg.(EventNewAssistantMessage).Content
		}
	}

	if tokenCount != 1 {
		t.Errorf("Expected 1 token in history (merged), got %d", tokenCount)
	}

	// Verify the content is merged
	if tokenContent != "Hello world" {
		t.Errorf("Expected token content 'Hello world', got '%s'", tokenContent)
	}
}

// TestSession_CollectsToolCalls
func TestSession_CollectsToolCalls(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	// Round 1: one tool call (Session generates CompletionEnded internally)
	// Round 2: empty (session exits cleanly, no extra tool call)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "tool1", `{}`)},
		{},
	})

	ctx := context.Background()
	events := chat.Session(ctx, mockClient)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range events {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()
	<-done

	messages := chat.Messages.Snapshot()
	var toolCallCount int
	for _, msg := range messages {
		if msg.GetType() == eventNewToolCall {
			toolCallCount++
		}
	}

	if toolCallCount != 1 {
		t.Errorf("Expected 1 tool call in history, got %d", toolCallCount)
	}
}

// TestSession_EmitsCompletionEnded tests that EventCompletionEnded is emitted
func TestSession_EmitsCompletionEnded(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	// Note: EventCompletionEnded is generated by Session, not the mock
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("test"),
	})

	ctx, cancel := context.WithCancel(context.Background())
	events := chat.Session(ctx, mockClient)

	var receivedCompletion bool
	for event := range events {
		if event.GetType() == eventCompletionEnded {
			receivedCompletion = true
			break
		}
	}

	cancel()

	if !receivedCompletion {
		t.Error("Expected EventCompletionEnded to be received")
	}
}

// TestSession_Refusal tests that refusal is added to history
func TestSession_Refusal(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	// Note: EventCompletionEnded is generated by Session, not the mock
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewRefusal("I cannot help with that"),
	})

	ctx, cancel := context.WithCancel(context.Background())
	events := chat.Session(ctx, mockClient)

	// Drain
	for range events {
		// Just drain
	}

	cancel()

	// Check refusal in history
	messages := chat.Messages.Snapshot()

	var refusalCount int
	for _, msg := range messages {
		if msg.GetType() == eventNewRefusal {
			refusalCount++
		}
	}

	if refusalCount != 1 {
		t.Errorf("Expected 1 refusal in history, got %d", refusalCount)
	}
}

// ==================== Session Tests - Tool Execution ====================

// TestSession_ToolAccepted tests tool execution when accepted
func TestSession_ToolAccepted(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	// Add a test tool
	tool, err := tools.NewTool[map[string]string]("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "success: " + input["key"], nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Round 1: tool call, Round 2: empty to finish (Session generates CompletionEnded)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`)},
		{},
	})

	ctx := context.Background()
	events := chat.Session(ctx, mockClient)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range events {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()

	<-done

	// Check that tool result was added to history
	messages := chat.Messages.Snapshot()

	var toolResultFound bool
	for _, msg := range messages {
		if msg.GetType() == eventNewToolMessage {
			toolMsg := msg.(EventNewToolMessage)
			if toolMsg.CallID == "call-1" && toolMsg.Success {
				toolResultFound = true
			}
		}
	}

	if !toolResultFound {
		t.Error("Expected successful tool result in history")
	}
}

// TestSession_ToolRejected tests tool execution when rejected
func TestSession_ToolRejected(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "should not be called", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Round 1: tool call, Round 2: empty to finish (Session generates CompletionEnded)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`)},
		{},
	})

	ctx := context.Background()
	events := chat.Session(ctx, mockClient)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range events {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(false)
			}
		}
	}()

	<-done

	// Check that rejection message was added to history
	messages := chat.Messages.Snapshot()

	var rejectionFound bool
	for _, msg := range messages {
		if msg.GetType() == eventNewToolMessage {
			toolMsg := msg.(EventNewToolMessage)
			if toolMsg.CallID == "call-1" && !toolMsg.Success &&
				toolMsg.Content == "User declined the tool call" {
				rejectionFound = true
			}
		}
	}

	if !rejectionFound {
		t.Error("Expected rejection message in history")
	}
}

// TestSession_NonExistentTool tests calling a tool that doesn't exist
func TestSession_NonExistentTool(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	// Round 1: tool call, Round 2: empty to finish (Session generates CompletionEnded)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "non-existent", `{}`)},
		{},
	})

	ctx := context.Background()
	events := chat.Session(ctx, mockClient)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for event := range events {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()

	<-done

	// Check that error message was added to history
	messages := chat.Messages.Snapshot()

	var errorFound bool
	for _, msg := range messages {
		if msg.GetType() == eventNewToolMessage {
			toolMsg := msg.(EventNewToolMessage)
			if toolMsg.CallID == "call-1" && !toolMsg.Success {
				errorFound = true
			}
		}
	}

	if !errorFound {
		t.Error("Expected error message for non-existent tool in history")
	}
}

// ==================== Helpers Tests ====================

// TestHelpers_AddMessage tests all AddMessage variants
func TestHelpers_AddMessage(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	// Test User message
	err := chat.AddMessage(SenderUser{}, "Hello")
	if err != nil {
		t.Errorf("AddMessage(SenderUser) failed: %v", err)
	}

	// Test Assistant message
	err = chat.AddMessage(SenderAssistant{}, "Response")
	if err != nil {
		t.Errorf("AddMessage(SenderAssistant) failed: %v", err)
	}

	// Test System message
	err = chat.AddMessage(SenderSystem{}, "System prompt")
	if err != nil {
		t.Errorf("AddMessage(SenderSystem) failed: %v", err)
	}

	// Verify all messages were added
	messages := chat.Messages.Snapshot()
	if len(messages) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewUserMessage {
		t.Errorf("Expected EventNewUserMessage, got %v", messages[0].GetType())
	}
	if messages[1].GetType() != eventNewAssistantMessage {
		t.Errorf("Expected EventNewAssistantMessage, got %v", messages[1].GetType())
	}
	if messages[2].GetType() != eventNewSystemMessage {
		t.Errorf("Expected EventNewSystemMessage, got %v", messages[2].GetType())
	}
}

// TestHelpers_AppendEvent tests AppendEvent helper
func TestHelpers_AppendEvent(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	event := NewEventNewToken("test")
	chat.AppendEvent(event)

	messages := chat.Messages.Snapshot()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewToken {
		t.Errorf("Expected EventNewToken, got %v", messages[0].GetType())
	}
}

// TestHelpers_SendUserStream tests SendUserStream helper
func TestHelpers_SendUserStream(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventCompletionEnded(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	events := chat.SendUserStream(ctx, mockClient, "User message")

	// Verify message was added
	messages := chat.Messages.Snapshot()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewUserMessage {
		t.Errorf("Expected EventNewUserMessage, got %v", messages[0].GetType())
	}

	// Drain events
	for range events {
	}

	cancel()
}

// TestHelpers_SendAssistantStream tests SendAssistantStream helper
func TestHelpers_SendAssistantStream(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventCompletionEnded(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	events := chat.SendAssistantStream(ctx, mockClient, "Assistant message")

	// Verify message was added
	messages := chat.Messages.Snapshot()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewAssistantMessage {
		t.Errorf("Expected EventNewAssistantMessage, got %v", messages[0].GetType())
	}

	// Drain events
	for range events {
	}

	cancel()
}

// TestHelpers_SendSystemStream tests SendSystemStream helper
func TestHelpers_SendSystemStream(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventCompletionEnded(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	events := chat.SendSystemStream(ctx, mockClient, "System message")

	// Verify message was added
	messages := chat.Messages.Snapshot()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	if messages[0].GetType() != eventNewSystemMessage {
		t.Errorf("Expected EventNewSystemMessage, got %v", messages[0].GetType())
	}

	// Drain events
	for range events {
	}

	cancel()
}

// ==================== Context Cancellation Tests ====================

// TestComplete_ContextCancelledDuringStream tests context cancellation while streaming tokens
func TestComplete_ContextCancelledDuringStream(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	// Slow stream that will be cancelled
	events := []StreamEvent{
		NewEventNewToken("Hello"),
		NewEventNewToken(" world"),
		// No completion - stream will be stuck waiting
	}

	mockClient := NewMockClient()
	mockClient.SetStreamingEvents(events)

	ctx, cancel := context.WithCancel(context.Background())
	result := chat.Complete(ctx, mockClient)

	// Receive first event
	event1 := <-result
	if event1.GetType() != eventNewToken {
		t.Errorf("Expected first event to be token, got %v", event1.GetType())
	}

	// Cancel context
	cancel()

	// Give some time for the goroutine to process cancellation
	time.Sleep(50 * time.Millisecond)

	// Drain remaining events - after cancellation there should be no more
	// because the channel is closed after all events are processed
	for event := range result {
		t.Logf("Got event after cancellation: %v", event.GetType())
	}
}

// TestSession_ContextCancelledDuringTokenCollection tests context cancellation during token collection
func TestSession_ContextCancelledDuringTokenCollection(t *testing.T) {
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    &tools.Tools{},
	}

	mockClient := NewMockClient()
	// Stream that won't send completion - simulates stuck stream
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Hello"),
		// Missing EventCompletionEnded - stream hangs
	})

	ctx, cancel := context.WithCancel(context.Background())
	result := chat.Session(ctx, mockClient)

	// Receive first token event
	var receivedToken bool
	for event := range result {
		if event.GetType() == eventNewToken {
			receivedToken = true
			break
		}
	}

	if !receivedToken {
		t.Fatal("Expected to receive at least one token")
	}

	// Cancel context - should stop the session
	cancel()

	// Key thing is session stopped - we verified we received at least one token above
}

// TestSession_ContextCancelledWhileWaitingForApproval tests context cancellation while waiting for tool approval
func TestSession_ContextCancelledWhileWaitingForApproval(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "result", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Round 1: tool call that requires approval (Session generates CompletionEnded)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`)},
		// Round 2 never comes - context cancelled
	})

	ctx, cancel := context.WithCancel(context.Background())
	result := chat.Session(ctx, mockClient)

	// Receive tool call event
	var toolCallEvent EventNewToolCall
	for event := range result {
		if tc, ok := event.(EventNewToolCall); ok {
			toolCallEvent = tc
			break
		}
	}

	if toolCallEvent.CallID != "call-1" {
		t.Fatalf("Expected to receive tool call event")
	}

	// Cancel context BEFORE resolving - simulates user cancelling while waiting for approval
	cancel()

	// Tool should NOT have been executed (context was cancelled)
	messages := chat.Messages.Snapshot()
	toolExecuted := false
	for _, msg := range messages {
		if msg.GetType() == eventNewToolMessage {
			toolExecuted = true
			break
		}
	}

	if toolExecuted {
		t.Error("Tool should not have been executed when context was cancelled")
	}
}

// TestSession_ContextCancelledBetweenRounds tests context cancellation between completion rounds
func TestSession_ContextCancelledBetweenRounds(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "result", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// First round: tool call (Session generates CompletionEnded)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`)},
	})

	ctx, cancel := context.WithCancel(context.Background())
	result := chat.Session(ctx, mockClient)

	// Handle tool call - resolve it
	go func() {
		for event := range result {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()

	// Let first round complete
	time.Sleep(100 * time.Millisecond)

	// Cancel context between rounds
	cancel()

	// Verify only first tool was executed, not second round
	messages := chat.Messages.Snapshot()
	toolMessageCount := 0
	for _, msg := range messages {
		if msg.GetType() == eventNewToolMessage {
			toolMessageCount++
		}
	}

	if toolMessageCount > 1 {
		t.Errorf("Expected at most 1 tool message, got %d", toolMessageCount)
	}
}

// ==================== Tool Call Assembly Tests ====================

// TestSession_ToolCallAssembly_Basic tests basic tool call assembly from multiple chunks
func TestSession_ToolCallAssembly_Basic(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("weather", "Get weather",
		func(input map[string]string) (string, error) {
			return "sunny", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Simulate streaming tool call in chunks:
	// 1. First chunk with CallID (start of tool call)
	// 2. Subsequent chunks without CallID (continuation)
	// Session generates CompletionEnded internally, so we don't add it to mocks
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{
			NewEventNewToolCall("call-1", "weather", `{"city": "`),
			NewEventNewToolCall("", "weather", "Moscow"),
			NewEventNewToolCall("", "weather", `"}`),
		},
		// Second round: empty (session exits cleanly)
		{},
	})

	ctx := context.Background()
	result := chat.Session(ctx, mockClient)

	// Resolve tool call
	go func() {
		for event := range result {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()

	// Wait for completion
	time.Sleep(200 * time.Millisecond)

	// Check that tool call was assembled correctly in history
	messages := chat.Messages.Snapshot()
	var toolCall EventNewToolCall
	for _, msg := range messages {
		if tc, ok := msg.(EventNewToolCall); ok {
			toolCall = tc
			break
		}
	}

	if toolCall.CallID != "call-1" {
		t.Errorf("Expected CallID 'call-1', got '%s'", toolCall.CallID)
	}

	// The content should contain assembled JSON
	if !strings.Contains(toolCall.Content, "Moscow") {
		t.Errorf("Expected assembled content to contain 'Moscow', got '%s'", toolCall.Content)
	}
}

// TestSession_ToolCallAssembly_MultipleToolsWithAssembly tests multiple tool calls where some are assembled
func TestSession_ToolCallAssembly_MultipleToolsWithAssembly(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool1, err := tools.NewTool[map[string]string]("tool1", "Tool 1",
		func(input map[string]string) (string, error) {
			return "result1", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	tool2, err := tools.NewTool[map[string]string]("tool2", "Tool 2",
		func(input map[string]string) (string, error) {
			return "result2", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool1)
	chat.Tools.Add(tool2)

	// 3 tool calls:
	// 1) assembled from chunks (call-1)
	// 2) assembled from chunks (call-2)
	// 3) complete (call-3)
	// Session generates CompletionEnded internally, so we don't add it to mocks
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{
			NewEventNewToolCall("call-1", "tool1", `{"param": `),
			NewEventNewToolCall("", "tool1", "123"),
			NewEventNewToolCall("", "tool1", "}"),
			NewEventNewToolCall("call-2", "tool2", `{"val": `),
			NewEventNewToolCall("", "tool2", "456"),
			NewEventNewToolCall("", "tool2", "}"),
			NewEventNewToolCall("call-3", "tool2", `{"value": 789}`),
		},
		// Second round: empty (session exits cleanly)
		{},
	})

	ctx := context.Background()
	result := chat.Session(ctx, mockClient)

	// Resolve tool calls - use channel to avoid race
	resolvedCh := make(chan int, 1)
	go func() {
		resolved := 0
		for event := range result {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
				resolved++
			}
		}
		resolvedCh <- resolved
	}()

	// Wait for completion
	resolved := <-resolvedCh
	close(resolvedCh)

	if resolved != 3 {
		t.Errorf("Expected 3 tool calls resolved, got %d", resolved)
	}

	// Verify all tool calls were assembled correctly in history
	messages := chat.Messages.Snapshot()
	toolCalls := make([]EventNewToolCall, 0)
	for _, msg := range messages {
		if tc, ok := msg.(EventNewToolCall); ok {
			toolCalls = append(toolCalls, tc)
		}
	}

	if len(toolCalls) != 3 {
		t.Fatalf("Expected 3 tool calls, got %d", len(toolCalls))
	}

	// First tool (assembled from chunks)
	if !strings.Contains(toolCalls[0].Content, "123") {
		t.Errorf("First tool should contain assembled '123', got '%s'", toolCalls[0].Content)
	}
	// Second tool (assembled from chunks)
	if !strings.Contains(toolCalls[1].Content, "456") {
		t.Errorf("Second tool should contain assembled '456', got '%s'", toolCalls[1].Content)
	}
	// Third tool (complete)
	if !strings.Contains(toolCalls[2].Content, "789") {
		t.Errorf("Third tool should contain '789', got '%s'", toolCalls[2].Content)
	}
}

// TestSession_ToolCallAssembly_LargeContent tests assembly of larger content
func TestSession_ToolCallAssembly_LargeContent(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("search", "Search",
		func(input map[string]string) (string, error) {
			return "results", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Build a large JSON by streaming many small chunks
	// Using array syntax: ["chunk0", "chunk1", ...]
	// Session generates CompletionEnded internally, so we don't add it to mocks
	events := []StreamEvent{
		NewEventNewToolCall("call-1", "search", `["chunk0"`),
	}
	// Add more chunks without CallID
	for i := 1; i < 20; i++ {
		chunk := fmt.Sprintf(",\"chunk%d\"", i)
		events = append(events, NewEventNewToolCall("", "search", chunk))
	}
	events = append(events, NewEventNewToolCall("", "search", `]`))

	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		events,
		// Second round: empty (session exits cleanly)
		{},
	})

	ctx := context.Background()
	result := chat.Session(ctx, mockClient)

	// Resolve tool call
	go func() {
		for event := range result {
			if tc, ok := event.(EventNewToolCall); ok {
				tc.Resolve(true)
			}
		}
	}()

	// Wait for completion
	time.Sleep(300 * time.Millisecond)

	// Verify tool call was assembled in history
	messages := chat.Messages.Snapshot()
	var toolCall EventNewToolCall
	for _, msg := range messages {
		if tc, ok := msg.(EventNewToolCall); ok {
			toolCall = tc
			break
		}
	}

	// Should contain assembled content
	if !strings.Contains(toolCall.Content, "chunk0") || !strings.Contains(toolCall.Content, "chunk19") {
		t.Errorf("Expected assembled content with chunks, got '%s'", toolCall.Content)
	}
}

// TestSession_MixedTokensAndToolCalls verifies the interleaved event ordering
// when text tokens and tool calls alternate in a single completion round.
//
// Stream sequence: token("A") → toolCall("call-1") → token("B") → toolCall("call-2") → token("C")
//
// Expected consumer order: A → call-1 → B → call-2 → C → CompletionEnded
//
// This confirms that tool calls are emitted at the point they occur
// (when the next non-toolcall event arrives), NOT batched at the end of the round.
func TestSession_MixedTokensAndToolCalls(t *testing.T) {
	chatTools := tools.NewTools()
	chat := &Chat{
		Messages: NewMessages(),
		Tools:    chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("lookup", "Lookup tool",
		func(input map[string]string) (string, error) {
			return "ok", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// No NewEventCompletionEnded() here — Session generates it from channel close.
	// Round 2 is empty — causes Session to emit final CompletionEnded and exit.
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{
			NewEventNewToken("A"),
			NewEventNewToolCall("call-1", "lookup", `{"q":"1"}`),
			NewEventNewToken("B"),
			NewEventNewToolCall("call-2", "lookup", `{"q":"2"}`),
			NewEventNewToken("C"),
		},
		{}, // empty round — triggers final CompletionEnded + Session exit
	})

	ctx := context.Background()
	result := chat.Session(ctx, mockClient)

	type record struct {
		kind    string // "token" | "tool" | "end"
		content string // token text or tool CallID
	}

	var order []record

	for event := range result {
		switch e := event.(type) {
		case EventNewToken:
			order = append(order, record{"token", e.Content})
		case EventNewToolCall:
			order = append(order, record{"tool", e.CallID})
			e.Resolve(true) // accept all tool calls
		case EventCompletionEnded:
			order = append(order, record{"end", ""})
		}
	}

	// --- helpers ---
	posOf := func(kind, content string) int {
		for i, r := range order {
			if r.kind == kind && r.content == content {
				return i
			}
		}
		return -1
	}

	posA := posOf("token", "A")
	posCall1 := posOf("tool", "call-1")
	posB := posOf("token", "B")
	posCall2 := posOf("tool", "call-2")
	posC := posOf("token", "C")
	posEnd := posOf("end", "")

	t.Logf("Event order: %v", order)

	// All events must be present
	for name, pos := range map[string]int{
		"token:A": posA,
		"call-1":  posCall1,
		"token:B": posB,
		"call-2":  posCall2,
		"token:C": posC,
		"end":     posEnd,
	} {
		if pos == -1 {
			t.Errorf("Missing event: %s", name)
		}
	}

	if t.Failed() {
		return
	}

	// Verify interleaved ordering:
	// A comes before call-1 (token before its following tool)
	if posA >= posCall1 {
		t.Errorf("Expected token:A (%d) before call-1 (%d)", posA, posCall1)
	}
	// call-1 is flushed BEFORE token:B (tool before next token)
	if posCall1 >= posB {
		t.Errorf("Expected call-1 (%d) before token:B (%d)", posCall1, posB)
	}
	// B comes before call-2
	if posB >= posCall2 {
		t.Errorf("Expected token:B (%d) before call-2 (%d)", posB, posCall2)
	}
	// call-2 is flushed BEFORE token:C
	if posCall2 >= posC {
		t.Errorf("Expected call-2 (%d) before token:C (%d)", posCall2, posC)
	}
	// CompletionEnded is last
	if posC >= posEnd {
		t.Errorf("Expected token:C (%d) before CompletionEnded (%d)", posC, posEnd)
	}

	// Verify tool calls also appear in history with correct data
	messages := chat.Messages.Snapshot()
	toolCallsInHistory := 0
	for _, msg := range messages {
		if msg.GetType() == eventNewToolCall {
			toolCallsInHistory++
		}
	}
	if toolCallsInHistory != 2 {
		t.Errorf("Expected 2 tool calls in history, got %d", toolCallsInHistory)
	}
}
