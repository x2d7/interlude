package tools

import (
	"encoding/json"
	"reflect"
	"sync"

	"github.com/invopop/jsonschema"
)

// cache for created anonymous struct types: key = original type, value = struct type
var anonStructCache sync.Map // map[reflect.Type]reflect.Type

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

func ensureInputStructType[T any]() reflect.Type {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Struct {
		t = constructInputStructForType(t)
	}
	return t
}

func constructInputStructForType(elem reflect.Type) reflect.Type {
	sf := reflect.StructField{
		Type: elem,
		Name: "Input",
		Tag:  reflect.StructTag(`json:"input"`),
	}

	t := reflect.StructOf([]reflect.StructField{sf})
	anonStructCache.Store(elem, t)
	return t
}
