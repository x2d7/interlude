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
