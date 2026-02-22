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

func NewTools() *Tools {
	t := Tools{
		tools: make(map[string]tool),
	}

	return &t
}

func (t *Tools) Add(tool tool, opts ...AddOption) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// applying options to config
	config := &toolAddConfig{}
	for _, opt := range opts {
		opt(config)
	}

	// default option cases:
	// no WithAutoIncrement? start increment from 1
	if !config.changedStartIncrement {
		config.startIncrement = 1
	}

	id := tool.Id
	if config.overrideName != "" {
		id = config.overrideName
	}

	if id == "" {
		return ErrEmptyToolID
	}

	if config.autoIncrement {
		id = nextID(t.tools, id, config)
	} else {
		_, ok := t.tools[id]
		if ok {
			return ErrToolAlreadyExists
		}
	}

	t.tools[id] = tool
	return nil
}

func nextID(m map[string]tool, id string, config *toolAddConfig) string {
	_, exists := m[id]
	if !exists {
		return id
	}

	new_id := fmt.Sprintf("%s_%d", id, config.startIncrement)

	_, exists = m[new_id]
	if exists {
		newConfig := &toolAddConfig{
			startIncrement: config.startIncrement + 1,
		}

		return nextID(m, id, newConfig)
	}

	return new_id
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
