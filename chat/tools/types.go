package tools

import (
	"fmt"
	"reflect"
	"sync"
)

type tool struct {
	Id          string
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

func (t *Tools) Add(tool tool, opts ...AddOption) (added bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	config := &toolAddConfig{}
	for _, opt := range opts {
		opt(config)
	}

	id := tool.Id
	if config.overrideName != "" {
		id = config.overrideName
	}

	if config.autoIncrement {
		id = nextID(t.tools, id, config)
	} else {
		_, ok := t.tools[id]
		if ok {
			return false
		}
	}

	t.tools[id] = tool
	return true
}

func nextID(m map[string]tool, id string, config *toolAddConfig) string {
	_, exists := m[id]
	if !exists {
		return id
	}
	return nextID(m, fmt.Sprintf("%s_%d", id, config.startIncrement+1), config)
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
	for id, tool := range t.tools {
		tool.Id = id
		out = append(out, tool)
	}

	return out
}
