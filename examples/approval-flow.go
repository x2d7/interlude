// approval-flow demonstrates user-controlled tool call approval.
// The assistant wants to execute tools — you decide whether to allow each one.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/chat/tools"
	openai "github.com/x2d7/interlude/connect/openai"
)

type ReadFileInput struct {
	Path string `json:"path" jsonschema:"description=Path to the file to read"`
}

type WriteFileInput struct {
	Path    string `json:"path"    jsonschema:"description=Path to write the file"`
	Content string `json:"content" jsonschema:"description=Content to write"`
}

// Some models underuse tools without a concrete prior example to anchor on.
func seedToolUsageExample(c *chat.Chat) {
	c.Messages.AddEvent(chat.NewEventNewSystemMessage(
		`This is a tryout. Open the file "input.txt" and tell me what's inside.`,
	))

	exampleCall := chat.NewEventNewToolCall(
		"example-seed-call-1",
		"read_file",
		`{"path": "input.txt"}`,
	)
	c.Messages.AddEvent(exampleCall)

	c.Messages.AddEvent(chat.NewEventNewToolMessage(
		"example-seed-call-1",
		"input.txt is not real ;(",
		true,
	))

	c.Messages.AddEvent(chat.NewEventNewSystemMessage(
		`The was just an example on how to use tools. input.txt might be empty, don't worry. Try it yourself!`,
	))
}

func main() {
	_ = godotenv.Load()

	client := openai.OpenAIClient{
		Endpoint: os.Getenv("OPENROUTER_BASEURL"),
		APIKey:   os.Getenv("OPENROUTER_TOKEN"),
		Model:    os.Getenv("OPENROUTER_MODEL"),
	}

	toolList := tools.NewTools()

	readTool, _ := tools.NewTool("read_file", "Reads a file from disk", func(input ReadFileInput) (string, error) {
		data, err := os.ReadFile(input.Path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})

	writeTool, _ := tools.NewTool("write_file", "Writes content to a file on disk", func(input WriteFileInput) (string, error) {
		err := os.WriteFile(input.Path, []byte(input.Content), 0644)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("written %d bytes to %s", len(input.Content), input.Path), nil
	})

	toolList.Add(readTool)
	toolList.Add(writeTool)

	c := chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &toolList,
	}

	// Prime the model with a concrete tool usage example before the real prompt
	seedToolUsageExample(&c)

	prompt := `Read the file "input.txt", then write its contents in uppercase to "output.txt".`
	fmt.Printf("Prompt: %s\n\n", prompt)

	pendingCalls := make([]chat.EventNewToolCall, 0)

	for event := range c.SendUserStream(context.Background(), &client, prompt) {
		switch v := event.(type) {
		case chat.EventNewToken:
			fmt.Print(v.Content)

		case chat.EventNewToolCall:
			pendingCalls = append(pendingCalls, v)

		case chat.EventCompletionEnded:
			fmt.Println()
			for _, call := range pendingCalls {
				fmt.Print("\033[31m")
				fmt.Printf("\n Tool: %s\n", call.Name)
				fmt.Printf("   Args: %s\n", call.Content)
				fmt.Print("   Allow? (y/n): ")
				fmt.Print("\033[0m")

				var answer string
				fmt.Scanln(&answer)
				call.Resolve(answer == "y")
			}
			pendingCalls = pendingCalls[:0]

		case chat.EventNewError:
			fmt.Fprintf(os.Stderr, "\nerror: %s\n", v.Error)
			os.Exit(1)
		}
	}

	fmt.Println()
}
