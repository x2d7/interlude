// simple-chat demonstrates a basic multi-turn conversation.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/x2d7/interlude/chat"
	"github.com/x2d7/interlude/chat/tools"
	openai "github.com/x2d7/interlude/connect/openai"
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
	_ = godotenv.Load()

	client := openai.OpenAIClient{
		Endpoint: os.Getenv("OPENROUTER_BASEURL"),
		APIKey:   os.Getenv("OPENROUTER_TOKEN"),
		Model:    os.Getenv("OPENROUTER_MODEL"),
	}

	t := tools.NewTools()
	c := chat.Chat{
		Messages: chat.NewMessages(),
		Tools:    &t,
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

		for event := range c.SendUserStream(ctx, &client, input) {
			switch v := event.(type) {

			case chat.EventNewToken:
				fmt.Print(v.Content)

			case chat.EventNewRefusal:
				// Provider-sensitive event. Some providers may not support it.
				fmt.Print(colorize(yellow, v.Content))

			case chat.EventNewToolCall:
				fmt.Printf(colorize(dim, "\n[tool call] %s(%s)\n"), v.Name, v.Content)

			case chat.EventCompletionEnded:
				// nothing to do, stream will close

			case chat.EventNewError:
				fmt.Fprintf(os.Stderr, "\n%s %s\n", colorize(bold+red, "error:"), v.Error)
				os.Exit(1)
			}
		}

		fmt.Printf("\n%s\n\n", colorize(dim, strings.Repeat("─", 40)))
	}
}
