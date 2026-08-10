package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	openai_connect "github.com/x2d7/interlude/connect/openai"
	"github.com/x2d7/interlude/provider"
)

func ptr[T any](v T) *T { return &v }

func TestRoundtripMinimal(t *testing.T) {
	cfg := &ProviderConfig{
		Conn: Connection{
			Endpoint: "https://api.openai.com",
			APIKey:   "sk-test",
			Model:    "gpt-4o",
		},
	}

	data, err := provider.Serialize(ProviderName, cfg)
	require.NoError(t, err)

	client, err := provider.Deserialize(data, provider.DefaultRegistry)
	require.NoError(t, err)

	oc := client.(*openai_connect.OpenAIClient)
	assert.Equal(t, "https://api.openai.com", oc.Endpoint)
	assert.Equal(t, "sk-test", oc.APIKey)
	assert.Equal(t, "gpt-4o", oc.Model)
}

func TestRoundtripFull(t *testing.T) {
	temp := 0.7
	topP := 0.9
	maxTokens := int64(4096)
	seed := int64(42)
	freqPenalty := 0.5
	presencePenalty := -0.2
	logprobs := true
	topLogprobs := int64(5)
	store := false
	reasoning := "medium"
	tier := "flex"
	verbosity := "low"
	parallel := true

	cfg := &ProviderConfig{
		Conn: Connection{
			Endpoint: "https://api.openai.com",
			APIKey:   "sk-test",
			Model:    "gpt-4o",
			Org:      ptr("org-123"),
			Project:  ptr("proj-456"),
		},
		Gen: Generation{
			Temperature:       &temp,
			TopP:              &topP,
			MaxTokens:         &maxTokens,
			FrequencyPenalty:  &freqPenalty,
			PresencePenalty:   &presencePenalty,
			Logprobs:          &logprobs,
			TopLogprobs:       &topLogprobs,
			Store:             &store,
			ReasoningEffort:   &reasoning,
			ServiceTier:       &tier,
			Verbosity:         &verbosity,
			ParallelToolCalls: &parallel,
			Seed:              &seed,
			LogitBias:         map[string]int64{"791": 1},
		},
		Cache: Cache{
			PromptCacheKey:       ptr("cache-key"),
			PromptCacheRetention: ptr("24h"),
			SafetyIdentifier:     ptr("user-hash"),
		},
		Format: &ResponseFormat{
			Type: "json_schema",
			JSONSchemaDef: &JSONSchemaDef{
				Name:   "answer",
				Schema: map[string]any{"type": "object"},
				Strict: ptr(true),
			},
		},
		Stop: &Stop{
			Type:   "array",
			Values: []string{"\n", "[DONE]"},
		},
		Meta:   Metadata{"key": "value"},
		Stream: &StreamConfig{IncludeUsage: ptr(true)},
	}

	data, err := provider.Serialize(ProviderName, cfg)
	require.NoError(t, err)

	var env provider.ProviderEnvelope
	require.NoError(t, json.Unmarshal(data, &env))
	assert.Equal(t, "openai", env.Provider)

	client, err := provider.Deserialize(data, provider.DefaultRegistry)
	require.NoError(t, err)

	oc := client.(*openai_connect.OpenAIClient)
	assert.Equal(t, "https://api.openai.com", oc.Endpoint)
	assert.Equal(t, "gpt-4o", oc.Model)
	assert.True(t, oc.Params.Temperature.Valid())
	assert.Equal(t, 0.7, oc.Params.Temperature.Value)
	assert.True(t, oc.Params.Logprobs.Valid())
	assert.True(t, oc.Params.Logprobs.Value)
}

func TestRoundtripFormatText(t *testing.T) {
	cfg := &ProviderConfig{
		Conn:   Connection{Endpoint: "https://api.openai.com", APIKey: "sk", Model: "gpt-4o"},
		Format: &ResponseFormat{Type: "text"},
	}

	data, err := provider.Serialize(ProviderName, cfg)
	require.NoError(t, err)

	client, err := provider.Deserialize(data, provider.DefaultRegistry)
	require.NoError(t, err)

	oc := client.(*openai_connect.OpenAIClient)
	assert.NotNil(t, oc.Params.ResponseFormat.OfText)
}

func TestRoundtripFormatJSONObject(t *testing.T) {
	cfg := &ProviderConfig{
		Conn:   Connection{Endpoint: "https://api.openai.com", APIKey: "sk", Model: "gpt-4o"},
		Format: &ResponseFormat{Type: "json_object"},
	}

	data, err := provider.Serialize(ProviderName, cfg)
	require.NoError(t, err)

	client, err := provider.Deserialize(data, provider.DefaultRegistry)
	require.NoError(t, err)

	oc := client.(*openai_connect.OpenAIClient)
	assert.NotNil(t, oc.Params.ResponseFormat.OfJSONObject)
}

func TestRoundtripStopString(t *testing.T) {
	stop := "END"
	cfg := &ProviderConfig{
		Conn: Connection{Endpoint: "https://api.openai.com", APIKey: "sk", Model: "gpt-4o"},
		Stop: &Stop{Type: "string", Value: &stop},
	}

	data, err := provider.Serialize(ProviderName, cfg)
	require.NoError(t, err)

	client, err := provider.Deserialize(data, provider.DefaultRegistry)
	require.NoError(t, err)

	oc := client.(*openai_connect.OpenAIClient)
	assert.True(t, oc.Params.Stop.OfString.Valid())
	assert.Equal(t, "END", oc.Params.Stop.OfString.Value)
}
