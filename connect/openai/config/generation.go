package config

type Generation struct {
	// What sampling temperature to use, between 0 and 2. Higher values like 0.8 will
	// make the output more random, while lower values like 0.2 will make it more
	// focused and deterministic. We generally recommend altering this or `top_p` but
	// not both.
	Temperature *float64 `json:"temperature,omitempty" jsonschema_description:"What sampling temperature to use, between 0 and 2. Higher values make the output more random, while lower values make it more focused and deterministic."`
	// An alternative to sampling with temperature, called nucleus sampling, where the
	// model considers the results of the tokens with top_p probability mass. So 0.1
	// means only the tokens comprising the top 10% probability mass are considered.
	//
	// We generally recommend altering this or `temperature` but not both.
	TopP *float64 `json:"top_p,omitempty" jsonschema_description:"An alternative to sampling with temperature, called nucleus sampling, where the model considers tokens with top_p probability mass. We generally recommend altering this or temperature but not both."`
	// The maximum number of [tokens](/tokenizer) that can be generated in the chat
	// completion. This value can be used to control
	// [costs](https://openai.com/api/pricing/) for text generated via API.
	//
	// This value is now deprecated in favor of `max_completion_tokens`, and is not
	// compatible with
	// [o-series models](https://platform.openai.com/docs/guides/reasoning).
	MaxTokens *int64 `json:"max_tokens,omitempty" jsonschema_description:"The maximum number of tokens that can be generated. Deprecated in favor of max_completion_tokens and not compatible with o-series models."`
	// An upper bound for the number of tokens that can be generated for a completion,
	// including visible output tokens and
	// [reasoning tokens](https://platform.openai.com/docs/guides/reasoning).
	MaxCompletionTokens *int64 `json:"max_completion_tokens,omitempty" jsonschema_description:"An upper bound for the number of tokens that can be generated, including visible output tokens and reasoning tokens."`
	// How many chat completion choices to generate for each input message. Note that
	// you will be charged based on the number of generated tokens across all of the
	// choices. Keep `n` as `1` to minimize costs.
	N *int64 `json:"n,omitempty" jsonschema_description:"How many chat completion choices to generate for each input message. Keep n as 1 to minimize costs."`
	// This feature is in Beta. If specified, our system will make a best effort to
	// sample deterministically, such that repeated requests with the same `seed` and
	// parameters should return the same result. Determinism is not guaranteed, and you
	// should refer to the `system_fingerprint` response parameter to monitor changes
	// in the backend.
	Seed *int64 `json:"seed,omitempty" jsonschema_description:"If specified, the system will make a best effort to sample deterministically, such that repeated requests with the same seed and parameters return the same result."`
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on their
	// existing frequency in the text so far, decreasing the model's likelihood to
	// repeat the same line verbatim.
	FrequencyPenalty *float64 `json:"frequency_penalty,omitempty" jsonschema_description:"Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing frequency in the text so far, decreasing the likelihood of repeating the same line verbatim."`
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on
	// whether they appear in the text so far, increasing the model's likelihood to
	// talk about new topics.
	PresencePenalty *float64 `json:"presence_penalty,omitempty" jsonschema_description:"Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they appear in the text so far, increasing the likelihood of new topics."`
	// Whether to return log probabilities of the output tokens or not. If true,
	// returns the log probabilities of each output token returned in the `content` of
	// `message`.
	Logprobs *bool `json:"logprobs,omitempty" jsonschema_description:"Whether to return log probabilities of the output tokens."`
	// An integer between 0 and 20 specifying the maximum number of most likely tokens
	// to return at each token position, each with an associated log probability. In
	// some cases, the number of returned tokens may be fewer than requested.
	// `logprobs` must be set to `true` if this parameter is used.
	TopLogprobs *int64 `json:"top_logprobs,omitempty" jsonschema_description:"An integer between 0 and 20 specifying the maximum number of most likely tokens to return at each position, with associated log probabilities. Requires logprobs to be true."`
	// Whether or not to store the output of this chat completion request for use in
	// our [model distillation](https://platform.openai.com/docs/guides/distillation)
	// or [evals](https://platform.openai.com/docs/guides/evals) products.
	//
	// Supports text and image inputs. Note: image inputs over 8MB will be dropped.
	Store *bool `json:"store,omitempty" jsonschema_description:"Whether to store the output of this request for use in model distillation or evals products."`
	// Constrains effort on reasoning for
	// [reasoning models](https://platform.openai.com/docs/guides/reasoning). Currently
	// supported values are `none`, `minimal`, `low`, `medium`, `high`, and `xhigh`.
	// Reducing reasoning effort can result in faster responses and fewer tokens used
	// on reasoning in a response.
	//
	//   - `gpt-5.1` defaults to `none`, which does not perform reasoning. The supported
	//     reasoning values for `gpt-5.1` are `none`, `low`, `medium`, and `high`. Tool
	//     calls are supported for all reasoning values in gpt-5.1.
	//   - All models before `gpt-5.1` default to `medium` reasoning effort, and do not
	//     support `none`.
	//   - The `gpt-5-pro` model defaults to (and only supports) `high` reasoning effort.
	//   - `xhigh` is supported for all models after `gpt-5.1-codex-max`.
	//
	// Any of "none", "minimal", "low", "medium", "high", "xhigh".
	ReasoningEffort *string `json:"reasoning_effort,omitempty" jsonschema_description:"Constrains effort on reasoning for reasoning models. Any of \"none\", \"minimal\", \"low\", \"medium\", \"high\", or \"xhigh\"."`
	// Specifies the processing type used for serving the request.
	//
	//   - If set to 'auto', then the request will be processed with the service tier
	//     configured in the Project settings. Unless otherwise configured, the Project
	//     will use 'default'.
	//   - If set to 'default', then the request will be processed with the standard
	//     pricing and performance for the selected model.
	//   - If set to '[flex](https://platform.openai.com/docs/guides/flex-processing)' or
	//     '[priority](https://openai.com/api-priority-processing/)', then the request
	//     will be processed with the corresponding service tier.
	//   - When not set, the default behavior is 'auto'.
	//
	// When the `service_tier` parameter is set, the response body will include the
	// `service_tier` value based on the processing mode actually used to serve the
	// request. This response value may be different from the value set in the
	// parameter.
	//
	// Any of "auto", "default", "flex", "scale", "priority".
	ServiceTier *string `json:"service_tier,omitempty" jsonschema_description:"Specifies the processing type used for serving the request. Any of \"auto\", \"default\", \"flex\", \"scale\", or \"priority\"."`
	// Constrains the verbosity of the model's response. Lower values will result in
	// more concise responses, while higher values will result in more verbose
	// responses. Currently supported values are `low`, `medium`, and `high`.
	//
	// Any of "low", "medium", "high".
	Verbosity *string `json:"verbosity,omitempty" jsonschema_description:"Constrains the verbosity of the model's response. Any of \"low\", \"medium\", or \"high\"."`
	// Whether to enable
	// [parallel function calling](https://platform.openai.com/docs/guides/function-calling#configuring-parallel-function-calling)
	// during tool use.
	ParallelToolCalls *bool `json:"parallel_tool_calls,omitempty" jsonschema_description:"Whether to enable parallel function calling during tool use."`
	// Modify the likelihood of specified tokens appearing in the completion.
	//
	// Accepts a JSON object that maps tokens (specified by their token ID in the
	// tokenizer) to an associated bias value from -100 to 100. Mathematically, the
	// bias is added to the logits generated by the model prior to sampling. The exact
	// effect will vary per model, but values between -1 and 1 should decrease or
	// increase likelihood of selection; values like -100 or 100 should result in a ban
	// or exclusive selection of the relevant token.
	LogitBias map[string]int64 `json:"logit_bias,omitempty" jsonschema_description:"Modifies the likelihood of specified tokens appearing in the completion. Maps token IDs to a bias value from -100 to 100."`
}
