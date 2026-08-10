package config

import (
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
	openai_connect "github.com/x2d7/interlude/connect/openai"
)

// ProviderConfig is the complete configuration for the OpenAI provider. It combines
// connection settings, generation parameters, cache options, response format, stop
// sequences, metadata, and streaming configuration.
type ProviderConfig struct {
	// Connection settings for the OpenAI API, including endpoint, API key, model,
	// organization, project, retries, and timeout.
	Conn Connection `json:"conn"`
	// Generation parameters controlling sampling, token limits, penalties, and other
	// model behavior.
	Gen Generation `json:"gen"`
	// Cache configuration for prompt caching and safety identifiers.
	Cache Cache `json:"cache"`
	// Response format configuration for structured outputs (text, JSON object, or JSON
	// schema).
	Format *ResponseFormat `json:"format,omitempty"`
	// Stop sequences that tell the API when to stop generating tokens.
	Stop *Stop `json:"stop,omitempty"`
	// Key-value metadata pairs attached to the request for tracking and querying.
	Meta Metadata `json:"meta,omitempty"`
	// Streaming configuration for controlling usage reporting and obfuscation.
	Stream *StreamConfig `json:"stream,omitempty"`
}

func (pc *ProviderConfig) ToClient() *openai_connect.OpenAIClient {
	return &openai_connect.OpenAIClient{
		Endpoint:       pc.Conn.Endpoint,
		APIKey:         pc.Conn.APIKey,
		Model:          pc.Conn.Model,
		Params:         pc.toChatCompletionParams(),
		RequestOptions: pc.Conn.toRequestOptions(),
	}
}

func (pc *ProviderConfig) toChatCompletionParams() openai.ChatCompletionNewParams {
	p := openai.ChatCompletionNewParams{}

	p.Temperature = optFloat64(pc.Gen.Temperature)
	p.TopP = optFloat64(pc.Gen.TopP)
	p.MaxTokens = optInt64(pc.Gen.MaxTokens)
	p.MaxCompletionTokens = optInt64(pc.Gen.MaxCompletionTokens)
	p.N = optInt64(pc.Gen.N)
	p.Seed = optInt64(pc.Gen.Seed)
	p.FrequencyPenalty = optFloat64(pc.Gen.FrequencyPenalty)
	p.PresencePenalty = optFloat64(pc.Gen.PresencePenalty)
	p.Logprobs = optBool(pc.Gen.Logprobs)
	p.TopLogprobs = optInt64(pc.Gen.TopLogprobs)
	p.Store = optBool(pc.Gen.Store)
	p.ParallelToolCalls = optBool(pc.Gen.ParallelToolCalls)

	if pc.Gen.ReasoningEffort != nil {
		p.ReasoningEffort = shared.ReasoningEffort(*pc.Gen.ReasoningEffort)
	}
	if pc.Gen.ServiceTier != nil {
		p.ServiceTier = openai.ChatCompletionNewParamsServiceTier(*pc.Gen.ServiceTier)
	}
	if pc.Gen.Verbosity != nil {
		p.Verbosity = openai.ChatCompletionNewParamsVerbosity(*pc.Gen.Verbosity)
	}

	if len(pc.Gen.LogitBias) > 0 {
		p.LogitBias = pc.Gen.LogitBias
	}

	p.PromptCacheKey = optString(pc.Cache.PromptCacheKey)
	if pc.Cache.PromptCacheRetention != nil {
		p.PromptCacheRetention = openai.ChatCompletionNewParamsPromptCacheRetention(*pc.Cache.PromptCacheRetention)
	}
	p.SafetyIdentifier = optString(pc.Cache.SafetyIdentifier)
	p.User = optString(pc.Cache.User)

	if len(pc.Meta) > 0 {
		p.Metadata = shared.Metadata(pc.Meta)
	}

	if pc.Format != nil {
		p.ResponseFormat = pc.Format.ToSDK()
	}

	if pc.Stop != nil {
		p.Stop = pc.Stop.ToSDK()
	}

	if pc.Stream != nil {
		p.StreamOptions = pc.Stream.ToSDK()
	}

	return p
}
