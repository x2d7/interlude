package config

import (
	"encoding/json"

	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/provider"
)

// ProviderName is the identifier used to register the OpenAI provider in the
// default provider registry.
const ProviderName = "openai"

func init() {
	_ = provider.DefaultRegistry.Register(ProviderName, func(data []byte) (chat.Client, error) {
		var cfg ProviderConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
		return cfg.ToClient(), nil
	})
	_ = provider.DefaultRegistry.RegisterSchema(ProviderName, ProviderConfigSchema)
}
