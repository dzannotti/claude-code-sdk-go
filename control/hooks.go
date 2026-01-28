package control

import "context"

type HookEvent string

const (
	HookPreToolUse         HookEvent = "PreToolUse"
	HookPostToolUse        HookEvent = "PostToolUse"
	HookPostToolUseFailure HookEvent = "PostToolUseFailure"
	HookNotification       HookEvent = "Notification"
	HookUserPromptSubmit   HookEvent = "UserPromptSubmit"
	HookSessionStart       HookEvent = "SessionStart"
	HookSessionEnd         HookEvent = "SessionEnd"
	HookStop               HookEvent = "Stop"
	HookSubagentStart      HookEvent = "SubagentStart"
	HookSubagentStop       HookEvent = "SubagentStop"
	HookPreCompact         HookEvent = "PreCompact"
	HookPermissionRequest  HookEvent = "PermissionRequest"
	HookSetup              HookEvent = "Setup"
)

type HookCallback func(
	ctx context.Context,
	input HookInput,
	toolUseID *string,
) (HookOutput, error)

type HookCallbackMatcher struct {
	Matcher *string
	Hooks   []HookCallback
	Timeout *int
}

type HookInput interface {
	HookEventName() HookEvent
}

type BaseHookInput struct {
	SessionID      string  `json:"session_id"`
	TranscriptPath string  `json:"transcript_path"`
	Cwd            string  `json:"cwd"`
	PermissionMode *string `json:"permission_mode,omitempty"`
}

type PreToolUseHookInput struct {
	BaseHookInput
	HookEvent string `json:"hook_event_name"`
	ToolName  string `json:"tool_name"`
	ToolInput any    `json:"tool_input"`
	ToolUseID string `json:"tool_use_id"`
}

func (h *PreToolUseHookInput) HookEventName() HookEvent { return HookPreToolUse }

type PostToolUseHookInput struct {
	BaseHookInput
	HookEvent    string `json:"hook_event_name"`
	ToolName     string `json:"tool_name"`
	ToolInput    any    `json:"tool_input"`
	ToolResponse any    `json:"tool_response"`
	ToolUseID    string `json:"tool_use_id"`
}

func (h *PostToolUseHookInput) HookEventName() HookEvent { return HookPostToolUse }

type PostToolUseFailureHookInput struct {
	BaseHookInput
	HookEvent   string `json:"hook_event_name"`
	ToolName    string `json:"tool_name"`
	ToolInput   any    `json:"tool_input"`
	ToolUseID   string `json:"tool_use_id"`
	Error       string `json:"error"`
	IsInterrupt *bool  `json:"is_interrupt,omitempty"`
}

func (h *PostToolUseFailureHookInput) HookEventName() HookEvent { return HookPostToolUseFailure }

type NotificationHookInput struct {
	BaseHookInput
	HookEvent        string  `json:"hook_event_name"`
	Message          string  `json:"message"`
	Title            *string `json:"title,omitempty"`
	NotificationType string  `json:"notification_type"`
}

func (h *NotificationHookInput) HookEventName() HookEvent { return HookNotification }

type UserPromptSubmitHookInput struct {
	BaseHookInput
	HookEvent string `json:"hook_event_name"`
	Prompt    string `json:"prompt"`
}

func (h *UserPromptSubmitHookInput) HookEventName() HookEvent { return HookUserPromptSubmit }

type SessionStartHookInput struct {
	BaseHookInput
	HookEvent string  `json:"hook_event_name"`
	Source    string  `json:"source"`
	AgentType *string `json:"agent_type,omitempty"`
	Model     *string `json:"model,omitempty"`
}

func (h *SessionStartHookInput) HookEventName() HookEvent { return HookSessionStart }

type SessionEndHookInput struct {
	BaseHookInput
	HookEvent string `json:"hook_event_name"`
	Reason    string `json:"reason"`
}

func (h *SessionEndHookInput) HookEventName() HookEvent { return HookSessionEnd }

type StopHookInput struct {
	BaseHookInput
	HookEvent      string `json:"hook_event_name"`
	StopHookActive bool   `json:"stop_hook_active"`
}

func (h *StopHookInput) HookEventName() HookEvent { return HookStop }

type SubagentStartHookInput struct {
	BaseHookInput
	HookEvent string `json:"hook_event_name"`
	AgentID   string `json:"agent_id"`
	AgentType string `json:"agent_type"`
}

func (h *SubagentStartHookInput) HookEventName() HookEvent { return HookSubagentStart }

type SubagentStopHookInput struct {
	BaseHookInput
	HookEvent           string `json:"hook_event_name"`
	StopHookActive      bool   `json:"stop_hook_active"`
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
}

func (h *SubagentStopHookInput) HookEventName() HookEvent { return HookSubagentStop }

type PreCompactHookInput struct {
	BaseHookInput
	HookEvent          string  `json:"hook_event_name"`
	Trigger            string  `json:"trigger"`
	CustomInstructions *string `json:"custom_instructions"`
}

func (h *PreCompactHookInput) HookEventName() HookEvent { return HookPreCompact }

type PermissionRequestHookInput struct {
	BaseHookInput
	HookEvent             string             `json:"hook_event_name"`
	ToolName              string             `json:"tool_name"`
	ToolInput             any                `json:"tool_input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
}

func (h *PermissionRequestHookInput) HookEventName() HookEvent { return HookPermissionRequest }

type SetupHookInput struct {
	BaseHookInput
	HookEvent string `json:"hook_event_name"`
	Trigger   string `json:"trigger"`
}

func (h *SetupHookInput) HookEventName() HookEvent { return HookSetup }

type HookOutput struct {
	Async              bool   `json:"async,omitempty"`
	AsyncTimeout       *int   `json:"asyncTimeout,omitempty"`
	Continue           *bool  `json:"continue,omitempty"`
	SuppressOutput     *bool  `json:"suppressOutput,omitempty"`
	StopReason         string `json:"stopReason,omitempty"`
	Decision           string `json:"decision,omitempty"`
	SystemMessage      string `json:"systemMessage,omitempty"`
	Reason             string `json:"reason,omitempty"`
	HookSpecificOutput any    `json:"hookSpecificOutput,omitempty"`
}
