package parser

import (
	"testing"

	"claudeagent/message"
)

func TestParser_ProcessLine_ValidUserMessage(t *testing.T) {
	p := New()

	line := `{"type":"user","message":{"role":"user","content":"Hello"},"parent_tool_use_id":null,"uuid":"abc","session_id":"sess-1"}`

	msgs, err := p.ProcessLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	_, ok := msgs[0].(*message.UserMessage)
	if !ok {
		t.Errorf("expected *UserMessage, got %T", msgs[0])
	}
}

func TestParser_ProcessLine_ValidAssistantMessage(t *testing.T) {
	p := New()

	line := `{"type":"assistant","message":{"id":"msg-1","type":"message","role":"assistant","content":[{"type":"text","text":"Hi!"}],"model":"claude-3"},"parent_tool_use_id":null,"uuid":"def","session_id":"sess-1"}`

	msgs, err := p.ProcessLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	asst, ok := msgs[0].(*message.AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msgs[0])
	}

	if len(asst.Message.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(asst.Message.Content))
	}
}

func TestParser_ProcessLine_ValidResultMessage(t *testing.T) {
	p := New()

	line := `{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"result":"Done","total_cost_usd":0.01,"usage":{"input_tokens":10,"output_tokens":20},"modelUsage":{},"permission_denials":[],"uuid":"res-1","session_id":"sess-1"}`

	msgs, err := p.ProcessLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	result, ok := msgs[0].(*message.ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msgs[0])
	}

	if result.Subtype != "success" {
		t.Errorf("expected subtype 'success', got %q", result.Subtype)
	}
}

func TestParser_ProcessLine_EmptyLine(t *testing.T) {
	p := New()

	msgs, err := p.ProcessLine("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msgs != nil {
		t.Errorf("expected nil for empty line, got %v", msgs)
	}
}

func TestParser_ProcessLine_InvalidJSON(t *testing.T) {
	p := New()

	_, err := p.ProcessLine("{invalid json}")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParser_ProcessLine_UnknownType(t *testing.T) {
	p := New()

	msgs, err := p.ProcessLine(`{"type":"unknown_type","session_id":"s","uuid":"u"}`)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	raw, ok := msgs[0].(*message.RawMessage)
	if !ok {
		t.Errorf("expected *message.RawMessage, got %T", msgs[0])
	}
	if raw.Type != "unknown_type" {
		t.Errorf("expected type 'unknown_type', got %q", raw.Type)
	}
}

func TestParser_ProcessLine_MaxBufferSize(t *testing.T) {
	p := New()

	largeContent := make([]byte, 2*1024*1024) // 2MB
	for i := range largeContent {
		largeContent[i] = 'a'
	}

	_, err := p.ProcessLine(string(largeContent))
	if err == nil {
		t.Error("expected error for oversized line")
	}
}

func TestParser_Reset(t *testing.T) {
	p := New()
	p.Reset()
}
