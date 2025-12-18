# Claude Code SDK for Go

Go SDK for programmatic control of [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code).

## Requirements

- Go 1.21+
- Claude Code CLI installed and authenticated (`claude` command available)

## Installation

```bash
go get github.com/dzannotti/claude-code-sdk-go
```

## Quick Start

### Query API (One-shot)

For simple, one-off queries:

```go
package main

import (
    "context"
    "errors"
    "fmt"
    "log"

    claudecode "github.com/dzannotti/claude-code-sdk-go"
)

func main() {
    ctx := context.Background()

    iter, err := claudecode.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatal(err)
    }
    defer iter.Close()

    for {
        msg, err := iter.Next(ctx)
        if errors.Is(err, claudecode.ErrDone) {
            break
        }
        if err != nil {
            log.Fatal(err)
        }

        if assistant, ok := msg.(*claudecode.AssistantMessage); ok {
            for _, block := range assistant.Message.Content {
                if text, ok := block.(*claudecode.TextBlock); ok {
                    fmt.Print(text.Text)
                }
            }
        }
    }
}
```

### Client API (Interactive)

For multi-turn conversations and streaming:

```go
package main

import (
    "context"
    "fmt"
    "log"

    claudecode "github.com/dzannotti/claude-code-sdk-go"
)

func main() {
    ctx := context.Background()

    err := claudecode.WithClient(ctx, func(c claudecode.Client) error {
        // First query
        if err := c.Query(ctx, "Remember the number 42"); err != nil {
            return err
        }
        for msg := range c.Messages(ctx) {
            handleMessage(msg)
        }

        // Follow-up (conversation continues)
        if err := c.Query(ctx, "What number did I ask you to remember?"); err != nil {
            return err
        }
        for msg := range c.Messages(ctx) {
            handleMessage(msg)
        }

        return nil
    })

    if err != nil {
        log.Fatal(err)
    }
}

func handleMessage(msg claudecode.Message) {
    switch m := msg.(type) {
    case *claudecode.AssistantMessage:
        for _, block := range m.Message.Content {
            if text, ok := block.(*claudecode.TextBlock); ok {
                fmt.Print(text.Text)
            }
        }
    }
}
```

## Options

```go
// Query with options
iter, err := claudecode.Query(ctx, "Hello",
    claudecode.WithModel("claude-sonnet-4-20250514"),
    claudecode.WithMaxTurns(3),
    claudecode.WithSystemPrompt("You are a helpful assistant"),
    claudecode.WithWorkingDirectory("/path/to/project"),
    claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits),
    claudecode.WithIncludePartialMessages(), // Enable token streaming
)

// Client with options
err := claudecode.WithClient(ctx, func(c claudecode.Client) error {
    // ...
},
    claudecode.WithModel("claude-sonnet-4-20250514"),
    claudecode.WithPermissionMode(claudecode.PermissionModeAcceptEdits),
)
```

### Available Options

| Option | Description |
|--------|-------------|
| `WithModel(model)` | Specify Claude model to use |
| `WithMaxTurns(n)` | Limit agentic turns |
| `WithSystemPrompt(prompt)` | Custom system prompt |
| `WithAppendSystemPrompt(prompt)` | Append to default system prompt |
| `WithWorkingDirectory(path)` | Set working directory for tools |
| `WithPermissionMode(mode)` | Control tool permissions |
| `WithAllowedTools(tools...)` | Restrict available tools |
| `WithDisallowedTools(tools...)` | Block specific tools |
| `WithMcpServers(config)` | Configure MCP servers |
| `WithIncludePartialMessages()` | Enable token-by-token streaming |

## Message Types

The SDK provides typed messages from the CLI:

- `*AssistantMessage` - Claude's responses with content blocks
- `*UserMessage` - User inputs (for context in multi-turn)
- `*ResultMessage` - Query completion with stats
- `*SystemMessage` - System events and tool outputs
- `*StreamEvent` - Token streaming events (with `WithIncludePartialMessages()`)
- `*ToolProgressMessage` - Long-running tool progress

### Content Blocks

Assistant messages contain content blocks:

- `*TextBlock` - Text content
- `*ToolUseBlock` - Tool invocations
- `*ToolResultBlock` - Tool results
- `*ThinkingBlock` - Extended thinking content

## Control Protocol

The Client API supports bidirectional control:

```go
err := claudecode.WithClient(ctx, func(c claudecode.Client) error {
    // Change model mid-conversation
    model := "claude-sonnet-4-20250514"
    c.SetModel(ctx, &model)

    // Interrupt current operation
    c.Interrupt(ctx)

    // Get MCP server status
    statuses, _ := c.McpStatus(ctx)

    return nil
})
```

## Examples

See the [`examples/`](./examples) directory:

- `01_quickstart/` - Basic Query API usage
- `02_client_streaming/` - Streaming with Client API
- `03_client_multi_turn/` - Multi-turn conversations
- `08_client_advanced/` - Advanced client features
- `10_context_manager/` - Context and cancellation
- `interactive_chat/` - Full interactive chat with tools
- `slash_commands/` - Discover custom slash commands

Run an example:

```bash
go run ./examples/01_quickstart
```

## License

MIT
