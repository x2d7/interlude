package config

import (
	"time"

	"github.com/openai/openai-go/v3/option"
)

type Connection struct {
	// The OpenAI API base URL. Use this to specify a custom base URL for the API,
	// such as a proxy or a different API endpoint.
	//
	// For security reasons, ensure that the base URL is trusted.
	Endpoint string `json:"endpoint" jsonschema_description:"The OpenAI API base URL. Use this to specify a custom base URL, such as a proxy. For security reasons, ensure that the base URL is trusted."`
	// The API key to use for authentication.
	APIKey string `json:"api_key" jsonschema_description:"The API key to use for authentication."`
	// The model to use for generating completions.
	Model string `json:"model" jsonschema_description:"The model to use for generating completions."`
	// ID of the organization to use for this request.
	Org *string `json:"org,omitempty" jsonschema_description:"ID of the organization to use for this request."`
	// ID of the project to use for this request.
	Project *string `json:"project,omitempty" jsonschema_description:"ID of the project to use for this request."`
	// Maximum number of retries the client attempts to make. When given 0, the client
	// only makes one request. By default, the client retries two times.
	MaxRetries *int `json:"max_retries,omitempty" jsonschema_description:"Maximum number of retries the client makes. When 0, only one request is made; by default the client retries two times."`
	// Timeout for each request attempt. This should be smaller than the timeout
	// defined in the context, which spans all retries.
	Timeout *string `json:"timeout,omitempty" jsonschema_description:"Timeout for each request attempt, smaller than the context timeout that spans all retries."`
}

func (c *Connection) toRequestOptions() []option.RequestOption {
	opts := make([]option.RequestOption, 0)
	if c.Endpoint != "" {
		opts = append(opts, option.WithBaseURL(c.Endpoint))
	}
	if c.APIKey != "" {
		opts = append(opts, option.WithAPIKey(c.APIKey))
	}
	if c.Org != nil {
		opts = append(opts, option.WithOrganization(*c.Org))
	}
	if c.Project != nil {
		opts = append(opts, option.WithProject(*c.Project))
	}
	if c.MaxRetries != nil {
		opts = append(opts, option.WithMaxRetries(*c.MaxRetries))
	}
	if c.Timeout != nil {
		if dur, err := time.ParseDuration(*c.Timeout); err == nil {
			opts = append(opts, option.WithRequestTimeout(dur))
		}
	}
	return opts
}
