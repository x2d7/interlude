package provider

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/x2d7/interlude/chat"
)

func TestSerialize(t *testing.T) {
	cfg := map[string]any{"key": "value"}

	data, err := Serialize("test", cfg)
	assert.NoError(t, err)

	var env ProviderEnvelope
	err = json.Unmarshal(data, &env)
	assert.NoError(t, err)
	assert.Equal(t, "test", env.Provider)

	var cfgMap map[string]any
	err = json.Unmarshal(env.Config, &cfgMap)
	assert.NoError(t, err)
	assert.Equal(t, "value", cfgMap["key"])
}

func TestDeserialize(t *testing.T) {
	reg := NewRegistry()
	reg.Register("mock", func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	})

	envelope := ProviderEnvelope{
		Provider: "mock",
		Config:   json.RawMessage(`{"key":"value"}`),
	}
	data, _ := json.Marshal(envelope)

	client, err := Deserialize(data, reg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestDeserializeUnknownProvider(t *testing.T) {
	reg := NewRegistry()

	envelope := ProviderEnvelope{
		Provider: "unknown",
		Config:   json.RawMessage(`{}`),
	}
	data, _ := json.Marshal(envelope)

	client, err := Deserialize(data, reg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestDeserializeInvalidJSON(t *testing.T) {
	reg := NewRegistry()

	client, err := Deserialize([]byte("not json"), reg)
	assert.Error(t, err)
	assert.Nil(t, client)
}
