package provider

import (
	"encoding/json"
	"fmt"

	"github.com/x2d7/interlude/chat"
)

// Serialize marshals a config struct into a ProviderEnvelope JSON.
func Serialize(provider string, config any) ([]byte, error) {
	cfgBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("provider: marshal config for %q: %w", provider, err)
	}
	return json.Marshal(ProviderEnvelope{
		Provider: provider,
		Config:   cfgBytes,
	})
}

// Deserialize unmarshals JSON into a ProviderEnvelope and dispatches to the registered handler.
func Deserialize(data []byte, reg *Registry) (chat.Client, error) {
	var env ProviderEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("provider: unmarshal envelope: %w", err)
	}
	handler := reg.Get(env.Provider)
	if handler == nil {
		return nil, fmt.Errorf("provider: %q not registered", env.Provider)
	}
	return handler(env.Config)
}
