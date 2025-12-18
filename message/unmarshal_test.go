package message

import (
	"testing"
)

func TestParseMessage_User(t *testing.T) {
	data := []byte(`{
		"type": "user",
		"message": {"role": "user", "content": "Hello"},
		"parent_tool_use_id": null,
		"uuid": "abc-123",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.MessageType() != "user" {
		t.Errorf("expected type 'user', got %q", user.MessageType())
	}
	if user.GetUUID() != "abc-123" {
		t.Errorf("expected uuid 'abc-123', got %q", user.GetUUID())
	}
	if user.GetSessionID() != "session-1" {
		t.Errorf("expected session_id 'session-1', got %q", user.GetSessionID())
	}
}

func TestParseMessage_Assistant(t *testing.T) {
	data := []byte(`{
		"type": "assistant",
		"message": {
			"id": "msg-1",
			"type": "message",
			"role": "assistant",
			"content": [
				{"type": "text", "text": "Hello!"},
				{"type": "thinking", "thinking": "Let me think...", "signature": "sig123"}
			],
			"model": "claude-3",
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 10, "output_tokens": 20}
		},
		"parent_tool_use_id": null,
		"uuid": "def-456",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	asst, ok := msg.(*AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}

	if asst.MessageType() != "assistant" {
		t.Errorf("expected type 'assistant', got %q", asst.MessageType())
	}
	if len(asst.Message.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(asst.Message.Content))
	}

	textBlock, ok := asst.Message.Content[0].(*TextBlock)
	if !ok {
		t.Fatalf("expected *TextBlock, got %T", asst.Message.Content[0])
	}
	if textBlock.Text != "Hello!" {
		t.Errorf("expected text 'Hello!', got %q", textBlock.Text)
	}

	thinkingBlock, ok := asst.Message.Content[1].(*ThinkingBlock)
	if !ok {
		t.Fatalf("expected *ThinkingBlock, got %T", asst.Message.Content[1])
	}
	if thinkingBlock.Thinking != "Let me think..." {
		t.Errorf("expected thinking 'Let me think...', got %q", thinkingBlock.Thinking)
	}
}

func TestParseMessage_Result(t *testing.T) {
	data := []byte(`{
		"type": "result",
		"subtype": "success",
		"duration_ms": 1000,
		"duration_api_ms": 800,
		"is_error": false,
		"num_turns": 1,
		"result": "Done",
		"total_cost_usd": 0.001,
		"usage": {"input_tokens": 100, "output_tokens": 50},
		"modelUsage": {},
		"permission_denials": [],
		"uuid": "res-789",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, ok := msg.(*ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msg)
	}

	if result.Subtype != "success" {
		t.Errorf("expected subtype 'success', got %q", result.Subtype)
	}
	if result.DurationMS != 1000 {
		t.Errorf("expected duration_ms 1000, got %d", result.DurationMS)
	}
	if result.TotalCostUSD != 0.001 {
		t.Errorf("expected total_cost_usd 0.001, got %f", result.TotalCostUSD)
	}
}

func TestParseMessage_System(t *testing.T) {
	data := []byte(`{
		"type": "system",
		"subtype": "init",
		"tools": ["Bash", "Read", "Write"],
		"model": "claude-3",
		"cwd": "/home/user",
		"uuid": "sys-123",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sys, ok := msg.(*SystemMessage)
	if !ok {
		t.Fatalf("expected *SystemMessage, got %T", msg)
	}

	if sys.Subtype != "init" {
		t.Errorf("expected subtype 'init', got %q", sys.Subtype)
	}
	if len(sys.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(sys.Tools))
	}
}

func TestParseMessage_StreamEvent(t *testing.T) {
	data := []byte(`{
		"type": "stream_event",
		"event": {"type": "content_block_delta"},
		"parent_tool_use_id": null,
		"uuid": "stream-1",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := msg.(*StreamEvent)
	if !ok {
		t.Fatalf("expected *StreamEvent, got %T", msg)
	}
}

func TestParseMessage_ToolProgress(t *testing.T) {
	data := []byte(`{
		"type": "tool_progress",
		"tool_use_id": "tool-1",
		"tool_name": "Bash",
		"parent_tool_use_id": null,
		"elapsed_time_seconds": 5.5,
		"uuid": "prog-1",
		"session_id": "session-1"
	}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prog, ok := msg.(*ToolProgressMessage)
	if !ok {
		t.Fatalf("expected *ToolProgressMessage, got %T", msg)
	}

	if prog.ToolName != "Bash" {
		t.Errorf("expected tool_name 'Bash', got %q", prog.ToolName)
	}
	if prog.ElapsedTimeSeconds != 5.5 {
		t.Errorf("expected elapsed_time_seconds 5.5, got %f", prog.ElapsedTimeSeconds)
	}
}

func TestParseMessage_UnknownType(t *testing.T) {
	data := []byte(`{"type": "unknown_type", "foo": "bar", "session_id": "test-session", "uuid": "test-uuid"}`)

	msg, err := ParseMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, ok := msg.(*RawMessage)
	if !ok {
		t.Fatalf("expected *RawMessage, got %T", msg)
	}
	if raw.Type != "unknown_type" {
		t.Errorf("expected type 'unknown_type', got %q", raw.Type)
	}
	if raw.GetSessionID() != "test-session" {
		t.Errorf("expected session_id 'test-session', got %q", raw.GetSessionID())
	}
	if raw.GetUUID() != "test-uuid" {
		t.Errorf("expected uuid 'test-uuid', got %q", raw.GetUUID())
	}
	if raw.Data["foo"] != "bar" {
		t.Errorf("expected data['foo'] = 'bar', got %v", raw.Data["foo"])
	}
}

func TestParseMessage_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid json}`)

	_, err := ParseMessage(data)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
