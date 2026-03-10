package chat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventToolCall_Resolve_ReturnsErrorOnDuplicate(t *testing.T) {
	tc := NewEventToolCall("call-1", "test-tool", `{}`)

	err := tc.Resolve(true)
	assert.NoError(t, err)

	err = tc.Resolve(true)
	assert.ErrorIs(t, err, ErrAlreadyResolved)
}

func TestEventToolCall_Resolve_CopiesShareResolution(t *testing.T) {
	tc := NewEventToolCall("call-1", "test-tool", `{}`)
	copy1 := tc
	copy2 := tc

	err := copy1.Resolve(true)
	assert.NoError(t, err)

	err = copy2.Resolve(true)
	assert.ErrorIs(t, err, ErrAlreadyResolved)
}
