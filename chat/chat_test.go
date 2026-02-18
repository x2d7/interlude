package chat

import (
	"context"
	"errors"
	"sync"
	"testing"

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

func (s *MockStream) Next() bool {
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
		return NewMockStream([]StreamEvent{NewEventCompletionEnded()}, nil)
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
		Tools:    &chatTools,
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
	// Need to send EventCompletionEnded to trigger collection of tokens
	// Note: tokens are accumulated into a single message
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("Hello"),
		NewEventNewToken(" world"),
		NewEventCompletionEnded(),
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
		if msg.GetType() == eventNewToken {
			tokenCount++
			tokenContent = msg.(EventNewToken).Content
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

	// Round 1: one tool call + completion signal
	// Round 2: just completion (session exits cleanly, no extra tool call)
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "tool1", `{}`), NewEventCompletionEnded()},
		{NewEventCompletionEnded()},
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
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewToken("test"),
		NewEventCompletionEnded(),
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
	mockClient.SetStreamingEvents([]StreamEvent{
		NewEventNewRefusal("I cannot help with that"),
		NewEventCompletionEnded(),
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
		Tools:    &chatTools,
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

	// Round 1: tool call, Round 2: empty completion to finish
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`), NewEventCompletionEnded()},
		{NewEventCompletionEnded()},
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
		Tools:    &chatTools,
	}

	tool, err := tools.NewTool[map[string]string]("test-tool", "Test tool",
		func(input map[string]string) (string, error) {
			return "should not be called", nil
		})
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	chat.Tools.Add(tool)

	// Round 1: tool call, Round 2: empty completion to finish
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "test-tool", `{"key": "value"}`), NewEventCompletionEnded()},
		{NewEventCompletionEnded()},
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
		Tools:    &chatTools,
	}

	// Round 1: tool call, Round 2: empty completion to finish
	mockClient := NewMultiRoundMockClient([][]StreamEvent{
		{NewEventNewToolCall("call-1", "non-existent", `{}`), NewEventCompletionEnded()},
		{NewEventCompletionEnded()},
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
