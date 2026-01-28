package claudeagent

import (
	"errors"
	"fmt"
)

var (
	ErrDone          = errors.New("no more messages")
	ErrNotConnected  = errors.New("client not connected")
	ErrAlreadyClosed = errors.New("client already closed")
	ErrAborted       = errors.New("operation aborted")
)

type AbortError struct {
	Message string
}

func (e *AbortError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("aborted: %s", e.Message)
	}
	return "aborted"
}

func (e *AbortError) Is(target error) bool {
	return target == ErrAborted
}

type CLINotFoundError struct {
	SearchedPaths []string
}

func (e *CLINotFoundError) Error() string {
	return fmt.Sprintf("claude CLI not found in paths: %v", e.SearchedPaths)
}

type ConnectionError struct {
	Message string
	Cause   error
}

func (e *ConnectionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("connection error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("connection error: %s", e.Message)
}

func (e *ConnectionError) Unwrap() error {
	return e.Cause
}

type ProcessError struct {
	Message  string
	ExitCode int
	Stderr   string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("process error (exit %d): %s", e.ExitCode, e.Message)
}

type JSONDecodeError struct {
	Line  string
	Cause error
}

func (e *JSONDecodeError) Error() string {
	return fmt.Sprintf("JSON decode error: %v (line: %s)", e.Cause, truncate(e.Line, 100))
}

func (e *JSONDecodeError) Unwrap() error {
	return e.Cause
}

type MessageParseError struct {
	Message string
	Type    string
	Data    []byte
}

func (e *MessageParseError) Error() string {
	return fmt.Sprintf("failed to parse message type %q: %s", e.Type, e.Message)
}

type ControlError struct {
	RequestID string
	Message   string
}

func (e *ControlError) Error() string {
	return fmt.Sprintf("control error (request %s): %s", e.RequestID, e.Message)
}

type TimeoutError struct {
	Operation string
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout: %s", e.Operation)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
