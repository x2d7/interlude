package tools

import (
	"testing"
)

func TestAddMethod(t *testing.T) {
	tests := []struct {
		name        string
		tool        tool
		expectedErr error
	}{
		{
			name:        "Standard Addition",
			tool:        tool{Id: "test", Description: "test tool"},
			expectedErr: nil,
		},
		{
			name:        "Duplicate Name",
			tool:        tool{Id: "test", Description: "test tool"},
			expectedErr: ErrToolAlreadyExists,
		},
		{
			name:        "Tool Without Name",
			tool:        tool{Id: "", Description: "test tool"},
			expectedErr: ErrEmptyToolID,
		},
		{
			name:        "Addition Without Description",
			tool:        tool{Id: "test_description", Description: ""},
			expectedErr: nil,
		},
	}

	tools := NewTools()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tools.Add(tt.tool)
			if err != tt.expectedErr {
				t.Errorf("Add() = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

func TestRemoveMethod(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		ok       bool
	}{
		{
			name:     "Non-existent Element",
			toolName: "nonexistent",
			ok:       false,
		},
		{
			name:     "Standard Removal",
			toolName: "existing",
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools()
			if tt.ok {
				tools.Add(tool{Id: tt.toolName, Description: "test tool"})
			}

			ok := tools.Remove(tt.toolName)
			if ok != tt.ok {
				t.Errorf("Remove() = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func TestSnapshotMethod(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Override Name",
		},
		{
			name: "Series of Add and Remove",
		},
		{
			name: "Safety of Ownership",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := NewTools()
			switch tt.name {
			case "Override Name":
				tool1 := tool{Id: "test", Description: "test tool"}
				tools.Add(tool1, WithOverrideName("override"))
				snapshot := tools.Snapshot()
				if snapshot[0].Id != "override" {
					t.Errorf("Snapshot() = %v, want %v", snapshot[0].Id, "override")
				}
			case "Series of Add and Remove":
				tool1 := tool{Id: "test1", Description: "test tool 1"}
				tool2 := tool{Id: "test2", Description: "test tool 2"}
				tools.Add(tool1)
				tools.Add(tool2)
				tools.Remove("test1")
				snapshot := tools.Snapshot()
				if len(snapshot) != 1 || snapshot[0].Id != "test2" {
					t.Errorf("Snapshot() = %v, want length 1 with Id 'test2'", snapshot)
				}
			case "Safety of Ownership":
				tool1 := tool{Id: "test", Description: "test tool"}
				tools.Add(tool1)
				snapshot := tools.Snapshot()
				snapshot[0].Id = "modified"
				if tools.Snapshot()[0].Id == "modified" {
					t.Errorf("Snapshot() returned reference, want copy")
				}
			}
		})
	}
}

// ============================================================================
// Tests for AddOptions - AutoIncrement
// ============================================================================

func TestAdd_AutoIncrement(t *testing.T) {
	tools := NewTools()

	// Add first tool with ID "test"
	err := tools.Add(tool{Id: "test", Description: "tool 1"}, WithAutoIncrement())
	if err != nil {
		t.Fatalf("Add() first tool error = %v", err)
	}

	// Add second tool with same ID - should auto-increment
	err = tools.Add(tool{Id: "test", Description: "tool 2"}, WithAutoIncrement())
	if err != nil {
		t.Fatalf("Add() second tool error = %v", err)
	}

	// Add third tool with same ID - should auto-increment again
	err = tools.Add(tool{Id: "test", Description: "tool 3"}, WithAutoIncrement())
	if err != nil {
		t.Fatalf("Add() third tool error = %v", err)
	}

	// Verify Snapshot returns correct auto-incremented names
	snapshot := tools.Snapshot()
	if len(snapshot) != 3 {
		t.Fatalf("Snapshot() length = %d, want 3", len(snapshot))
	}

	// Collect IDs to verify they are "test", "test_1", "test_2"
	ids := make(map[string]bool)
	for _, t := range snapshot {
		ids[t.Id] = true
	}

	t.Log("ids: ", ids)

	if !ids["test"] {
		t.Error("Snapshot() should contain 'test'")
	}
	if !ids["test_1"] {
		t.Error("Snapshot() should contain 'test_1'")
	}
	if !ids["test_2"] {
		t.Error("Snapshot() should contain 'test_2'")
	}
}

func TestAdd_AutoIncrementWithStartIncrement(t *testing.T) {
	tools := NewTools()

	// Add first tool with ID "test" and start increment from 0
	err := tools.Add(tool{Id: "test", Description: "tool 1"}, WithAutoIncrement(), WithStartIncrement(0))
	if err != nil {
		t.Fatalf("Add() first tool error = %v", err)
	}

	// Add second tool - should be test_0
	err = tools.Add(tool{Id: "test", Description: "tool 2"}, WithAutoIncrement(), WithStartIncrement(0))
	if err != nil {
		t.Fatalf("Add() second tool error = %v", err)
	}

	// Add third tool - should be test_1
	err = tools.Add(tool{Id: "test", Description: "tool 3"}, WithAutoIncrement(), WithStartIncrement(0))
	if err != nil {
		t.Fatalf("Add() third tool error = %v", err)
	}

	snapshot := tools.Snapshot()
	if len(snapshot) != 3 {
		t.Fatalf("Snapshot() length = %d, want 3", len(snapshot))
	}

	// Collect IDs - should be "test", "test_0", "test_1"
	ids := make(map[string]bool)
	for _, t := range snapshot {
		ids[t.Id] = true
	}

	if !ids["test"] {
		t.Error("Snapshot() should contain 'test'")
	}
	if !ids["test_0"] {
		t.Error("Snapshot() should contain 'test_0'")
	}
	if !ids["test_1"] {
		t.Error("Snapshot() should contain 'test_1'")
	}
}

func TestNextID_NotExists(t *testing.T) {
	// Test nextID when ID does not exist in the map
	m := make(map[string]tool)
	config := &toolAddConfig{startIncrement: 1}

	result := nextID(m, "new_tool", config)

	if result != "new_tool" {
		t.Errorf("nextID() = %v, want 'new_tool'", result)
	}
}

func TestNextID_Exists(t *testing.T) {
	// Test nextID when ID already exists in the map
	m := map[string]tool{
		"test":   {Id: "test", Description: "tool 1"},
		"test_1": {Id: "test_1", Description: "tool 2"},
		"test_2": {Id: "test_2", Description: "tool 3"},
	}
	config := &toolAddConfig{startIncrement: 1}

	result := nextID(m, "test", config)

	// Should return next available ID: test_3
	if result != "test_3" {
		t.Errorf("nextID() = %v, want 'test_3'", result)
	}
}

func TestNextID_ExistsWithStartIncrement(t *testing.T) {
	// Test nextID when ID exists and startIncrement is 0
	m := map[string]tool{
		"test":   {Id: "test", Description: "tool 1"},
		"test_0": {Id: "test_0", Description: "tool 2"},
		"test_1": {Id: "test_1", Description: "tool 3"},
	}
	config := &toolAddConfig{startIncrement: 0}

	result := nextID(m, "test", config)

	// Should return test_2
	if result != "test_2" {
		t.Errorf("nextID() = %v, want 'test_2'", result)
	}
}
