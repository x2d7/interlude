package chat

import "errors"

var (
    ErrUnsupportedSender = errors.New("unsupported sender type")
)
