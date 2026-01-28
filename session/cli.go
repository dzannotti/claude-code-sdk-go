// Package session provides utilities for loading conversation history from Claude CLI.
package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"claudeagent/message"
)

type Info struct {
	ID        string
	Path      string
	ModTime   time.Time
	SizeBytes int64
}

func CLIDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir: %w", err)
	}
	return filepath.Join(home, ".claude"), nil
}

func ProjectDir(workingDir string) (string, error) {
	cliDir, err := CLIDir()
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	// CLI uses path with dashes instead of slashes: /home/user/code -> -home-user-code
	projectHash := strings.ReplaceAll(absPath, string(filepath.Separator), "-")
	projectPath := filepath.Join(cliDir, "projects", projectHash)

	if _, err := os.Stat(projectPath); err != nil {
		return "", fmt.Errorf("project dir not found: %w", err)
	}

	return projectPath, nil
}

func ListSessions(projectDir string) ([]Info, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, fmt.Errorf("read project dir: %w", err)
	}

	sessions := make([]Info, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		// Skip agent sidechain files
		if strings.HasPrefix(entry.Name(), "agent-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		sessionID := strings.TrimSuffix(entry.Name(), ".jsonl")
		sessions = append(sessions, Info{
			ID:        sessionID,
			Path:      filepath.Join(projectDir, entry.Name()),
			ModTime:   info.ModTime(),
			SizeBytes: info.Size(),
		})
	}

	// Sort by modification time, most recent first
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ModTime.After(sessions[j].ModTime)
	})

	return sessions, nil
}

func Load(sessionPath string) ([]message.Message, error) {
	f, err := os.Open(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("open session file: %w", err)
	}
	defer f.Close()

	var messages []message.Message
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB max line

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		msg, err := parseSessionLine(line)
		if err != nil {
			continue // Skip unparseable lines
		}
		if msg != nil {
			messages = append(messages, msg)
		}
	}

	if err := scanner.Err(); err != nil {
		return messages, fmt.Errorf("scan session file: %w", err)
	}

	return messages, nil
}

func LoadByID(projectDir, sessionID string) ([]message.Message, error) {
	sessionPath := filepath.Join(projectDir, sessionID+".jsonl")
	return Load(sessionPath)
}

type sessionEntry struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
	UUID    string          `json:"uuid"`
}

func parseAPIMessage(data []byte) (message.APIMessage, error) {
	var raw struct {
		ID           string            `json:"id"`
		Type         string            `json:"type"`
		Role         string            `json:"role"`
		Content      []json.RawMessage `json:"content"`
		Model        string            `json:"model"`
		StopReason   *string           `json:"stop_reason"`
		StopSequence *string           `json:"stop_sequence"`
		Usage        *message.Usage    `json:"usage"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return message.APIMessage{}, err
	}

	blocks := make([]message.ContentBlock, 0, len(raw.Content))
	for _, rawBlock := range raw.Content {
		block, err := parseContentBlock(rawBlock)
		if err != nil {
			continue // Skip unparseable blocks
		}
		if block != nil {
			blocks = append(blocks, block)
		}
	}

	return message.APIMessage{
		ID:           raw.ID,
		Type:         raw.Type,
		Role:         raw.Role,
		Content:      blocks,
		Model:        raw.Model,
		StopReason:   raw.StopReason,
		StopSequence: raw.StopSequence,
		Usage:        raw.Usage,
	}, nil
}

func parseContentBlock(data []byte) (message.ContentBlock, error) {
	var typeOnly struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeOnly); err != nil {
		return nil, err
	}

	switch typeOnly.Type {
	case "text":
		var block message.TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case "thinking":
		var block message.ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case "tool_use":
		var block message.ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	case "tool_result":
		var block message.ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, err
		}
		return &block, nil
	default:
		return nil, nil
	}
}

func parseSessionLine(line []byte) (message.Message, error) {
	var entry sessionEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		return nil, err
	}

	switch entry.Type {
	case "user":
		var msg struct {
			Message struct {
				Role    string `json:"role"`
				Content any    `json:"content"`
			} `json:"message"`
			UUID string `json:"uuid"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			return nil, err
		}
		return &message.UserMessage{
			Type: "user",
			Message: message.UserContent{
				Role:    msg.Message.Role,
				Content: msg.Message.Content,
			},
			UUID: msg.UUID,
		}, nil

	case "assistant":
		var msg struct {
			Message json.RawMessage `json:"message"`
			UUID    string          `json:"uuid"`
		}
		if err := json.Unmarshal(line, &msg); err != nil {
			return nil, err
		}

		apiMsg, err := parseAPIMessage(msg.Message)
		if err != nil {
			return nil, err
		}

		return &message.AssistantMessage{
			Type:    "assistant",
			Message: apiMsg,
			UUID:    msg.UUID,
		}, nil

	case "result":
		var result message.ResultMessage
		if err := json.Unmarshal(line, &result); err != nil {
			return nil, err
		}
		return &result, nil

	default:
		// Skip queue-operation, system, and other internal types
		return nil, nil
	}
}
