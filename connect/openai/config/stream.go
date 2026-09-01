package config

import "github.com/openai/openai-go/v3"

type StreamConfig struct {
	// If set, an additional chunk will be streamed before the `data: [DONE]` message.
	// The `usage` field on this chunk shows the token usage statistics for the entire
	// request, and the `choices` field will always be an empty array.
	//
	// All other chunks will also include a `usage` field, but with a null value.
	// **NOTE:** If the stream is interrupted, you may not receive the final usage
	// chunk which contains the total token usage for the request.
	IncludeUsage *bool `json:"include_usage,omitempty" jsonschema_description:"If set, an additional chunk is streamed before the data: [DONE] message showing token usage statistics for the entire request."`
	// When true, stream obfuscation will be enabled. Stream obfuscation adds random
	// characters to an `obfuscation` field on streaming delta events to normalize
	// payload sizes as a mitigation to certain side-channel attacks. These obfuscation
	// fields are included by default, but add a small amount of overhead to the data
	// stream. You can set `include_obfuscation` to false to optimize for bandwidth if
	// you trust the network links between your application and the OpenAI API.
	IncludeObfuscation *bool `json:"include_obfuscation,omitempty" jsonschema_description:"When true, stream obfuscation adds random characters to streaming delta events to normalize payload sizes as a side-channel mitigation. Set to false to optimize for bandwidth."`
}

func (sc *StreamConfig) ToSDK() openai.ChatCompletionStreamOptionsParam {
	out := openai.ChatCompletionStreamOptionsParam{}
	if sc.IncludeUsage != nil {
		out.IncludeUsage = openai.Bool(*sc.IncludeUsage)
	}
	if sc.IncludeObfuscation != nil {
		out.IncludeObfuscation = openai.Bool(*sc.IncludeObfuscation)
	}
	return out
}
