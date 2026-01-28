package control

import "context"

type PermissionMode string

const (
	PermissionModeDefault           PermissionMode = "default"
	PermissionModeAcceptEdits       PermissionMode = "acceptEdits"
	PermissionModeBypassPermissions PermissionMode = "bypassPermissions"
	PermissionModePlan              PermissionMode = "plan"
	PermissionModeDelegate          PermissionMode = "delegate"
	PermissionModeDontAsk           PermissionMode = "dontAsk"
)

type PermissionBehavior string

const (
	PermissionAllow PermissionBehavior = "allow"
	PermissionDeny  PermissionBehavior = "deny"
	PermissionAsk   PermissionBehavior = "ask"
)

type CanUseToolFunc func(
	ctx context.Context,
	toolName string,
	input map[string]any,
	options CanUseToolOptions,
) (PermissionResult, error)

type CanUseToolOptions struct {
	Suggestions    []PermissionUpdate
	BlockedPath    *string
	DecisionReason *string
	ToolUseID      string
	AgentID        *string
}

type PermissionResult struct {
	Behavior           PermissionBehavior
	UpdatedInput       map[string]any
	UpdatedPermissions []PermissionUpdate
	Message            string
	Interrupt          bool
	ToolUseID          *string
}

type PermissionUpdate struct {
	Type        string             `json:"type"`
	Rules       []PermissionRule   `json:"rules,omitempty"`
	Behavior    PermissionBehavior `json:"behavior,omitempty"`
	Destination string             `json:"destination,omitempty"`
	Mode        *PermissionMode    `json:"mode,omitempty"`
	Directories []string           `json:"directories,omitempty"`
}

type PermissionRule struct {
	ToolName    string `json:"toolName"`
	RuleContent string `json:"ruleContent,omitempty"`
}
