package transport

import (
	"context"

	"claudeagent/message"
)

type StreamMessage struct {
	Type            string              `json:"type"`
	Message         message.UserContent `json:"message"`
	ParentToolUseID *string             `json:"parent_tool_use_id"`
	SessionID       string              `json:"session_id"`
}

type Transport interface {
	Connect(ctx context.Context) error
	SendMessage(ctx context.Context, msg StreamMessage) error
	ReceiveMessages(ctx context.Context) (<-chan message.Message, <-chan error)
	Interrupt(ctx context.Context) error
	Close() error
	IsConnected() bool
}
