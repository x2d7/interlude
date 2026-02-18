package tools

import "errors"

var (
	ErrEmptyToolID      = errors.New("tool ID cannot be empty")
	ErrToolAlreadyExists = errors.New("tool with this ID already exists")
)