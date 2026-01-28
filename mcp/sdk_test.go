package mcp

import (
	"context"
	"testing"
)

type EchoInput struct {
	Message string `json:"message" description:"The message to echo"`
}

func TestTool(t *testing.T) {
	tool := Tool("echo", "Echoes back the message", func(ctx context.Context, args EchoInput) (*ToolResult, error) {
		return &ToolResult{
			Content: []ToolResultContent{{Type: "text", Text: args.Message}},
		}, nil
	})

	if tool.Definition.Name != "echo" {
		t.Errorf("expected name 'echo', got %s", tool.Definition.Name)
	}

	schema := tool.Definition.InputSchema
	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties in schema")
	}

	msgProp, ok := props["message"].(map[string]any)
	if !ok {
		t.Fatal("expected message property")
	}

	if msgProp["type"] != "string" {
		t.Errorf("expected message type 'string', got %v", msgProp["type"])
	}
}

func TestToolCall(t *testing.T) {
	tool := Tool("echo", "Echoes back the message", func(ctx context.Context, args EchoInput) (*ToolResult, error) {
		return &ToolResult{
			Content: []ToolResultContent{{Type: "text", Text: args.Message}},
		}, nil
	})

	result, err := tool.Call(context.Background(), map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Text != "hello" {
		t.Errorf("expected 'hello', got %s", result.Content[0].Text)
	}
}

func TestSdkMcpServer(t *testing.T) {
	tool := Tool("echo", "Echoes back the message", func(ctx context.Context, args EchoInput) (*ToolResult, error) {
		return &ToolResult{
			Content: []ToolResultContent{{Type: "text", Text: args.Message}},
		}, nil
	})

	server := CreateSdkMcpServer("test-server", WithVersion("2.0.0"), AddTool(tool))

	if server.Name() != "test-server" {
		t.Errorf("expected name 'test-server', got %s", server.Name())
	}

	if server.Version() != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %s", server.Version())
	}

	tools := server.ListTools()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	result, err := server.CallTool(context.Background(), "echo", map[string]any{"message": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content[0].Text != "world" {
		t.Errorf("expected 'world', got %s", result.Content[0].Text)
	}
}

func TestSdkServerConfig(t *testing.T) {
	server := CreateSdkMcpServer("my-server")
	config := server.Config()

	if config.Type != "sdk" {
		t.Errorf("expected type 'sdk', got %s", config.Type)
	}

	if config.Name != "my-server" {
		t.Errorf("expected name 'my-server', got %s", config.Name)
	}

	if config.Instance != server {
		t.Error("expected instance to be the server")
	}
}
