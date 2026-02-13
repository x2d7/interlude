package tools

import (
	"encoding/json"
	"fmt"
)

func NewTool[T any](name, description string, f func(T) (string, error)) (tool, error) {
	inputType := GetInputStructType[T]()

	wrapper := func(input string) (string, error) {
		raw := []byte(input)

		var parsed T
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return "", fmt.Errorf("unmarshal into %T: %w", parsed, err)
		}
		return f(parsed)
	}

	t := tool{
		Id:          name,
		Description: description,
		function:    wrapper,

		inputType: inputType,
	}

	schema, err := t.GetSchema()
	if err != nil {
		return tool{}, err
	}
	t.schema = schema

	return t, nil
}

func (t *Tools) Execute(name string, arguments string) (result string, ok bool) {
	for _, tool := range t.Snapshot() {
		if tool.Id != name {
			continue
		}

		result, err := tool.function(arguments)
		if err != nil {
			return err.Error(), false
		}
		return result, true
	}

	return fmt.Sprintf("error: tool %q not found", name), false
}
