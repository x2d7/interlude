package tools

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/invopop/jsonschema"
)

type tool struct {
	Name        string
	Description string

	function  ToolFunction
	inputType reflect.Type
	schema    map[string]any
}

func (t *tool) GetSchema() (map[string]any, error) {
	if t.schema != nil {
		return t.schema, nil
	}

	inputType := t.inputType
	ptr := reflect.New(inputType).Interface()
	s := jsonschema.Reflect(ptr)
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var schemaMap map[string]any
	err = json.Unmarshal(b, &schemaMap)
	if err != nil {
		return nil, err
	}

	return schemaMap, nil
}

type ToolFunction func(input string) (string, error)

type Tools struct {
	mu   sync.RWMutex
	list []tool
}

func NewTools() Tools { return Tools{} }

// TODO: Переработка Tools - добавление и удаление по ID инструмента, получение списка инструментов (локальные Name)

func (t *Tools) Add(tool tool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.list = append(t.list, tool)
}

func (t *Tools) Snapshot() []tool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	out := make([]tool, len(t.list))
	copy(out, t.list)
	return out
}
