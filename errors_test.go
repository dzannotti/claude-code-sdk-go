package claudecode

import (
	"errors"
	"testing"
)

func TestCLINotFoundError(t *testing.T) {
	err := &CLINotFoundError{
		SearchedPaths: []string{"/usr/bin/claude", "/usr/local/bin/claude"},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestConnectionError(t *testing.T) {
	cause := errors.New("connection refused")
	err := &ConnectionError{
		Message: "failed to connect",
		Cause:   cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	if !errors.Is(err, cause) {
		t.Error("expected Unwrap to return cause")
	}
}

func TestProcessError(t *testing.T) {
	err := &ProcessError{
		Message:  "process crashed",
		ExitCode: 1,
		Stderr:   "error output",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestJSONDecodeError(t *testing.T) {
	cause := errors.New("unexpected token")
	err := &JSONDecodeError{
		Line:  `{"invalid": json}`,
		Cause: cause,
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}

	if !errors.Is(err, cause) {
		t.Error("expected Unwrap to return cause")
	}
}

func TestMessageParseError(t *testing.T) {
	err := &MessageParseError{
		Message: "unknown field",
		Type:    "assistant",
		Data:    []byte(`{}`),
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestControlError(t *testing.T) {
	err := &ControlError{
		RequestID: "req-123",
		Message:   "timeout",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Operation: "connect",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"hi", 2, "hi"},
		{"", 5, ""},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}
