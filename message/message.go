package message

type Message interface {
	MessageType() string
	GetSessionID() string
	GetUUID() string
}

type UserMessage struct {
	Type            string      `json:"type"`
	Message         UserContent `json:"message"`
	ParentToolUseID *string     `json:"parent_tool_use_id"`
	IsSynthetic     bool        `json:"isSynthetic,omitempty"`
	ToolUseResult   any         `json:"tool_use_result,omitempty"`
	UUID            string      `json:"uuid,omitempty"`
	SessionID       string      `json:"session_id"`
}

type UserContent struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func (m *UserMessage) MessageType() string  { return "user" }
func (m *UserMessage) GetSessionID() string { return m.SessionID }
func (m *UserMessage) GetUUID() string      { return m.UUID }

type AssistantMessage struct {
	Type            string     `json:"type"`
	Message         APIMessage `json:"message"`
	ParentToolUseID *string    `json:"parent_tool_use_id"`
	Error           *string    `json:"error,omitempty"`
	UUID            string     `json:"uuid"`
	SessionID       string     `json:"session_id"`
}

type APIMessage struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Role         string         `json:"role"`
	Content      []ContentBlock `json:"content"`
	Model        string         `json:"model"`
	StopReason   *string        `json:"stop_reason"`
	StopSequence *string        `json:"stop_sequence"`
	Usage        *Usage         `json:"usage"`
}

func (m *AssistantMessage) MessageType() string  { return "assistant" }
func (m *AssistantMessage) GetSessionID() string { return m.SessionID }
func (m *AssistantMessage) GetUUID() string      { return m.UUID }

type ResultMessage struct {
	Type              string                `json:"type"`
	Subtype           string                `json:"subtype"`
	DurationMS        int                   `json:"duration_ms"`
	DurationAPIMS     int                   `json:"duration_api_ms"`
	IsError           bool                  `json:"is_error"`
	NumTurns          int                   `json:"num_turns"`
	Result            string                `json:"result,omitempty"`
	TotalCostUSD      float64               `json:"total_cost_usd"`
	Usage             *Usage                `json:"usage"`
	ModelUsage        map[string]ModelUsage `json:"modelUsage"`
	PermissionDenials []PermissionDenial    `json:"permission_denials"`
	StructuredOutput  any                   `json:"structured_output,omitempty"`
	Errors            []string              `json:"errors,omitempty"`
	UUID              string                `json:"uuid"`
	SessionID         string                `json:"session_id"`
}

func (m *ResultMessage) MessageType() string  { return "result" }
func (m *ResultMessage) GetSessionID() string { return m.SessionID }
func (m *ResultMessage) GetUUID() string      { return m.UUID }

type SystemMessage struct {
	Type              string       `json:"type"`
	Subtype           string       `json:"subtype"`
	Agents            []string     `json:"agents,omitempty"`
	APIKeySource      string       `json:"apiKeySource,omitempty"`
	Betas             []string     `json:"betas,omitempty"`
	ClaudeCodeVersion string       `json:"claude_code_version,omitempty"`
	Cwd               string       `json:"cwd,omitempty"`
	Tools             []string     `json:"tools,omitempty"`
	McpServers        []McpServer  `json:"mcp_servers,omitempty"`
	Model             string       `json:"model,omitempty"`
	PermissionMode    string       `json:"permissionMode,omitempty"`
	SlashCommands     []string     `json:"slash_commands,omitempty"`
	OutputStyle       string       `json:"output_style,omitempty"`
	Skills            []string     `json:"skills,omitempty"`
	Plugins           []PluginInfo `json:"plugins,omitempty"`
	UUID              string       `json:"uuid"`
	SessionID         string       `json:"session_id"`

	CompactMetadata *CompactMetadata `json:"compact_metadata,omitempty"`
	Status          *string          `json:"status,omitempty"`

	HookName  string `json:"hook_name,omitempty"`
	HookEvent string `json:"hook_event,omitempty"`
	Stdout    string `json:"stdout,omitempty"`
	Stderr    string `json:"stderr,omitempty"`
	ExitCode  *int   `json:"exit_code,omitempty"`
}

func (m *SystemMessage) MessageType() string  { return "system" }
func (m *SystemMessage) GetSessionID() string { return m.SessionID }
func (m *SystemMessage) GetUUID() string      { return m.UUID }

type StreamEvent struct {
	Type            string  `json:"type"`
	Event           any     `json:"event"`
	ParentToolUseID *string `json:"parent_tool_use_id"`
	UUID            string  `json:"uuid"`
	SessionID       string  `json:"session_id"`
}

func (m *StreamEvent) MessageType() string  { return "stream_event" }
func (m *StreamEvent) GetSessionID() string { return m.SessionID }
func (m *StreamEvent) GetUUID() string      { return m.UUID }

type ToolProgressMessage struct {
	Type               string  `json:"type"`
	ToolUseID          string  `json:"tool_use_id"`
	ToolName           string  `json:"tool_name"`
	ParentToolUseID    *string `json:"parent_tool_use_id"`
	ElapsedTimeSeconds float64 `json:"elapsed_time_seconds"`
	UUID               string  `json:"uuid"`
	SessionID          string  `json:"session_id"`
}

func (m *ToolProgressMessage) MessageType() string  { return "tool_progress" }
func (m *ToolProgressMessage) GetSessionID() string { return m.SessionID }
func (m *ToolProgressMessage) GetUUID() string      { return m.UUID }

type AuthStatusMessage struct {
	Type             string   `json:"type"`
	IsAuthenticating bool     `json:"isAuthenticating"`
	Output           []string `json:"output"`
	Error            *string  `json:"error,omitempty"`
	UUID             string   `json:"uuid"`
	SessionID        string   `json:"session_id"`
}

func (m *AuthStatusMessage) MessageType() string  { return "auth_status" }
func (m *AuthStatusMessage) GetSessionID() string { return m.SessionID }
func (m *AuthStatusMessage) GetUUID() string      { return m.UUID }

// RawMessage captures unknown message types to avoid breaking on new CLI message types.
// This provides forward compatibility - the SDK won't error on new types it doesn't recognize.
type RawMessage struct {
	Type string
	Data map[string]any
}

func (m *RawMessage) MessageType() string { return m.Type }
func (m *RawMessage) GetSessionID() string {
	if s, ok := m.Data["session_id"].(string); ok {
		return s
	}
	return ""
}
func (m *RawMessage) GetUUID() string {
	if s, ok := m.Data["uuid"].(string); ok {
		return s
	}
	return ""
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

type ModelUsage struct {
	InputTokens              int     `json:"inputTokens"`
	OutputTokens             int     `json:"outputTokens"`
	CacheReadInputTokens     int     `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int     `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
	ContextWindow            int     `json:"contextWindow"`
}

type PermissionDenial struct {
	ToolName  string         `json:"tool_name"`
	ToolUseID string         `json:"tool_use_id"`
	ToolInput map[string]any `json:"tool_input"`
}

type McpServer struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PluginInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type CompactMetadata struct {
	Trigger   string `json:"trigger"`
	PreTokens int    `json:"pre_tokens"`
}
