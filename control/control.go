package control

type Request struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Request   RequestPayload `json:"request"`
}

type RequestPayload struct {
	Subtype            string          `json:"subtype"`
	Mode               *PermissionMode `json:"mode,omitempty"`
	Model              *string         `json:"model,omitempty"`
	MaxThinkingTokens  *int            `json:"maxThinkingTokens,omitempty"`
	UserMessageID      string          `json:"userMessageId,omitempty"`
	Servers            map[string]any  `json:"servers,omitempty"`
	ServerName         string          `json:"serverName,omitempty"`
	Message            any             `json:"message,omitempty"`
	Hooks              map[string]any  `json:"hooks,omitempty"`
	SdkMcpServers      []string        `json:"sdkMcpServers,omitempty"`
	JSONSchema         map[string]any  `json:"jsonSchema,omitempty"`
	SystemPrompt       *string         `json:"systemPrompt,omitempty"`
	AppendSystemPrompt *string         `json:"appendSystemPrompt,omitempty"`
	Agents             map[string]any  `json:"agents,omitempty"`
}

type Response struct {
	Type     string          `json:"type"`
	Response ResponsePayload `json:"response"`
}

type ResponsePayload struct {
	Subtype   string         `json:"subtype"`
	RequestID string         `json:"request_id"`
	Response  map[string]any `json:"response,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type CanUseToolRequest struct {
	Subtype               string             `json:"subtype"`
	ToolName              string             `json:"tool_name"`
	Input                 map[string]any     `json:"input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
	BlockedPath           *string            `json:"blocked_path,omitempty"`
	DecisionReason        *string            `json:"decision_reason,omitempty"`
	ToolUseID             string             `json:"tool_use_id"`
	AgentID               *string            `json:"agent_id,omitempty"`
}

type HookCallbackRequest struct {
	Subtype    string  `json:"subtype"`
	CallbackID string  `json:"callback_id"`
	Input      any     `json:"input"`
	ToolUseID  *string `json:"tool_use_id,omitempty"`
}

type MCPMessageRequest struct {
	Subtype    string `json:"subtype"`
	ServerName string `json:"server_name"`
	Message    any    `json:"message"`
}
