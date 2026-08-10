// simple-chat demonstrates a basic multi-turn conversation using provider-agnostic config.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/x2d7/interlude/chat"
	_ "github.com/x2d7/interlude/connect/openai/config"
	"github.com/x2d7/interlude/provider"
)

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	blue   = "\033[34m"
)

func colorize(color, s string) string { return color + s + reset }

func main() {
	// normal way: create config directly
	// (!add import openai_config "github.com/x2d7/interlude/connect/openai/config")
	// client := &openai_config.ProviderConfig{
	// 	Conn: openai_config.Connection{
	// 		Endpoint: "http://localhost:9000/v1/",
	// 		APIKey:   "sk-...",
	// 		Model:    "current",
	// 	},
	// }.ToClient()

	// provider-agnostic way: from JSON string
	configJSON := `{
		"provider": "openai",
		"config": {
			"conn": {
				"endpoint": "http://localhost:9000/v1/",
				"api_key":  "sk-...",
				"model":    "current"
			}
		}
	}`

	client, err := provider.Deserialize([]byte(configJSON), provider.DefaultRegistry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "deserialize: %v\n", err)
		os.Exit(1)
	}

	c := chat.Chat{
		Messages: chat.NewMessages(),
	}

	fmt.Println(colorize(dim, "Type your message and press Enter. Ctrl+C or Ctrl+D to exit."))
	fmt.Println(colorize(dim, strings.Repeat("─", 40)))
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	ctx := context.Background()

	for {
		fmt.Print(colorize(bold+blue, "You: "))
		if !scanner.Scan() {
			fmt.Println()
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		fmt.Println()
		fmt.Print(colorize(bold+green, "Assistant: "))

		for event := range c.SendUserStream(ctx, client, input) {
			switch v := event.(type) {

			case chat.EventToken:
				fmt.Print(v.Content)

			case chat.EventRefusal:
				// Provider-sensitive event. Some providers may not support it.
				fmt.Print(colorize(yellow, v.Content))

			case chat.EventToolCall:
				fmt.Printf(colorize(dim, "\n[tool call] %s(%s)\n"), v.Name, v.Content)

			case chat.EventCompletionEnded:
				// nothing to do, stream will close

			case chat.EventError:
				fmt.Fprintf(os.Stderr, "\n%s %s\n", colorize(bold+red, "error:"), v.Error)
				os.Exit(1)
			}
		}

		fmt.Printf("\n%s\n\n", colorize(dim, strings.Repeat("─", 40)))
	}
}
