package tools

import (
	"reflect"
	"sync"
)

type Tools struct {
	mu   sync.RWMutex
	list []Tool
}

type Tool struct {
	Name        string
	Description string
	Func        ToolFunction

	InputType reflect.Type
	Schema    map[string]any
}

type ToolFunction func(input any) (string, error)

func NewTools() Tools { return Tools{} }

func (t *Tools) Add(tool Tool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.list = append(t.list, tool)
}

func (t *Tools) Snapshot() []Tool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]Tool, len(t.list))
	copy(out, t.list)
	return out
}
