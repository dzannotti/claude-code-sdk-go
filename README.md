# Claude Code SDK for Go

Go SDK for programmatic control of [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code).

> **Feature parity** with [`@anthropic-ai/claude-agent-sdk`](https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk) v0.2.22 (all stable APIs). Unstable/preview APIs (`unstable_v2_*`) are not ported.

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
    claudecode.WithCwd("/path/to/project"),
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
| `WithFallbackModel(model)` | Fallback model if primary unavailable |
| `WithMaxTurns(n)` | Limit agentic turns |
| `WithMaxBudgetUSD(budget)` | Maximum cost budget |
| `WithMaxThinkingTokens(n)` | Max tokens for extended thinking |
| `WithSystemPrompt(prompt)` | Custom system prompt |
| `WithSystemPromptPreset(preset, append)` | Use a system prompt preset |
| `WithCwd(path)` | Set working directory |
| `WithAdditionalDirectories(dirs...)` | Add extra directories |
| `WithPermissionMode(mode)` | Control tool permissions |
| `WithCanUseTool(fn)` | Custom permission callback |
| `WithAllowedTools(tools...)` | Restrict available tools |
| `WithDisallowedTools(tools...)` | Block specific tools |
| `WithMcpServers(config)` | Configure MCP servers |
| `WithHooks(event, matchers...)` | Register lifecycle hooks |
| `WithIncludePartialMessages()` | Enable token-by-token streaming |
| `WithOutputFormat(format)` | Structured JSON output |
| `WithSandbox(settings)` | Sandbox configuration |
| `WithExecutable(exe, args...)` | Node runtime to use (bun, deno, node) |
| `WithEnv(env)` | Environment variables |
| `WithStderr(fn)` | Stderr callback |
| `WithResume(sessionID)` | Resume a session |
| `WithContinue()` | Continue last conversation |
| `WithForkSession()` | Fork an existing session |
| `WithBetas(betas...)` | Enable beta features |
| `WithAgents(agents)` | Define sub-agents |

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
    c.SetModel(ctx, "claude-sonnet-4-20250514")

    // Interrupt current operation
    c.Interrupt(ctx)

    // Get MCP server status
    statuses, _ := c.McpServerStatus(ctx)

    return nil
})
```

## Session History

Load conversation history from previous sessions:

```go
import "claudecode/session"

// Find project directory for current working directory
projectDir, _ := session.ProjectDir(".")

// List available sessions
sessions, _ := session.ListSessions(projectDir)
for _, s := range sessions {
    fmt.Printf("%s - %s\n", s.ID, s.ModTime)
}

// Load messages from a session
msgs, _ := session.LoadByID(projectDir, sessionID)
for _, msg := range msgs {
    // Display previous messages in your UI
}

// Resume the session
iter, _ := claudecode.Query(ctx, "continue",
    claudecode.WithResume(sessionID),
)
```

## Examples

See the [`examples/`](./examples) directory:

- `01_quickstart/` - Basic Query API usage
- `02_client_streaming/` - Streaming with Client API
- `03_client_multi_turn/` - Multi-turn conversations
- `04_session_resume/` - Load history and resume sessions
- `05_hooks/` - Lifecycle hooks
- `06_mcp_servers/` - MCP server configuration
- `07_permissions/` - Custom permission handling
- `08_client_advanced/` - Advanced client features
- `09_output_format/` - Structured JSON output
- `10_context_manager/` - Context and cancellation
- `interactive_chat/` - Full interactive chat with tools
- `slash_commands/` - Discover custom slash commands

Run an example:

```bash
go run ./examples/01_quickstart
```

## License

MIT
