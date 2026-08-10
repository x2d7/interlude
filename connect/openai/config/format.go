package config

import (
	"github.com/openai/openai-go/v3"
)

// JSON Schema response format. Used to generate structured JSON responses. Learn
// more about
// [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs).
type ResponseFormat struct {
	// The type of response format to use. Any of "text", "json_object", or "json_schema".
	Type string `json:"type"`
	// Structured Outputs configuration options, including a JSON Schema. Only used
	// when type is "json_schema".
	JSONSchemaDef *JSONSchemaDef `json:"json_schema,omitempty"`
}

// Structured Outputs configuration options, including a JSON Schema.
type JSONSchemaDef struct {
	// The name of the response format. Must be a-z, A-Z, 0-9, or contain underscores
	// and dashes, with a maximum length of 64.
	Name string `json:"name"`
	// A description of what the response format is for, used by the model to determine
	// how to respond in the format.
	Description *string `json:"description,omitempty"`
	// The schema for the response format, described as a JSON Schema object. Learn how
	// to build JSON schemas [here](https://json-schema.org/).
	Schema any `json:"schema"`
	// Whether to enable strict schema adherence when generating the output. If set to
	// true, the model will always follow the exact schema defined in the `schema`
	// field. Only a subset of JSON Schema is supported when `strict` is `true`. To
	// learn more, read the
	// [Structured Outputs guide](https://platform.openai.com/docs/guides/structured-outputs).
	Strict *bool `json:"strict,omitempty"`
}

func (rf *ResponseFormat) ToSDK() openai.ChatCompletionNewParamsResponseFormatUnion {
	switch rf.Type {
	case "text":
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfText: &openai.ResponseFormatTextParam{Type: "text"},
		}
	case "json_object":
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &openai.ResponseFormatJSONObjectParam{Type: "json_object"},
		}
	case "json_schema":
		schema := openai.ResponseFormatJSONSchemaJSONSchemaParam{
			Name:   rf.JSONSchemaDef.Name,
			Schema: rf.JSONSchemaDef.Schema,
		}
		if rf.JSONSchemaDef.Description != nil {
			schema.Description = openai.String(*rf.JSONSchemaDef.Description)
		}
		if rf.JSONSchemaDef.Strict != nil {
			schema.Strict = openai.Bool(*rf.JSONSchemaDef.Strict)
		}
		return openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{
				Type:       "json_schema",
				JSONSchema: schema,
			},
		}
	default:
		return openai.ChatCompletionNewParamsResponseFormatUnion{}
	}
}
