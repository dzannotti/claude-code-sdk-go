package message

import (
	"encoding/json"
	"fmt"
)

type ContentBlock interface {
	BlockType() string
}

type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (b *TextBlock) BlockType() string { return "text" }

type ThinkingBlock struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

func (b *ThinkingBlock) BlockType() string { return "thinking" }

type ToolUseBlock struct {
	Type  string         `json:"type"`
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (b *ToolUseBlock) BlockType() string { return "tool_use" }

type ToolResultBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content"`
	IsError   *bool  `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() string { return "tool_result" }

type RedactedThinkingBlock struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func (b *RedactedThinkingBlock) BlockType() string { return "redacted_thinking" }

type RawContentBlock struct {
	data map[string]any
}

func (b *RawContentBlock) BlockType() string {
	if t, ok := b.data["type"].(string); ok {
		return t
	}
	return "unknown"
}

func (b *RawContentBlock) Data() map[string]any {
	return b.data
}

func ParseContentBlock(data json.RawMessage) (ContentBlock, error) {
	var typeHolder struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeHolder); err != nil {
		return nil, fmt.Errorf("failed to determine content block type: %w", err)
	}

	switch typeHolder.Type {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse text block: %w", err)
		}
		return &block, nil

	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse thinking block: %w", err)
		}
		return &block, nil

	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_use block: %w", err)
		}
		return &block, nil

	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_result block: %w", err)
		}
		return &block, nil

	case "redacted_thinking":
		var block RedactedThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse redacted_thinking block: %w", err)
		}
		return &block, nil

	default:
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse raw content block: %w", err)
		}
		return &RawContentBlock{data: raw}, nil
	}
}
