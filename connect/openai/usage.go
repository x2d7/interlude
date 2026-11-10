package openai_connect

import (
	"encoding/json"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/x2d7/interlude/chat"
)

// CompletionMetaFromCompletion extracts completion metadata from a
// non-streaming chat completion. Absent fields stay zero or empty.
func CompletionMetaFromCompletion(completion openai.ChatCompletion) chat.CompletionMeta {
	finishReason := ""
	if len(completion.Choices) > 0 {
		finishReason = completion.Choices[0].FinishReason
	}
	return newCompletionMeta(completion.Model, completion.Usage, finishReason)
}

// newCompletionMeta builds completion metadata from an SDK usage block.
func newCompletionMeta(model string, usage openai.CompletionUsage, finishReason string) chat.CompletionMeta {
	meta := chat.CompletionMeta{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
		FinishReason:     finishReason,
		Model:            model,
	}

	metadata := make(map[string]string)
	for name, field := range usage.JSON.ExtraFields {
		switch name {
		case "cost":
			meta.CostUSD = parseUSD(field.Raw())
		case "opencode":
			parseOpencode(field.Raw(), &meta, metadata)
		default:
			metadata[name] = rawText(field.Raw())
		}
	}
	if len(metadata) > 0 {
		meta.Metadata = metadata
	}

	return meta
}

func usagePresent(usage openai.CompletionUsage) bool {
	return usage.JSON.PromptTokens.Valid() ||
		usage.JSON.CompletionTokens.Valid() ||
		usage.JSON.TotalTokens.Valid() ||
		len(usage.JSON.ExtraFields) > 0
}

// parseOpencode reads the vendor billing block that some OpenAI-compatible
// gateways add to usage. Malformed values are ignored.
func parseOpencode(raw string, meta *chat.CompletionMeta, metadata map[string]string) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return
	}

	for name, value := range fields {
		switch name {
		case "billing_source":
			var source string
			if json.Unmarshal(value, &source) == nil {
				meta.BillingSource = strings.TrimSpace(source)
			}
		case "list_cost":
			meta.ListCostUSD = parseUSD(string(value))
		default:
			metadata["opencode."+name] = rawText(string(value))
		}
	}
}

// parseUSD decodes a numeric JSON value. Absent or malformed values stay zero.
func parseUSD(raw string) float64 {
	var usd float64
	if err := json.Unmarshal([]byte(raw), &usd); err != nil {
		return 0
	}
	return usd
}

// rawText converts a raw JSON value into a compact string. JSON strings are
// unquoted; everything else keeps its JSON representation.
func rawText(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return ""
	}
	if raw[0] == '"' {
		var text string
		if err := json.Unmarshal([]byte(raw), &text); err == nil {
			return text
		}
	}
	return raw
}
