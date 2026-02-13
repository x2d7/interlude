package tools

import (
	"reflect"
	"sync"
)

type tool struct {
	Name        string
	Description string

	function  toolFunction
	inputType reflect.Type
	schema    map[string]any
}

type toolFunction func(input string) (string, error)

type Tools struct {
	mu    sync.RWMutex
	tools map[string]tool
}

func NewTools() Tools {
	return Tools{
		tools: make(map[string]tool),
	}
}

func (t *Tools) Add(tool tool) (added bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, ok := t.tools[tool.Name]
	if ok {
		return false
	}
	t.tools[tool.Name] = tool
	return true
}

func (t *Tools) Remove(name string) (removed bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	_, ok := t.tools[name]
	if !ok {
		return false
	}
	delete(t.tools, name)
	return true
}

func (t *Tools) Snapshot() []tool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]tool, 0, len(t.tools))
	for _, tool := range t.tools {
		out = append(out, tool)
	}

	return out
}
