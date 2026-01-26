package types

import "reflect"

type Tools []Tool

type Tool struct {
	Name        string
	Description string
	Func        ToolFunction

	InputType reflect.Type
	Schema    map[string]any
}

type ToolFunction func(input any) (string, error)
