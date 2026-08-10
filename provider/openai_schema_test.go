package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/x2d7/interlude/connect/openai/config"
	"github.com/x2d7/interlude/provider"
)

func TestOpenAISchemaRegistered(t *testing.T) {
	handler := provider.DefaultRegistry.GetSchema("openai")
	assert.NotNil(t, handler, "openai schema handler should be registered")

	schema, err := provider.DefaultRegistry.Schema("openai")
	assert.NoError(t, err)
	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])
}
