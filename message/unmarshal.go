package message

import (
	"encoding/json"
	"fmt"
)

func ParseMessage(data []byte) (Message, error) {
	var typeHolder struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeHolder); err != nil {
		return nil, fmt.Errorf("failed to determine message type: %w", err)
	}

	switch typeHolder.Type {
	case "user":
		var msg UserMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse user message: %w", err)
		}
		return &msg, nil

	case "assistant":
		var raw struct {
			Type            string          `json:"type"`
			Message         json.RawMessage `json:"message"`
			ParentToolUseID *string         `json:"parent_tool_use_id"`
			Error           *string         `json:"error,omitempty"`
			UUID            string          `json:"uuid"`
			SessionID       string          `json:"session_id"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse assistant message: %w", err)
		}

		var apiMsg struct {
			ID           string            `json:"id"`
			Type         string            `json:"type"`
			Role         string            `json:"role"`
			Content      []json.RawMessage `json:"content"`
			Model        string            `json:"model"`
			StopReason   *string           `json:"stop_reason"`
			StopSequence *string           `json:"stop_sequence"`
			Usage        *Usage            `json:"usage"`
		}
		if err := json.Unmarshal(raw.Message, &apiMsg); err != nil {
			return nil, fmt.Errorf("failed to parse API message: %w", err)
		}

		contentBlocks := make([]ContentBlock, 0, len(apiMsg.Content))
		for _, rawBlock := range apiMsg.Content {
			block, err := ParseContentBlock(rawBlock)
			if err != nil {
				return nil, fmt.Errorf("failed to parse content block: %w", err)
			}
			contentBlocks = append(contentBlocks, block)
		}

		return &AssistantMessage{
			Type: raw.Type,
			Message: APIMessage{
				ID:           apiMsg.ID,
				Type:         apiMsg.Type,
				Role:         apiMsg.Role,
				Content:      contentBlocks,
				Model:        apiMsg.Model,
				StopReason:   apiMsg.StopReason,
				StopSequence: apiMsg.StopSequence,
				Usage:        apiMsg.Usage,
			},
			ParentToolUseID: raw.ParentToolUseID,
			Error:           raw.Error,
			UUID:            raw.UUID,
			SessionID:       raw.SessionID,
		}, nil

	case "result":
		var msg ResultMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse result message: %w", err)
		}
		return &msg, nil

	case "system":
		var msg SystemMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse system message: %w", err)
		}
		return &msg, nil

	case "stream_event":
		var msg StreamEvent
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse stream event: %w", err)
		}
		return &msg, nil

	case "tool_progress":
		var msg ToolProgressMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse tool progress message: %w", err)
		}
		return &msg, nil

	case "auth_status":
		var msg AuthStatusMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse auth status message: %w", err)
		}
		return &msg, nil

	case "user_message_replay":
		var msg UserMessageReplay
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse user message replay: %w", err)
		}
		return &msg, nil

	case "compact_boundary":
		var msg CompactBoundaryMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse compact boundary message: %w", err)
		}
		return &msg, nil

	case "status":
		var msg StatusMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse status message: %w", err)
		}
		return &msg, nil

	case "hook_started":
		var msg HookStartedMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse hook started message: %w", err)
		}
		return &msg, nil

	case "hook_progress":
		var msg HookProgressMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse hook progress message: %w", err)
		}
		return &msg, nil

	case "hook_response":
		var msg HookResponseMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse hook response message: %w", err)
		}
		return &msg, nil

	case "task_notification":
		var msg TaskNotificationMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse task notification message: %w", err)
		}
		return &msg, nil

	case "tool_use_summary":
		var msg ToolUseSummaryMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to parse tool use summary message: %w", err)
		}
		return &msg, nil

	default:
		// Unknown message types - parse as raw to avoid breaking on new CLI message types
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse unknown message type %s: %w", typeHolder.Type, err)
		}
		return &RawMessage{Type: typeHolder.Type, Data: raw}, nil
	}
}
