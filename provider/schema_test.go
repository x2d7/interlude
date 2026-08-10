package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Name    string `json:"name" jsonschema_description:"The name of the config"`
	Enabled bool   `json:"enabled"`
	Limit   int    `json:"limit,omitempty"`
}

func TestReflectSchema(t *testing.T) {
	schema, err := ReflectSchema(&testConfig{})
	assert.NoError(t, err)
	assert.NotNil(t, schema)

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	assert.True(t, ok, "properties should be a map")

	nameProp, ok := props["name"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "string", nameProp["type"])

	enabledProp, ok := props["enabled"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "boolean", enabledProp["type"])
}

func TestRegisterSchema(t *testing.T) {
	reg := NewRegistry()

	handler := func() (map[string]any, error) {
		return map[string]any{"type": "object"}, nil
	}

	err := reg.RegisterSchema("test", handler)
	assert.NoError(t, err)

	got := reg.GetSchema("test")
	assert.NotNil(t, got)

	schema, err := got()
	assert.NoError(t, err)
	assert.Equal(t, "object", schema["type"])
}

func TestRegisterSchemaDuplicate(t *testing.T) {
	reg := NewRegistry()

	handler := func() (map[string]any, error) {
		return map[string]any{"type": "object"}, nil
	}

	err := reg.RegisterSchema("test", handler)
	assert.NoError(t, err)

	err = reg.RegisterSchema("test", handler)
	assert.Error(t, err)
}

func TestGetSchemaMissing(t *testing.T) {
	reg := NewRegistry()

	got := reg.GetSchema("nonexistent")
	assert.Nil(t, got)
}

func TestSchemaMissing(t *testing.T) {
	reg := NewRegistry()

	_, err := reg.Schema("nonexistent")
	assert.Error(t, err)
}

func TestSchema(t *testing.T) {
	reg := NewRegistry()

	handler := func() (map[string]any, error) {
		return map[string]any{"type": "string"}, nil
	}

	_ = reg.RegisterSchema("str", handler)

	schema, err := reg.Schema("str")
	assert.NoError(t, err)
	assert.Equal(t, "string", schema["type"])
}
