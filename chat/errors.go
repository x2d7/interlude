package chat

import "errors"

var (
	ErrUnsupportedSender        = errors.New("unsupported sender type")
	ErrNilStreaming             = errors.New("streaming object is nil")
	ErrAssistantMessageNotFound = errors.New("assistant message not found")
)
