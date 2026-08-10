package provider

import (
	"context"
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
