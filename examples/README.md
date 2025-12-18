# Claude Code SDK for Go - Examples

Working examples demonstrating the Claude Code SDK for Go.

## Prerequisites

- Go 1.21+
- Claude Code CLI: `npm install -g @anthropic-ai/claude-code`

## Examples

### 01_quickstart - Basic Query API
Simple one-shot query demonstrating the Query API.

```bash
cd examples/01_quickstart
go run main.go
```

### 02_client_streaming - Real-Time Streaming
Stream responses in real-time using the Client API with WithClient.

```bash
cd examples/02_client_streaming
go run main.go
```

### 03_client_multi_turn - Conversations
Multi-turn conversations with context preservation.

```bash
cd examples/03_client_multi_turn
go run main.go
```

### 08_client_advanced - Production Patterns
Advanced error handling, custom options, and production patterns.

```bash
cd examples/08_client_advanced
go run main.go
```

### 10_context_manager - WithClient Pattern
Comparison of WithClient (recommended) vs manual connection management.

```bash
cd examples/10_context_manager
go run main.go
```

## API Patterns

### Query API - One-Shot Operations
```go
iterator, err := claudecode.Query(ctx, "What is 2+2?")
if err != nil {
    log.Fatal(err)
}
defer iterator.Close()

for {
    msg, err := iterator.Next(ctx)
    if errors.Is(err, claudecode.ErrDone) {
        break
    }
    // Handle message...
}
```

### Client API - WithClient (Recommended)
```go
err := claudecode.WithClient(ctx, func(client claudecode.Client) error {
    if err := client.Query(ctx, "Hello"); err != nil {
        return err
    }

    for msg := range client.Messages(ctx) {
        // Handle message...
    }
    return nil
})
```

### Client API - Manual Pattern
```go
client, err := claudecode.NewClient()
if err != nil {
    log.Fatal(err)
}

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

client.Query(ctx, "Hello")
for msg := range client.Messages(ctx) {
    // Handle message...
}
```

## When to Use Which API

### Query API
- One-shot questions
- Batch processing
- CI/CD scripts
- Simple automation

### Client API
- Multi-turn conversations
- Interactive applications
- Context-dependent workflows
- Real-time streaming
