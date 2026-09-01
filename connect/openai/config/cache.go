package config

type Cache struct {
	// Used by OpenAI to cache responses for similar requests to optimize your cache
	// hit rates. Replaces the `user` field.
	// [Learn more](https://platform.openai.com/docs/guides/prompt-caching).
	PromptCacheKey *string `json:"prompt_cache_key,omitempty" jsonschema_description:"Used by OpenAI to cache responses for similar requests to optimize cache hit rates. Replaces the user field."`
	// The retention policy for the prompt cache. Set to `24h` to enable extended
	// prompt caching, which keeps cached prefixes active for longer, up to a maximum
	// of 24 hours.
	// [Learn more](https://platform.openai.com/docs/guides/prompt-caching#prompt-cache-retention).
	//
	// Any of "in_memory", "24h".
	PromptCacheRetention *string `json:"prompt_cache_retention,omitempty" jsonschema_description:"The retention policy for the prompt cache. Set to \"24h\" to keep cached prefixes active for up to 24 hours. Any of \"in_memory\" or \"24h\"."`
	// A stable identifier used to help detect users of your application that may be
	// violating OpenAI's usage policies. The IDs should be a string that uniquely
	// identifies each user, with a maximum length of 64 characters. We recommend
	// hashing their username or email address, in order to avoid sending us any
	// identifying information.
	// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices#safety-identifiers).
	SafetyIdentifier *string `json:"safety_identifier,omitempty" jsonschema_description:"A stable identifier used to help detect users of your application that may be violating OpenAI's usage policies. Maximum length 64 characters."`
	// This field is being replaced by `safety_identifier` and `prompt_cache_key`. Use
	// `prompt_cache_key` instead to maintain caching optimizations. A stable
	// identifier for your end-users. Used to boost cache hit rates by better bucketing
	// similar requests and to help OpenAI detect and prevent abuse.
	// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices#safety-identifiers).
	User *string `json:"user,omitempty" jsonschema_description:"A stable identifier for your end-users, replaced by safety_identifier and prompt_cache_key. Use prompt_cache_key instead to maintain caching optimizations."`
}
