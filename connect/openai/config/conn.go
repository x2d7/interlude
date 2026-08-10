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
	Endpoint string `json:"endpoint"`
	// The API key to use for authentication.
	APIKey string `json:"api_key"`
	// The model to use for generating completions.
	Model string `json:"model"`
	// ID of the organization to use for this request.
	Org *string `json:"org,omitempty"`
	// ID of the project to use for this request.
	Project *string `json:"project,omitempty"`
	// Maximum number of retries the client attempts to make. When given 0, the client
	// only makes one request. By default, the client retries two times.
	MaxRetries *int `json:"max_retries,omitempty"`
	// Timeout for each request attempt. This should be smaller than the timeout
	// defined in the context, which spans all retries.
	Timeout *string `json:"timeout,omitempty"`
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
