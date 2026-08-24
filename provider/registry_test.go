package provider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/x2d7/interlude/chat"
)

type mockClient struct{}

func (m *mockClient) NewStreaming(ctx context.Context) chat.Stream[chat.StreamEvent] { return nil }
func (m *mockClient) SyncInput(c *chat.Chat) chat.Client                             { return m }

func TestRegister(t *testing.T) {
	reg := NewRegistry()

	handler := func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	}

	err := reg.Register("test", handler)
	assert.NoError(t, err)

	got := reg.Get("test")
	assert.NotNil(t, got)
}

func TestRegisterDuplicate(t *testing.T) {
	reg := NewRegistry()

	handler := func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	}

	err := reg.Register("test", handler)
	assert.NoError(t, err)

	err = reg.Register("test", handler)
	assert.Error(t, err)
}

func TestGetMissing(t *testing.T) {
	reg := NewRegistry()

	got := reg.Get("nonexistent")
	assert.Nil(t, got)
}

func TestProviders(t *testing.T) {
	reg := NewRegistry()

	// alpha has a handler and a schema, so it is creatable.
	assert.NoError(t, reg.Register("alpha", func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	}))
	assert.NoError(t, reg.RegisterSchema("alpha", func() (map[string]any, error) {
		return map[string]any{"type": "object", "title": "alpha"}, nil
	}))

	// beta has a handler but no schema, so it is not creatable.
	assert.NoError(t, reg.Register("beta", func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	}))

	// gamma has a handler and a schema, so it is creatable.
	assert.NoError(t, reg.Register("gamma", func(configBytes []byte) (chat.Client, error) {
		return &mockClient{}, nil
	}))
	assert.NoError(t, reg.RegisterSchema("gamma", func() (map[string]any, error) {
		return map[string]any{"type": "object", "title": "gamma"}, nil
	}))

	// delta has a schema handler that fails, so it is skipped.
	assert.NoError(t, reg.RegisterSchema("delta", func() (map[string]any, error) {
		return nil, errors.New("schema generation failed")
	}))

	providers := reg.Providers()
	assert.Len(t, providers, 2)

	// Providers are sorted by name.
	assert.Equal(t, "alpha", providers[0].Name)
	assert.Equal(t, "gamma", providers[1].Name)

	// Each provider carries its config schema.
	assert.Equal(t, map[string]any{"type": "object", "title": "alpha"}, providers[0].Schema)
	assert.Equal(t, map[string]any{"type": "object", "title": "gamma"}, providers[1].Schema)
}

func TestProvidersEmpty(t *testing.T) {
	reg := NewRegistry()

	providers := reg.Providers()
	assert.Empty(t, providers)
}
