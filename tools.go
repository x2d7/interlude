package interlude

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
)

type Tool struct {
	Name        string
	Description string
	Func        func(input any) (string, error)

	InputType reflect.Type
	Schema    map[string]any
}

type Tools []Tool

func NewTools() Tools { return Tools{} }

func (t *Tools) Add(tool Tool) { *t = append(*t, tool) }

func NewTool[T any](name, description string, f func(T) (string, error)) Tool {
	inputType := reflect.TypeOf((*T)(nil)).Elem()

	// генерируем схему
	ptr := reflect.New(inputType).Interface()
	s := jsonschema.Reflect(ptr)
	b, _ := json.Marshal(s)
	var schemaMap map[string]any
	_ = json.Unmarshal(b, &schemaMap)

	wrapper := func(input any) (string, error) {
		// прямое соответствие типу T
		if v, ok := input.(T); ok {
			return f(v)
		}
		// указатель на T
		if pv, ok := input.(*T); ok && pv != nil {
			return f(*pv)
		}

		var parsed T

		switch v := input.(type) {
		case string:
			// json-object
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				return f(parsed)
			}
			// quoted string
			if s, err := strconv.Unquote(v); err == nil {
				if err := json.Unmarshal([]byte(s), &parsed); err == nil {
					return f(parsed)
				}
			}
			return "", fmt.Errorf("cannot unmarshal string into %T", parsed)

		case []byte:
			if err := json.Unmarshal(v, &parsed); err == nil {
				return f(parsed)
			}
			return "", fmt.Errorf("cannot unmarshal bytes into %T", parsed)

		default:
			b, err := json.Marshal(input)
			if err != nil {
				return "", fmt.Errorf("marshal input: %w", err)
			}
			if err := json.Unmarshal(b, &parsed); err != nil {
				return "", fmt.Errorf("unmarshal to %T: %w", parsed, err)
			}
			return f(parsed)
		}
	}

	return Tool{
		Name:        name,
		Description: description,
		Func:        wrapper,
		InputType:   inputType,
		Schema:      schemaMap,
	}
}

func ConvertTools(t Tools) []openai.ChatCompletionToolUnionParam {
	out := make([]openai.ChatCompletionToolUnionParam, 0, len(t))
	for _, tool := range t {
		if tool.Schema == nil {
			out = append(out, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
				Name:        tool.Name,
				Description: openai.String(tool.Description),
			}))
			continue
		}
		out = append(out, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name,
			Description: openai.String(tool.Description),
			Parameters:  openai.FunctionParameters(tool.Schema),
		}))
	}
	return out
}

func (t *Tools) Execute(toolCall openai.ChatCompletionChunkChoiceDeltaToolCall) (result string, ok bool) {
	for _, tool := range *t {
		if tool.Name == toolCall.Function.Name {
			raw := toolCall.Function.JSON.Arguments.Raw()
			result, err := tool.Func(raw)
			id := toolCall.ID

			ok := true
			if err != nil {
				result = err.Error()
				ok = false
			}

			return fmt.Sprintf("%v>%v", id, result), ok
		}
	}
	return fmt.Sprintf("%v>error: tool %q not found", toolCall.ID, toolCall.Function.Name), false
}
