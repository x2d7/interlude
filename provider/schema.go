package provider

import (
	"encoding/json"

	"github.com/invopop/jsonschema"
)

// SchemaProvider is implemented by provider config types that can generate
// their own JSON Schema definition.
type SchemaProvider interface {
	// Schema returns the JSON Schema for this config as map[string]any.
	Schema() (map[string]any, error)
}

// SchemaHandler returns a JSON Schema for a provider's config.
type SchemaHandler func() (map[string]any, error)

// ReflectSchema generates a JSON Schema for any config struct using reflection.
// Use this for providers whose config doesn't implement SchemaProvider.
func ReflectSchema(config any) (map[string]any, error) {
	reflector := jsonschema.Reflector{
		Anonymous:      true,
		ExpandedStruct: true,
	}

	s := reflector.Reflect(config)
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
