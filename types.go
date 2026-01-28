package claudeagent

import (
	"claudeagent/control"
	"claudeagent/mcp"
	"claudeagent/message"
)

type Message = message.Message
type UserMessage = message.UserMessage
type AssistantMessage = message.AssistantMessage
type ResultMessage = message.ResultMessage
type SystemMessage = message.SystemMessage
type StreamEvent = message.StreamEvent
type ToolProgressMessage = message.ToolProgressMessage
type AuthStatusMessage = message.AuthStatusMessage
type UserMessageReplay = message.UserMessageReplay
type CompactBoundaryMessage = message.CompactBoundaryMessage
type StatusMessage = message.StatusMessage
type HookStartedMessage = message.HookStartedMessage
type HookProgressMessage = message.HookProgressMessage
type HookResponseMessage = message.HookResponseMessage
type TaskNotificationMessage = message.TaskNotificationMessage
type ToolUseSummaryMessage = message.ToolUseSummaryMessage
type RawMessage = message.RawMessage

type ContentBlock = message.ContentBlock
type TextBlock = message.TextBlock
type ThinkingBlock = message.ThinkingBlock
type ToolUseBlock = message.ToolUseBlock
type ToolResultBlock = message.ToolResultBlock

type Usage = message.Usage
type ModelUsage = message.ModelUsage
type PermissionDenial = message.PermissionDenial

type PermissionMode = control.PermissionMode
type PermissionBehavior = control.PermissionBehavior
type PermissionResult = control.PermissionResult
type PermissionUpdate = control.PermissionUpdate
type CanUseToolFunc = control.CanUseToolFunc
type CanUseToolOptions = control.CanUseToolOptions

type HookEvent = control.HookEvent
type HookCallback = control.HookCallback
type HookCallbackMatcher = control.HookCallbackMatcher
type HookInput = control.HookInput
type HookOutput = control.HookOutput

type McpServerConfig = mcp.ServerConfig
type McpStdioServerConfig = mcp.StdioServerConfig
type McpSSEServerConfig = mcp.SSEServerConfig
type McpHTTPServerConfig = mcp.HTTPServerConfig
type McpSdkServerConfig = mcp.SdkServerConfig
type McpClaudeAIProxyServerConfig = mcp.ClaudeAIProxyServerConfig
type McpServerStatus = mcp.ServerStatus
type McpSetServersResult = mcp.SetServersResult
type McpServerInfo = mcp.ServerInfo
type McpServerStatusConfig = mcp.ServerStatusConfig
type McpToolInfo = mcp.ToolInfo
type McpToolAnnotations = mcp.ToolAnnotations
type McpServer = mcp.McpServer
type McpToolDefinition = mcp.ToolDefinition
type McpToolResult = mcp.ToolResult
type McpToolResultContent = mcp.ToolResultContent
type SdkMcpServer = mcp.SdkMcpServer

const (
	PermissionModeDefault           = control.PermissionModeDefault
	PermissionModeAcceptEdits       = control.PermissionModeAcceptEdits
	PermissionModeBypassPermissions = control.PermissionModeBypassPermissions
	PermissionModePlan              = control.PermissionModePlan
	PermissionModeDelegate          = control.PermissionModeDelegate
	PermissionModeDontAsk           = control.PermissionModeDontAsk

	PermissionAllow = control.PermissionAllow
	PermissionDeny  = control.PermissionDeny
	PermissionAsk   = control.PermissionAsk

	HookPreToolUse         = control.HookPreToolUse
	HookPostToolUse        = control.HookPostToolUse
	HookPostToolUseFailure = control.HookPostToolUseFailure
	HookNotification       = control.HookNotification
	HookUserPromptSubmit   = control.HookUserPromptSubmit
	HookSessionStart       = control.HookSessionStart
	HookSessionEnd         = control.HookSessionEnd
	HookStop               = control.HookStop
	HookSubagentStart      = control.HookSubagentStart
	HookSubagentStop       = control.HookSubagentStop
	HookPreCompact         = control.HookPreCompact
	HookPermissionRequest  = control.HookPermissionRequest
	HookSetup              = control.HookSetup
)

type SlashCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ArgumentHint string `json:"argumentHint"`
}

type ModelInfo struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type AccountInfo struct {
	Email            *string `json:"email,omitempty"`
	Organization     *string `json:"organization,omitempty"`
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	TokenSource      *string `json:"tokenSource,omitempty"`
	APIKeySource     *string `json:"apiKeySource,omitempty"`
}
