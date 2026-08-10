package config

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func (rf ResponseFormat) JSONSchema() *jsonschema.Schema {
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{"text", "json_object", "json_schema"},
	})
	props.Set("json_schema", &jsonschema.Schema{
		Type: "object",
		Properties: (func() *orderedmap.OrderedMap[string, *jsonschema.Schema] {
			p := orderedmap.New[string, *jsonschema.Schema]()
			p.Set("name", &jsonschema.Schema{Type: "string"})
			p.Set("description", &jsonschema.Schema{Type: "string"})
			p.Set("schema", &jsonschema.Schema{})
			p.Set("strict", &jsonschema.Schema{Type: "boolean"})
			return p
		})(),
		Required: []string{"name", "schema"},
	})
	return &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Required:    []string{"type"},
		Description: "The type of response format to use. Any of \"text\", \"json_object\", or \"json_schema\".",
	}
}

func (s Stop) JSONSchema() *jsonschema.Schema {
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("type", &jsonschema.Schema{
		Type: "string",
		Enum: []any{"string", "array"},
	})
	props.Set("value", &jsonschema.Schema{
		Type:        "string",
		Description: "A single stop sequence string. The API will stop generating further tokens when this sequence is encountered. The returned text will not contain the stop sequence.",
	})
	props.Set("values", &jsonschema.Schema{
		Type: "array",
		Items: &jsonschema.Schema{
			Type: "string",
		},
		Description: "Up to 4 sequences where the API will stop generating further tokens. The returned text will not contain the stop sequence.",
	})
	return &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Required:    []string{"type"},
		Description: "Stop sequences that tell the API when to stop generating tokens.",
	}
}

func (j JSONSchemaDef) JSONSchema() *jsonschema.Schema {
	props := orderedmap.New[string, *jsonschema.Schema]()
	props.Set("name", &jsonschema.Schema{
		Type:        "string",
		Description: "The name of the response format. Must be a-z, A-Z, 0-9, or contain underscores and dashes, with a maximum length of 64.",
	})
	props.Set("description", &jsonschema.Schema{
		Type:        "string",
		Description: "A description of what the response format is for, used by the model to determine how to respond in the format.",
	})
	props.Set("schema", &jsonschema.Schema{
		Description: "The schema for the response format, described as a JSON Schema object.",
	})
	props.Set("strict", &jsonschema.Schema{
		Type:        "boolean",
		Description: "Whether to enable strict schema adherence when generating the output.",
	})
	return &jsonschema.Schema{
		Type:        "object",
		Properties:  props,
		Required:    []string{"name", "schema"},
		Description: "Structured Outputs configuration options, including a JSON Schema.",
	}
}

// ProviderConfigSchema generates a JSON Schema document for ProviderConfig
// and returns it as a map[string]any.
func ProviderConfigSchema() (map[string]any, error) {
	reflector := jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}

	s := reflector.Reflect(&ProviderConfig{})
	b, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var schemaMap map[string]any
	if err := json.Unmarshal(b, &schemaMap); err != nil {
		return nil, err
	}

	return schemaMap, nil
}
