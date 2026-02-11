package tools

import (
	"sync"
)

type Tools struct {
	mu   sync.RWMutex
	list []Tool
}

// TODO: Не экспортировать Schema, сделать GetSchema.

type Tool struct {
	Name        string
	Description string
	Func        ToolFunction
	Schema      map[string]any
}

type ToolFunction func(input any) (string, error)

func NewTools() Tools { return Tools{} }

// TODO: Переработка Tools - добавление и удаление по ID инструмента, получение списка инструментов (локальные Name)

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
