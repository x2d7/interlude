package tools

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func NewTool[T any](name, description string, f func(T) (string, error)) (tool, error) {
	inputType := ensureInputStructType[T]()

	var extract func(reflect.Value) T
	if inputType == reflect.TypeFor[T]() {
		extract = func(v reflect.Value) T { return v.Interface().(T) }
	} else {
		extract = func(v reflect.Value) T { return v.Field(0).Interface().(T) }
	}

	wrapper := func(input string) (string, error) {
		ptr := reflect.New(inputType)
		if err := json.Unmarshal([]byte(input), ptr.Interface()); err != nil {
			return "", fmt.Errorf("unmarshal into %v: %w", inputType, err)
		}
		return f(extract(ptr.Elem()))
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
