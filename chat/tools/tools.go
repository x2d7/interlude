package tools

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/invopop/jsonschema"
)

// TODO: Поддержка более простых тулов, которые принимают на вход примитивные типы (учитывать сигнатуру)

func NewTool[T any](name, description string, f func(T) (string, error)) Tool {
	inputType := reflect.TypeFor[T]()

	ptr := reflect.New(inputType).Interface()
	s := jsonschema.Reflect(ptr)
	b, _ := json.Marshal(s)
	var schemaMap map[string]any
	_ = json.Unmarshal(b, &schemaMap)

	wrapper := func(input any) (string, error) {
		var raw []byte
		switch v := input.(type) {
		case string:
			raw = []byte(v)
		case []byte:
			raw = v
		default:
			return "", fmt.Errorf("tool %q expects a JSON string or []byte", name)
		}

		var parsed T
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return "", fmt.Errorf("unmarshal into %T: %w", parsed, err)
		}
		return f(parsed)
	}

	return Tool{
		Name:        name,
		Description: description,
		Func:        wrapper,
		Schema:      schemaMap,
	}
}

func (t *Tools) Execute(name string, arguments string) (result string, ok bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, tool := range t.list {
		if tool.Name != name {
			continue
		}

		result, err := tool.Func(arguments)
		if err != nil {
			return err.Error(), false
		}
		return result, true
	}

	return fmt.Sprintf("error: tool %q not found", name), false
}
