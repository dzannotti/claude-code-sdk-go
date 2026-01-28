package session

import (
	"os"
	"path/filepath"
	"testing"

	"claudeagent/message"
)

func TestLoad(t *testing.T) {
	// Create temp session file matching real CLI format
	dir := t.TempDir()
	sessionFile := filepath.Join(dir, "test-session.jsonl")

	// Note: Result messages are NOT stored in session files - only streamed
	content := `{"type":"queue-operation","operation":"dequeue","timestamp":"2025-12-17T20:14:55.681Z","sessionId":"test"}
{"type":"user","message":{"role":"user","content":"Hello"},"uuid":"uuid-1","sessionId":"test"}
{"type":"assistant","message":{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"text","text":"Hi there!"}],"model":"claude-3"},"uuid":"uuid-2","sessionId":"test"}
`
	if err := os.WriteFile(sessionFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	msgs, err := Load(sessionFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Should have 2 messages (queue-operation is skipped, results not stored)
	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(msgs))
	}

	// Check user message
	userMsg, ok := msgs[0].(*message.UserMessage)
	if !ok {
		t.Fatalf("Expected UserMessage, got %T", msgs[0])
	}
	if userMsg.UUID != "uuid-1" {
		t.Errorf("Expected UUID 'uuid-1', got %q", userMsg.UUID)
	}

	// Check assistant message
	assistantMsg, ok := msgs[1].(*message.AssistantMessage)
	if !ok {
		t.Fatalf("Expected AssistantMessage, got %T", msgs[1])
	}
	if assistantMsg.UUID != "uuid-2" {
		t.Errorf("Expected UUID 'uuid-2', got %q", assistantMsg.UUID)
	}
	if len(assistantMsg.Message.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(assistantMsg.Message.Content))
	}
}

func TestListSessions(t *testing.T) {
	dir := t.TempDir()

	// Create some session files
	files := []string{
		"session-1.jsonl",
		"session-2.jsonl",
		"agent-abc123.jsonl", // Should be skipped
		"not-jsonl.txt",      // Should be skipped
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("{}"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	sessions, err := ListSessions(dir)
	if err != nil {
		t.Fatalf("ListSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}

	// Check IDs (should not have .jsonl suffix)
	ids := make(map[string]bool)
	for _, s := range sessions {
		ids[s.ID] = true
	}
	if !ids["session-1"] || !ids["session-2"] {
		t.Errorf("Expected sessions 'session-1' and 'session-2', got %v", ids)
	}
}

func TestLoadByID(t *testing.T) {
	dir := t.TempDir()
	sessionID := "test-session-123"
	sessionFile := filepath.Join(dir, sessionID+".jsonl")

	content := `{"type":"user","message":{"role":"user","content":"Test"},"uuid":"u1","sessionId":"test"}`
	if err := os.WriteFile(sessionFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	msgs, err := LoadByID(dir, sessionID)
	if err != nil {
		t.Fatalf("LoadByID failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(msgs))
	}
}

func TestLoadEmptyFile(t *testing.T) {
	dir := t.TempDir()
	sessionFile := filepath.Join(dir, "empty.jsonl")

	if err := os.WriteFile(sessionFile, []byte(""), 0600); err != nil {
		t.Fatal(err)
	}

	msgs, err := Load(sessionFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(msgs) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(msgs))
	}
}

func TestLoadNonexistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/session.jsonl")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
