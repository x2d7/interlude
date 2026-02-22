package openai_connect

import (
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/chat/tools"
)

// ==================== isEmpty Tests ====================

func TestIsEmpty_EmptyStruct(t *testing.T) {
	m := openai.ChatCompletionMessageParamUnion{}
	if !isEmpty(m) {
		t.Error("Expected isEmpty=true for zero-value struct")
	}
}

func TestIsEmpty_WithUser(t *testing.T) {
	msg := openai.UserMessage("hello")
	if isEmpty(msg) {
		t.Error("Expected isEmpty=false for OfUser message")
	}
}

func TestIsEmpty_WithAssistant(t *testing.T) {
	msg := openai.AssistantMessage("hello")
	if isEmpty(msg) {
		t.Error("Expected isEmpty=false for OfAssistant message")
	}
}

func TestIsEmpty_WithSystem(t *testing.T) {
	msg := openai.SystemMessage("hello")
	if isEmpty(msg) {
		t.Error("Expected isEmpty=false for OfSystem message")
	}
}

func TestIsEmpty_WithTool(t *testing.T) {
	msg := openai.ToolMessage("result", "call-id")
	if isEmpty(msg) {
		t.Error("Expected isEmpty=false for OfTool message")
	}
}

// ==================== openAIMessages.findLastAssistantMessage Tests ====================

func TestFindLastAssistantMessage_EmptyList(t *testing.T) {
	m := openAIMessages{}
	result := m.findLastAssistantMessage()
	if result != nil {
		t.Errorf("Expected nil for empty list, got %v", result)
	}
}

func TestFindLastAssistantMessage_OnlyUserAndSystem(t *testing.T) {
	m := openAIMessages{
		openai.UserMessage("hello"),
		openai.SystemMessage("system"),
	}
	result := m.findLastAssistantMessage()
	if result != nil {
		t.Errorf("Expected nil when no assistant message, got %v", result)
	}
}

func TestFindLastAssistantMessage_SingleAssistant(t *testing.T) {
	m := openAIMessages{
		openai.UserMessage("hello"),
		openai.AssistantMessage("response"),
	}
	result := m.findLastAssistantMessage()
	if result == nil {
		t.Fatal("Expected non-nil result for list with one assistant message")
	}
	if result.OfAssistant == nil {
		t.Error("Expected OfAssistant to be non-nil")
	}
}

func TestFindLastAssistantMessage_MultipleAssistant_ReturnsLast(t *testing.T) {
	m := openAIMessages{
		openai.AssistantMessage("first"),
		openai.UserMessage("user"),
		openai.AssistantMessage("last"),
	}
	result := m.findLastAssistantMessage()
	if result == nil {
		t.Fatal("Expected non-nil result")
	}
	// Should point to the last element
	if result != &m[2] {
		t.Error("Expected pointer to last assistant message")
	}
}

// ==================== openAIMessages.Add Tests ====================

func TestOpenAIMessages_Add_AssistantMessage(t *testing.T) {
	m := openAIMessages{}
	m.Add(chat.NewEventNewAssistantMessage("Hello, I am an assistant"))

	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfAssistant == nil {
		t.Error("Expected OfAssistant to be non-nil")
	}
}

func TestOpenAIMessages_Add_SystemMessage(t *testing.T) {
	m := openAIMessages{}
	m.Add(chat.NewEventNewSystemMessage("You are a helpful assistant"))

	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfSystem == nil {
		t.Error("Expected OfSystem to be non-nil")
	}
}

func TestOpenAIMessages_Add_UserMessage(t *testing.T) {
	m := openAIMessages{}
	m.Add(chat.NewEventNewUserMessage("Hello"))

	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfUser == nil {
		t.Error("Expected OfUser to be non-nil")
	}
}

func TestOpenAIMessages_Add_Refusal(t *testing.T) {
	m := openAIMessages{}
	m.Add(chat.NewEventNewRefusal("I cannot help with that"))

	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfAssistant == nil {
		t.Error("Expected OfAssistant to be non-nil for refusal")
	}
}

func TestOpenAIMessages_Add_ToolMessage(t *testing.T) {
	m := openAIMessages{}
	event := chat.NewEventNewToolMessage("call-id", "result content", true)
	m.Add(event)

	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfTool == nil {
		t.Error("Expected OfTool to be non-nil")
	}
}

func TestOpenAIMessages_Add_ToolCall_NoExistingAssistant(t *testing.T) {
	m := openAIMessages{}
	// No prior assistant message — should create a new assistant message
	event := chat.NewEventNewToolCall("call-1", "my_tool", `{"arg": "val"}`)
	m.Add(event)

	// Should have one assistant message with the tool call
	if len(m) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(m))
	}
	if m[0].OfAssistant == nil {
		t.Error("Expected OfAssistant to be non-nil")
	}
	if len(m[0].OfAssistant.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(m[0].OfAssistant.ToolCalls))
	}
	tc := m[0].OfAssistant.ToolCalls[0].OfFunction
	if tc.ID != "call-1" {
		t.Errorf("Expected tool call ID 'call-1', got '%s'", tc.ID)
	}
	if tc.Function.Name != "my_tool" {
		t.Errorf("Expected function name 'my_tool', got '%s'", tc.Function.Name)
	}
	if tc.Function.Arguments != `{"arg": "val"}` {
		t.Errorf("Expected arguments '{\"arg\": \"val\"}', got '%s'", tc.Function.Arguments)
	}
}

func TestOpenAIMessages_Add_ToolCall_WithExistingAssistant(t *testing.T) {
	m := openAIMessages{}
	// Add an existing assistant message first
	m.Add(chat.NewEventNewAssistantMessage("I'll call a tool"))

	event := chat.NewEventNewToolCall("call-1", "my_tool", `{}`)
	m.Add(event)

	// Should still be 1 message (tool call appended to existing assistant)
	if len(m) != 1 {
		t.Fatalf("Expected 1 message (tool call merged into assistant), got %d", len(m))
	}
	if len(m[0].OfAssistant.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call on existing assistant, got %d", len(m[0].OfAssistant.ToolCalls))
	}
}

func TestOpenAIMessages_Add_TwoToolCalls_MergedIntoOneAssistant(t *testing.T) {
	m := openAIMessages{}
	m.Add(chat.NewEventNewToolCall("call-1", "tool_a", `{}`))
	m.Add(chat.NewEventNewToolCall("call-2", "tool_b", `{}`))

	// Both tool calls should be merged into one assistant message
	if len(m) != 1 {
		t.Fatalf("Expected 1 message (both tool calls in one assistant), got %d", len(m))
	}
	if len(m[0].OfAssistant.ToolCalls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(m[0].OfAssistant.ToolCalls))
	}
}

func TestOpenAIMessages_Add_UnknownEventType_NoMessageAdded(t *testing.T) {
	m := openAIMessages{}
	// EventNewToken is not handled in Add — should be silently ignored
	m.Add(chat.NewEventNewToken("some token"))

	if len(m) != 0 {
		t.Errorf("Expected 0 messages for unknown event type, got %d", len(m))
	}
}

func TestOpenAIMessages_Add_Sequence(t *testing.T) {
	m := openAIMessages{}

	m.Add(chat.NewEventNewSystemMessage("System prompt"))
	m.Add(chat.NewEventNewUserMessage("Hello"))
	m.Add(chat.NewEventNewAssistantMessage("Hi there"))
	m.Add(chat.NewEventNewUserMessage("How are you?"))

	if len(m) != 4 {
		t.Fatalf("Expected 4 messages, got %d", len(m))
	}
	if m[0].OfSystem == nil {
		t.Error("Expected first message to be system")
	}
	if m[1].OfUser == nil {
		t.Error("Expected second message to be user")
	}
	if m[2].OfAssistant == nil {
		t.Error("Expected third message to be assistant")
	}
	if m[3].OfUser == nil {
		t.Error("Expected fourth message to be user")
	}
}

// ==================== ConvertTools Tests ====================

func TestConvertTools_EmptyTools(t *testing.T) {
	ts := tools.NewTools()
	result := ConvertTools(ts)

	if len(result) != 0 {
		t.Errorf("Expected empty slice for empty tools, got %d elements", len(result))
	}
}

func TestConvertTools_SingleTool(t *testing.T) {
	ts := tools.NewTools()

	tool, err := tools.NewTool("weather", "Get current weather", func(input struct {
		City string `json:"city"`
	}) (string, error) {
		return "sunny", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}
	ts.Add(tool)

	result := ConvertTools(ts)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result))
	}
	if result[0].OfFunction == nil {
		t.Fatal("Expected OfFunction to be non-nil")
	}
	def := result[0].OfFunction.Function
	if def.Name != "weather" {
		t.Errorf("Expected name 'weather', got '%s'", def.Name)
	}
	if !def.Description.Valid() || def.Description.Value != "Get current weather" {
		t.Errorf("Expected description 'Get current weather', got %v", def.Description)
	}
}

func TestConvertTools_MultipleTools(t *testing.T) {
	ts := tools.NewTools()

	for _, name := range []string{"tool_a", "tool_b", "tool_c"} {
		n := name
		tool, err := tools.NewTool(n, "description of "+n, func(input struct{}) (string, error) {
			return "ok", nil
		})
		if err != nil {
			t.Fatalf("NewTool(%s) error = %v", n, err)
		}
		ts.Add(tool)
	}

	result := ConvertTools(ts)

	if len(result) != 3 {
		t.Fatalf("Expected 3 tools, got %d", len(result))
	}

	// Collect names
	names := make(map[string]bool)
	for _, r := range result {
		if r.OfFunction != nil {
			names[r.OfFunction.Function.Name] = true
		}
	}
	for _, expected := range []string{"tool_a", "tool_b", "tool_c"} {
		if !names[expected] {
			t.Errorf("Expected tool '%s' in result, but not found", expected)
		}
	}
}

func TestConvertTools_ParametersIncluded(t *testing.T) {
	ts := tools.NewTools()

	type SearchInput struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}

	tool, err := tools.NewTool("search", "Search for items", func(input SearchInput) (string, error) {
		return "results", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}
	ts.Add(tool)

	result := ConvertTools(ts)

	if len(result) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(result))
	}
	// Parameters should be non-nil (schema was generated)
	if result[0].OfFunction.Function.Parameters == nil {
		t.Error("Expected Parameters to be non-nil")
	}
}

// ==================== OpenAIClient.SyncInput Tests ====================

func TestSyncInput_ReturnsNewInstance(t *testing.T) {
	original := &OpenAIClient{
		Model:  "gpt-4o",
		APIKey: "test-key",
	}

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &tools.Tools{},
	}

	result := original.SyncInput(c)

	if result == original {
		t.Error("SyncInput should return a new instance, not the original")
	}
}

func TestSyncInput_OriginalNotModified(t *testing.T) {
	original := &OpenAIClient{
		Model:  "gpt-4o",
		APIKey: "test-key",
	}

	originalParamsMsgCount := len(original.Params.Messages)

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &tools.Tools{},
	}
	c.AddMessage(chat.SenderUser{}, "Hello")

	original.SyncInput(c)

	// Original params should not be modified
	if len(original.Params.Messages) != originalParamsMsgCount {
		t.Error("SyncInput should not modify the original client's Params.Messages")
	}
}

func TestSyncInput_MessagesAreSynced(t *testing.T) {
	original := &OpenAIClient{
		Model: "gpt-4o",
	}

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &tools.Tools{},
	}
	c.AddMessage(chat.SenderSystem{}, "System prompt")
	c.AddMessage(chat.SenderUser{}, "Hello")
	c.AddMessage(chat.SenderAssistant{}, "Hi there")

	result := original.SyncInput(c)
	newClient := result.(*OpenAIClient)

	if len(newClient.Params.Messages) != 3 {
		t.Fatalf("Expected 3 messages synced, got %d", len(newClient.Params.Messages))
	}

	if newClient.Params.Messages[0].OfSystem == nil {
		t.Error("Expected first synced message to be system")
	}
	if newClient.Params.Messages[1].OfUser == nil {
		t.Error("Expected second synced message to be user")
	}
	if newClient.Params.Messages[2].OfAssistant == nil {
		t.Error("Expected third synced message to be assistant")
	}
}

func TestSyncInput_ToolsAreSynced(t *testing.T) {
	original := &OpenAIClient{
		Model: "gpt-4o",
	}

	ts := tools.NewTools()
	tool, err := tools.NewTool("my_tool", "My tool", func(input struct{}) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}
	ts.Add(tool)

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    ts,
	}

	result := original.SyncInput(c)
	newClient := result.(*OpenAIClient)

	if len(newClient.Params.Tools) != 1 {
		t.Fatalf("Expected 1 tool synced, got %d", len(newClient.Params.Tools))
	}
	if newClient.Params.Tools[0].OfFunction == nil {
		t.Error("Expected OfFunction to be non-nil for synced tool")
	}
	if newClient.Params.Tools[0].OfFunction.Function.Name != "my_tool" {
		t.Errorf("Expected tool name 'my_tool', got '%s'", newClient.Params.Tools[0].OfFunction.Function.Name)
	}
}

func TestSyncInput_EmptyChatProducesEmptyParams(t *testing.T) {
	original := &OpenAIClient{
		Model: "gpt-4o",
	}

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &tools.Tools{},
	}

	result := original.SyncInput(c)
	newClient := result.(*OpenAIClient)

	if len(newClient.Params.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(newClient.Params.Messages))
	}
	if len(newClient.Params.Tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(newClient.Params.Tools))
	}
}

func TestSyncInput_PreservesModelAndAPIKey(t *testing.T) {
	original := &OpenAIClient{
		Model:    "gpt-4o-mini",
		APIKey:   "sk-test",
		Endpoint: "https://custom.endpoint.com",
	}

	c := &chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &tools.Tools{},
	}

	result := original.SyncInput(c)
	newClient := result.(*OpenAIClient)

	if newClient.Model != "gpt-4o-mini" {
		t.Errorf("Expected Model 'gpt-4o-mini', got '%s'", newClient.Model)
	}
	if newClient.APIKey != "sk-test" {
		t.Errorf("Expected APIKey 'sk-test', got '%s'", newClient.APIKey)
	}
	if newClient.Endpoint != "https://custom.endpoint.com" {
		t.Errorf("Expected Endpoint to be preserved, got '%s'", newClient.Endpoint)
	}
}
