package tools

import (
	"encoding/json"
	"reflect"

	"github.com/invopop/jsonschema"
)

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

// TODO: Поддержка более простых тулов, которые принимают на вход примитивные типы (учитывать сигнатуру)

func GetInputStructType[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}
