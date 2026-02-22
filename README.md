# interlude

Go library for streaming LLM interactions with tool calling support.

Handles streaming, multi-step tool calls, and user approval flows — so you don't have to.

## Install

```bash
go get github.com/x2d7/interlude
```

## Quick Start

See [examples/](examples/) for usage examples.

```go
client := openai.OpenAIClient{
    Endpoint: "https://openrouter.ai/api/v1",
    APIKey:   os.Getenv("API_KEY"),
    Model:    "gpt-4o",
}

c := chat.Chat{
    Messages: chat.NewMessages(),
}

for event := range c.SendUserStream(ctx, &client, "Hello!") {
    switch v := event.(type) {
    case chat.EventNewToken:
        fmt.Print(v.Content)
    case chat.EventNewError:
        log.Fatal(v.Error)
    }
}
```

## Tool Calling

```go
toolList := tools.NewTools()

tool, _ := tools.NewTool("get_weather", "Returns weather for a city", func(input struct {
    City string `json:"city"`
}) (string, error) {
    return "Sunny, 22°C", nil
})

toolList.Add(tool)

c := chat.Chat{
    Messages: chat.NewMessages(),
    Tools:    toolList,
}

for event := range c.SendUserStream(ctx, &client, "What's the weather in Berlin?") {
    switch v := event.(type) {
    case chat.EventNewToken:
        fmt.Print(v.Content)
    case chat.EventNewToolCall:
        v.Resolve(true) // approve the call
    case chat.EventCompletionEnded:
        fmt.Println()
    }
}
```

Tool input schema is generated automatically from your struct using `jsonschema` tags.

## Providers

- OpenAI-compatible APIs (OpenAI, OpenRouter, etc.)

## License

MIT
