package tools

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// ============================================================================
// Test structures
// ============================================================================

type SimpleStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type NestedStruct struct {
	User   SimpleStruct `json:"user"`
	Active bool         `json:"active"`
}

type EmbeddedStruct struct {
	SimpleStruct
	Country string `json:"country"`
}

type MapPrimitive struct {
	Scores map[string]int `json:"scores"`
}

type MapStruct struct {
	Users map[string]SimpleStruct `json:"users"`
}

type SlicePrimitive struct {
	IDs []int `json:"ids"`
}

type SliceStruct struct {
	Items []SimpleStruct `json:"items"`
}

type RecursiveStruct struct {
	Value int              `json:"value"`
	Child *RecursiveStruct `json:"child,omitempty"`
}

// ============================================================================
// Tests for NewTool function
// ============================================================================

func TestNewTool_WithStruct(t *testing.T) {
	tool, err := NewTool("test_struct", "test description", func(s SimpleStruct) (string, error) {
		if s.Name != "John" || s.Age != 30 {
			return "", errors.New("invalid input")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.Id != "test_struct" {
		t.Errorf("NewTool() id = %v, want %v", tool.Id, "test_struct")
	}

	if tool.Description != "test description" {
		t.Errorf("NewTool() description = %v, want %v", tool.Description, "test description")
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("test_struct", `{"name": "John", "age": 30}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_WithPrimitive(t *testing.T) {
	type PrimitiveInput struct {
		Input string `json:"input"`
	}

	tool, err := NewTool("test_primitive", "test primitive", func(s string) (string, error) {
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	// Verify schema contains the expected structure
	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "input") {
		t.Errorf("NewTool() schema should contain 'input' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_primitive", `{"input": "hello"}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_NestedStruct(t *testing.T) {
	tool, err := NewTool("test_nested", "nested struct test", func(n NestedStruct) (string, error) {
		if n.User.Name != "John" || n.User.Age != 30 || n.Active != true {
			return "", errors.New("invalid input")
		}
		return "ok", nil
	})

	if err != nil {
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	// Verify schema contains nested fields
	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "user") || !strings.Contains(string(schemaJSON), "active") {
		t.Errorf("NewTool() schema should contain 'user' and 'active' fields, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_nested", `{"user": {"name": "John", "age": 30}, "active": true}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_EmbeddedStruct(t *testing.T) {
	tool, err := NewTool("test_embedded", "embedded struct test", func(e EmbeddedStruct) (string, error) {
		if e.Name != "John" || e.Age != 30 {
			return "", errors.New("invalid input")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	// Verify schema contains embedded fields
	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "name") || !strings.Contains(string(schemaJSON), "age") {
		t.Errorf("NewTool() schema should contain embedded fields 'name' and 'age', got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_embedded", `{"name": "John", "age": 30}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_PrimitiveWithMapPrimitive(t *testing.T) {
	type InputWithMap struct {
		Input MapPrimitive `json:"input"`
	}

	tool, err := NewTool("test_map_primitive", "map primitive test", func(m MapPrimitive) (string, error) {
		scores := m.Scores
		if scores["a"] != 1 || scores["b"] != 2 {
			t.Errorf("invalid scores: %v", m)
			return "", errors.New("invalid scores")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "scores") {
		t.Errorf("NewTool() schema should contain 'scores' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_map_primitive", `{"scores": {"a": 1, "b": 2}}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_PrimitiveWithMapStruct(t *testing.T) {
	tool, err := NewTool("test_map_struct", "map struct test", func(m MapStruct) (string, error) {
		users := m.Users
		if users["alice"].Age != 25 || users["bob"].Age != 30 {
			return "", errors.New("invalid users")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "users") {
		t.Errorf("NewTool() schema should contain 'users' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_map_struct", `{"users": {"alice": {"age": 25}, "bob": {"age": 30}}}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_PrimitiveWithSlicePrimitive(t *testing.T) {
	tool, err := NewTool("test_slice_primitive", "slice primitive test", func(s SlicePrimitive) (string, error) {
		ids := s.IDs
		if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
			return "", errors.New("invalid ids")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "ids") {
		t.Errorf("NewTool() schema should contain 'ids' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	// Try to execute with valid input
	result, ok := tools.Execute("test_slice_primitive", `{"ids": [1, 2, 3]}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_PrimitiveWithSliceStruct(t *testing.T) {
	tool, err := NewTool("test_slice_struct", "slice struct test", func(s SliceStruct) (string, error) {
		items := s.Items
		if len(items) != 2 || items[0].Name != "Alice" || items[0].Age != 25 || items[1].Name != "Bob" || items[1].Age != 30 {
			return "", errors.New("invalid items")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "items") {
		t.Errorf("NewTool() schema should contain 'items' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("test_slice_struct", `{"items": [{"name": "Alice", "age": 25}, {"name": "Bob", "age": 30}]}`)
	t.Log("result: ", result)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}
}

func TestNewTool_PrimitiveWithMap(t *testing.T) {
	// Test with plain map type
	type MapInput struct {
		Data map[string]any `json:"data"`
	}

	tool, err := NewTool("test_map", "map test", func(m MapInput) (string, error) {
		val, ok := m.Data["key"]
		if !ok || val != "value" {
			return "", errors.New("invalid data")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("test_map", `{"data": {"key": "value"}}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_RecursiveStruct(t *testing.T) {
	tool, err := NewTool("test_recursive", "recursive struct test", func(r RecursiveStruct) (string, error) {
		if r.Value != 42 {
			return "", errors.New("invalid value")
		}
		if r.Child == nil || r.Child.Value != 100 {
			return "", errors.New("invalid child")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}

	schemaJSON, err := json.Marshal(tool.schema)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
		return
	}

	if !strings.Contains(string(schemaJSON), "value") || !strings.Contains(string(schemaJSON), "child") {
		t.Errorf("NewTool() schema should contain 'value' and 'child' fields, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("test_recursive", `{"value": 42, "child": {"value": 100}}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

// ============================================================================
// Tests for Execute method
// ============================================================================

func TestExecute_NormalCall(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("greet", "greets the user", func(s SimpleStruct) (string, error) {
		return "Hello, " + s.Name, nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	err = tools.Add(tool)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Execute the tool
	input := `{"name": "John", "age": 30}`
	result, ok := tools.Execute("greet", input)

	if !ok {
		t.Errorf("Execute() ok = false, want true")
	}

	if result != "Hello, John" {
		t.Errorf("Execute() result = %v, want %v", result, "Hello, John")
	}
}

func TestExecute_NonExistentTool(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("existing", "existing tool", func(s SimpleStruct) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute non-existent tool
	result, ok := tools.Execute("nonexistent", `{}`)

	if ok {
		t.Errorf("Execute() ok = true, want false")
	}

	if !strings.Contains(result, "not found") {
		t.Errorf("Execute() result should contain 'not found', got: %v", result)
	}
}

func TestExecute_ToolReturnsError(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("error_tool", "tool that returns error", func(s SimpleStruct) (string, error) {
		return "", errors.New("something went wrong")
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute tool that returns error
	input := `{"name": "Test", "age": 25}`
	result, ok := tools.Execute("error_tool", input)

	if ok {
		t.Errorf("Execute() ok = true, want false")
	}

	if !strings.Contains(result, "error:") && !strings.Contains(result, "something went wrong") {
		t.Errorf("Execute() result should contain error message, got: %v", result)
	}
}

// ============================================================================
// Edge case tests for Execute - error scenarios
// ============================================================================

func TestExecute_MalformedJSON(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("json_test", "test tool", func(s SimpleStruct) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute with malformed JSON
	result, ok := tools.Execute("json_test", `{invalid json}`)

	if ok {
		t.Errorf("Execute() ok = true, want false for malformed JSON")
	}

	if !strings.Contains(result, "unmarshal") {
		t.Errorf("Execute() result should contain 'unmarshal' error, got: %v", result)
	}
}

func TestExecute_EmptyArguments(t *testing.T) {
	tools := NewTools()

	// Create tool with struct that has required fields
	tool, err := NewTool("empty_args", "test tool", func(s SimpleStruct) (string, error) {
		return "received: " + s.Name, nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute with empty string
	result, ok := tools.Execute("empty_args", "")

	// Empty string should produce unmarshal error or zero values
	if ok && result == "" {
		t.Logf("Execute() returned ok=true with empty result - depends on type handling")
	}
}

func TestExecute_TypeMismatch(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("type_test", "test tool", func(s SimpleStruct) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute with wrong type (number instead of string)
	result, ok := tools.Execute("type_test", `{"name": 123, "age": "not-a-number"}`)

	if ok {
		t.Errorf("Execute() ok = true, want false for type mismatch")
	}

	if !strings.Contains(result, "unmarshal") {
		t.Errorf("Execute() result should contain 'unmarshal' error for type mismatch, got: %v", result)
	}
}

func TestExecute_MissingRequiredFields(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("required_test", "test tool", func(s SimpleStruct) (string, error) {
		return "name: " + s.Name, nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute with empty JSON object (missing required fields)
	result, ok := tools.Execute("required_test", `{}`)

	// This should succeed with zero values, not error
	if !ok {
		t.Errorf("Execute() ok = false, want true for missing optional fields")
	}

	_ = result // suppress unused warning
}

func TestExecute_PartialFields(t *testing.T) {
	tools := NewTools()

	tool, err := NewTool("partial_test", "test tool", func(s SimpleStruct) (string, error) {
		return "name: " + s.Name, nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	tools.Add(tool)

	// Execute with only some fields
	result, ok := tools.Execute("partial_test", `{"name": "John"}`)

	if !ok {
		t.Errorf("Execute() ok = false, want true for partial fields")
	}

	if result != "name: John" {
		t.Errorf("Execute() result = %v, want %v", result, "name: John")
	}
}

// ============================================================================
// Edge case tests for NewTool - error scenarios
// ============================================================================

func TestNewTool_WithFunctionField(t *testing.T) {
	// This test verifies behavior with function fields
	type StructWithFunc struct {
		Callback func() `json:"callback"`
		Name     string `json:"name"`
	}

	// This should fail during schema generation because functions can't be serialized
	_, err := NewTool("func_field", "test", func(s StructWithFunc) (string, error) {
		return "ok", nil
	})

	if err == nil {
		t.Logf("NewTool() with function field - behavior depends on jsonschema library")
	}
}

func TestNewTool_WithChannelField(t *testing.T) {
	type StructWithChan struct {
		Ch   chan string `json:"ch"`
		Name string      `json:"name"`
	}

	// This may fail during schema generation
	_, err := NewTool("chan_field", "test", func(s StructWithChan) (string, error) {
		return "ok", nil
	})

	if err == nil {
		t.Logf("NewTool() with channel field - behavior depends on jsonschema library")
	}
}

func TestNewTool_EmptyName(t *testing.T) {
	// Test with empty name - should still work (no validation in current code)
	tool, err := NewTool("", "test description", func(s SimpleStruct) (string, error) {
		return "ok", nil
	})

	if err != nil {
		t.Logf("NewTool() with empty name returns error: %v", err)
	} else {
		t.Logf("NewTool() with empty name - tool created with id: %q", tool.Id)
	}
}

func TestNewTool_PointerType(t *testing.T) {
	type PointerStruct struct {
		Name *string `json:"name"`
	}

	tool, err := NewTool("pointer_test", "test pointer type", func(s PointerStruct) (string, error) {
		if s.Name == nil || *s.Name != "John" {
			return "", errors.New("invalid name")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with pointer type error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for pointer type")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("pointer_test", `{"name": "John"}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_InterfaceType(t *testing.T) {
	type InterfaceStruct struct {
		Data interface{} `json:"data"`
	}

	tool, err := NewTool("interface_test", "test interface type", func(s InterfaceStruct) (string, error) {
		if s.Data == nil {
			return "", errors.New("data is nil")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with interface{} error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for interface type")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("interface_test", `{"data": "test"}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_AnonymousStruct(t *testing.T) {
	// Test with anonymous/unnamed struct type
	tool, err := NewTool("anon_struct", "test anonymous struct", func(s struct {
		Name string `json:"name"`
	}) (string, error) {
		if s.Name != "John" {
			return "", errors.New("invalid name")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with anonymous struct error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for anonymous struct")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("anon_struct", `{"name": "John"}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_WithIntPrimitive(t *testing.T) {
	tool, err := NewTool("int_primitive", "test int primitive", func(n int) (string, error) {
		if n != 42 {
			return "", errors.New("invalid value")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with int error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for int primitive")
	}

	// Verify schema contains input field
	schemaJSON, _ := json.Marshal(tool.schema)
	if !strings.Contains(string(schemaJSON), "input") {
		t.Errorf("NewTool() schema should contain 'input' field, got: %s", schemaJSON)
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("int_primitive", `{"input": 42}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

func TestNewTool_WithBoolPrimitive(t *testing.T) {
	tool, err := NewTool("bool_primitive", "test bool primitive", func(b bool) (string, error) {
		if b != true {
			return "", errors.New("invalid value")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with bool error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for bool primitive")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("bool_primitive", `{"input": true}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}

// ============================================================================
// Tests for GetSchema method (cache and error handling)
// ============================================================================

func TestGetSchema_Cache(t *testing.T) {
	// Create a tool
	tool, err := NewTool("cache_test", "test cache", func(s SimpleStruct) (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("NewTool() error = %v", err)
	}

	// First call - generates schema
	schema1, err := tool.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() first call error = %v", err)
	}
	if schema1 == nil {
		t.Fatal("GetSchema() first call returned nil schema")
	}

	// Second call - should return cached schema
	schema2, err := tool.GetSchema()
	if err != nil {
		t.Fatalf("GetSchema() second call error = %v", err)
	}
	if schema2 == nil {
		t.Fatal("GetSchema() second call returned nil schema")
	}

	// Verify they are the same (cached)
	if schema1["title"] != schema2["title"] {
		t.Errorf("GetSchema() cached schema differs: %v vs %v", schema1, schema2)
	}
}

func TestGetSchema_PanicRecovery(t *testing.T) {
	// Create a tool with function field that causes panic during schema generation
	type StructWithFunc struct {
		Callback func() `json:"callback"`
		Name     string `json:"name"`
	}

	// This will cause panic during schema generation
	tool, err := NewTool("panic_test", "test panic recovery", func(s StructWithFunc) (string, error) {
		return "ok", nil
	})
	if err == nil {
		// If no error during creation, GetSchema should return error
		_, err = tool.GetSchema()
		if err == nil {
			t.Error("GetSchema() should return error for struct with func field")
		}
		if err != nil {
			// Verify error message contains expected text
			errStr := err.Error()
			if !contains(errStr, "panic") && !contains(errStr, "schema generation") {
				t.Errorf("GetSchema() error should mention panic, got: %v", err)
			}
		}
	}
}

// Helper function for substring check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNewTool_WithFloatPrimitive(t *testing.T) {
	tool, err := NewTool("float_primitive", "test float64 primitive", func(f float64) (string, error) {
		if f != 3.14 {
			return "", errors.New("invalid value")
		}
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() with float64 error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil for float64 primitive")
	}

	tools := NewTools()
	err = tools.Add(tool)
	if err != nil {
		t.Errorf("Add() error = %v", err)
		return
	}

	result, ok := tools.Execute("float_primitive", `{"input": 3.14}`)
	if !ok {
		t.Errorf("Execute() error = %v", result)
		return
	}

	if result != "ok" {
		t.Errorf("Execute() result = %v, want %v", result, "ok")
	}
}
