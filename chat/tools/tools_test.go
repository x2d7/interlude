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
	Value  int             `json:"value"`
	Child  *RecursiveStruct `json:"child,omitempty"`
}

// ============================================================================
// Tests for NewTool function
// ============================================================================

func TestNewTool_WithStruct(t *testing.T) {
	tool, err := NewTool("test_struct", "test description", func(s SimpleStruct) (string, error) {
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
}

func TestNewTool_NestedStruct(t *testing.T) {
	tool, err := NewTool("test_nested", "nested struct test", func(n NestedStruct) (string, error) {
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
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
}

func TestNewTool_EmbeddedStruct(t *testing.T) {
	tool, err := NewTool("test_embedded", "embedded struct test", func(e EmbeddedStruct) (string, error) {
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
}

func TestNewTool_PrimitiveWithMapPrimitive(t *testing.T) {
	type InputWithMap struct {
		Input MapPrimitive `json:"input"`
	}

	tool, err := NewTool("test_map_primitive", "map primitive test", func(m MapPrimitive) (string, error) {
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
}

func TestNewTool_PrimitiveWithMapStruct(t *testing.T) {
	tool, err := NewTool("test_map_struct", "map struct test", func(m MapStruct) (string, error) {
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
}

func TestNewTool_PrimitiveWithSlicePrimitive(t *testing.T) {
	tool, err := NewTool("test_slice_primitive", "slice primitive test", func(s SlicePrimitive) (string, error) {
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
}

func TestNewTool_PrimitiveWithSliceStruct(t *testing.T) {
	tool, err := NewTool("test_slice_struct", "slice struct test", func(s SliceStruct) (string, error) {
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
}

func TestNewTool_PrimitiveWithMap(t *testing.T) {
	// Test with plain map type
	type MapInput struct {
		Data map[string]any `json:"data"`
	}

	tool, err := NewTool("test_map", "map test", func(m MapInput) (string, error) {
		return "ok", nil
	})

	if err != nil {
		t.Errorf("NewTool() error = %v", err)
		return
	}

	if tool.schema == nil {
		t.Error("NewTool() schema is nil")
	}
}

func TestNewTool_RecursiveStruct(t *testing.T) {
	tool, err := NewTool("test_recursive", "recursive struct test", func(r RecursiveStruct) (string, error) {
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
