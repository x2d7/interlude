package config

import "github.com/openai/openai-go/v3"

// Not supported with latest reasoning models `o3` and `o4-mini`.
type Stop struct {
	// The type of stop sequence to use. Use "string" for a single stop sequence or
	// "array" for multiple stop sequences.
	Type string `json:"type"`
	// A single stop sequence string. The API will stop generating further tokens when
	// this sequence is encountered. The returned text will not contain the stop sequence.
	Value *string `json:"value,omitempty"`
	// Up to 4 sequences where the API will stop generating further tokens. The
	// returned text will not contain the stop sequence.
	Values []string `json:"values,omitempty"`
}

func (s *Stop) ToSDK() openai.ChatCompletionNewParamsStopUnion {
	switch s.Type {
	case "string":
		if s.Value != nil {
			return openai.ChatCompletionNewParamsStopUnion{
				OfString: openai.String(*s.Value),
			}
		}
	case "array":
		return openai.ChatCompletionNewParamsStopUnion{
			OfStringArray: s.Values,
		}
	}
	return openai.ChatCompletionNewParamsStopUnion{}
}
