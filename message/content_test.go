package message

import (
	"encoding/json"
	"testing"
)

func TestParseContentBlock_Text(t *testing.T) {
	data := json.RawMessage(`{"type": "text", "text": "Hello, world!"}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text, ok := block.(*TextBlock)
	if !ok {
		t.Fatalf("expected *TextBlock, got %T", block)
	}

	if text.BlockType() != "text" {
		t.Errorf("expected type 'text', got %q", text.BlockType())
	}
	if text.Text != "Hello, world!" {
		t.Errorf("expected text 'Hello, world!', got %q", text.Text)
	}
}

func TestParseContentBlock_Thinking(t *testing.T) {
	data := json.RawMessage(`{"type": "thinking", "thinking": "Let me analyze this...", "signature": "abc123"}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	thinking, ok := block.(*ThinkingBlock)
	if !ok {
		t.Fatalf("expected *ThinkingBlock, got %T", block)
	}

	if thinking.BlockType() != "thinking" {
		t.Errorf("expected type 'thinking', got %q", thinking.BlockType())
	}
	if thinking.Thinking != "Let me analyze this..." {
		t.Errorf("expected thinking 'Let me analyze this...', got %q", thinking.Thinking)
	}
	if thinking.Signature != "abc123" {
		t.Errorf("expected signature 'abc123', got %q", thinking.Signature)
	}
}

func TestParseContentBlock_ToolUse(t *testing.T) {
	data := json.RawMessage(`{
		"type": "tool_use",
		"id": "tool-123",
		"name": "Bash",
		"input": {"command": "ls -la"}
	}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	toolUse, ok := block.(*ToolUseBlock)
	if !ok {
		t.Fatalf("expected *ToolUseBlock, got %T", block)
	}

	if toolUse.BlockType() != "tool_use" {
		t.Errorf("expected type 'tool_use', got %q", toolUse.BlockType())
	}
	if toolUse.ID != "tool-123" {
		t.Errorf("expected id 'tool-123', got %q", toolUse.ID)
	}
	if toolUse.Name != "Bash" {
		t.Errorf("expected name 'Bash', got %q", toolUse.Name)
	}
	if toolUse.Input["command"] != "ls -la" {
		t.Errorf("expected input.command 'ls -la', got %v", toolUse.Input["command"])
	}
}

func TestParseContentBlock_ToolResult(t *testing.T) {
	isErr := true
	data := json.RawMessage(`{
		"type": "tool_result",
		"tool_use_id": "tool-123",
		"content": "file1.txt\nfile2.txt",
		"is_error": true
	}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	toolResult, ok := block.(*ToolResultBlock)
	if !ok {
		t.Fatalf("expected *ToolResultBlock, got %T", block)
	}

	if toolResult.BlockType() != "tool_result" {
		t.Errorf("expected type 'tool_result', got %q", toolResult.BlockType())
	}
	if toolResult.ToolUseID != "tool-123" {
		t.Errorf("expected tool_use_id 'tool-123', got %q", toolResult.ToolUseID)
	}
	if toolResult.IsError == nil || *toolResult.IsError != isErr {
		t.Errorf("expected is_error true, got %v", toolResult.IsError)
	}
}

func TestParseContentBlock_RedactedThinking(t *testing.T) {
	data := json.RawMessage(`{"type": "redacted_thinking", "data": "base64encodeddata"}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	redacted, ok := block.(*RedactedThinkingBlock)
	if !ok {
		t.Fatalf("expected *RedactedThinkingBlock, got %T", block)
	}

	if redacted.BlockType() != "redacted_thinking" {
		t.Errorf("expected type 'redacted_thinking', got %q", redacted.BlockType())
	}
}

func TestParseContentBlock_Unknown(t *testing.T) {
	data := json.RawMessage(`{"type": "future_block_type", "data": "something"}`)

	block, err := ParseContentBlock(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := block.(*RawContentBlock)
	if !ok {
		t.Fatalf("expected *RawContentBlock, got %T", block)
	}

	if raw.BlockType() != "future_block_type" {
		t.Errorf("expected type 'future_block_type', got %q", raw.BlockType())
	}
}

func TestParseContentBlock_InvalidJSON(t *testing.T) {
	data := json.RawMessage(`{invalid json}`)

	_, err := ParseContentBlock(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
