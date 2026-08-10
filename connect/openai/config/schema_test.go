package config

import (
	"encoding/json"
	"testing"
)

func TestProviderConfigSchema(t *testing.T) {
	schema, err := ProviderConfigSchema()
	if err != nil {
		t.Fatalf("ProviderConfigSchema() error: %v", err)
	}

	// Verify top-level structure
	if schema["type"] != "object" {
		t.Errorf("expected type=object, got %v", schema["type"])
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties to be map[string]any")
	}

	// Verify required top-level properties exist
	requiredProps := []string{"conn", "gen", "cache"}
	for _, p := range requiredProps {
		if _, exists := props[p]; !exists {
			t.Errorf("missing property %q in schema", p)
		}
	}

	// Verify optional properties exist
	optionalProps := []string{"format", "stop", "meta", "stream"}
	for _, p := range optionalProps {
		if _, exists := props[p]; !exists {
			t.Errorf("missing optional property %q in schema", p)
		}
	}

	// Verify Connection sub-schema (may be inlined or $ref)
	connSchema := props["conn"]
	if connMap, ok := connSchema.(map[string]any); ok {
		if connMap["$ref"] != nil {
			// $ref - just verify it references something
			t.Logf("conn uses $ref: %v", connMap["$ref"])
		} else {
			// Inlined
			connProps, ok := connMap["properties"].(map[string]any)
			if !ok {
				t.Fatal("expected conn.properties to be map[string]any")
			}
			for _, field := range []string{"endpoint", "api_key", "model"} {
				if _, exists := connProps[field]; !exists {
					t.Errorf("missing conn property %q", field)
				}
			}
		}
	}

	// Verify Generation sub-schema
	genSchema := props["gen"]
	if genMap, ok := genSchema.(map[string]any); ok && genMap["$ref"] != nil {
		t.Logf("gen uses $ref: %v", genMap["$ref"])
	} else if genMap, ok := genSchema.(map[string]any); ok {
		genProps, ok := genMap["properties"].(map[string]any)
		if !ok {
			t.Fatal("expected gen.properties to be map[string]any")
		}
		for _, field := range []string{"temperature", "top_p", "max_tokens"} {
			if _, exists := genProps[field]; !exists {
				t.Errorf("missing gen property %q", field)
			}
		}
	}

	// Verify $defs contains ResponseFormat and Stop
	defs, ok := schema["$defs"].(map[string]any)
	if !ok {
		t.Fatal("expected $defs to be map[string]any")
	}

	// Check ResponseFormat has type enum
	rfDef, ok := defs["ResponseFormat"].(map[string]any)
	if !ok {
		t.Fatal("expected ResponseFormat in $defs")
	}
	rfProps, ok := rfDef["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected ResponseFormat.properties")
	}
	typeField, ok := rfProps["type"].(map[string]any)
	if !ok {
		t.Fatal("expected ResponseFormat.type property")
	}
	enum, ok := typeField["enum"].([]any)
	if !ok {
		t.Fatal("expected ResponseFormat.type.enum")
	}
	if len(enum) != 3 {
		t.Errorf("expected 3 enum values for ResponseFormat.type, got %d", len(enum))
	}

	// Check Stop has type enum
	stopDef, ok := defs["Stop"].(map[string]any)
	if !ok {
		t.Fatal("expected Stop in $defs")
	}
	stopProps, ok := stopDef["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected Stop.properties")
	}
	stopTypeField, ok := stopProps["type"].(map[string]any)
	if !ok {
		t.Fatal("expected Stop.type property")
	}
	stopEnum, ok := stopTypeField["enum"].([]any)
	if !ok {
		t.Fatal("expected Stop.type.enum")
	}
	if len(stopEnum) != 2 {
		t.Errorf("expected 2 enum values for Stop.type, got %d", len(stopEnum))
	}

	// Verify JSONSchemaDef is inlined in ResponseFormat (not in $defs since it's inlined)
	if jsSchema, ok := rfProps["json_schema"].(map[string]any); ok {
		if jsProps, ok := jsSchema["properties"].(map[string]any); ok {
			if _, exists := jsProps["schema"]; !exists {
				t.Error("missing json_schema.schema property")
			}
		}
	}

	// Pretty-print for debugging
	j, _ := json.MarshalIndent(schema, "", "  ")
	t.Logf("Schema:\n%s", string(j))
}

func TestResponseFormatJSONSchema(t *testing.T) {
	schema := (ResponseFormat{}).JSONSchema()

	if schema.Type != "object" {
		t.Errorf("expected type=object, got %v", schema.Type)
	}

	if len(schema.Required) != 1 || schema.Required[0] != "type" {
		t.Errorf("expected required=[type], got %v", schema.Required)
	}

	typeProp, ok := schema.Properties.Get("type")
	if !ok {
		t.Fatal("missing type property")
	}
	if typeProp.Type != "string" {
		t.Errorf("expected type prop type=string, got %v", typeProp.Type)
	}
	if len(typeProp.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(typeProp.Enum))
	}
}

func TestStopJSONSchema(t *testing.T) {
	schema := (Stop{}).JSONSchema()

	if schema.Type != "object" {
		t.Errorf("expected type=object, got %v", schema.Type)
	}

	if len(schema.Required) != 1 || schema.Required[0] != "type" {
		t.Errorf("expected required=[type], got %v", schema.Required)
	}

	typeProp, ok := schema.Properties.Get("type")
	if !ok {
		t.Fatal("missing type property")
	}
	if len(typeProp.Enum) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(typeProp.Enum))
	}

	valuesProp, ok := schema.Properties.Get("values")
	if !ok {
		t.Fatal("missing values property")
	}
	if valuesProp.Type != "array" {
		t.Errorf("expected values type=array, got %v", valuesProp.Type)
	}
}

func TestJSONSchemaDefJSONSchema(t *testing.T) {
	schema := (JSONSchemaDef{}).JSONSchema()

	if schema.Type != "object" {
		t.Errorf("expected type=object, got %v", schema.Type)
	}

	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}
	if !requiredSet["name"] || !requiredSet["schema"] {
		t.Errorf("expected name and schema to be required, got %v", schema.Required)
	}

	// The "schema" property should exist and not have a specific type (allows any)
	schemaProp, ok := schema.Properties.Get("schema")
	if !ok {
		t.Error("missing schema property")
	} else if schemaProp.Type != "" {
		t.Errorf("expected schema property to have no type constraint, got %v", schemaProp.Type)
	}
}
