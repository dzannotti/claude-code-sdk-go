package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"claudeagent/control"
	"claudeagent/mcp"
)

type CommandOptions struct {
	AllowedTools                    []string
	DisallowedTools                 []string
	SystemPrompt                    *string
	AppendSystemPrompt              *string
	Model                           *string
	FallbackModel                   *string
	MaxThinkingTokens               *int
	PermissionMode                  *control.PermissionMode
	PermissionPromptToolName        *string
	Continue                        bool
	Resume                          *string
	ResumeSessionAt                 *string
	ForkSession                     bool
	PersistSession                  *bool
	MaxTurns                        *int
	MaxBudgetUSD                    *float64
	Cwd                             *string
	AdditionalDirectories           []string
	McpServers                      map[string]mcp.ServerConfig
	StrictMcpConfig                 bool
	Agent                           *string
	EnableFileCheckpointing         bool
	Betas                           []string
	ExtraArgs                       map[string]*string
	SettingSources                  []string
	AllowDangerouslySkipPermissions bool
	IncludePartialMessages          bool
	Tools                           any
	Sandbox                         any
	Plugins                         any
	OutputFormat                    any
	Executable                      *string
	ExecutableArgs                  []string
}

func BuildCommand(cliPath string, opts *CommandOptions, closeStdin bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	if closeStdin {
		cmd = append(cmd, "--print")
	} else {
		cmd = append(cmd, "--input-format", "stream-json")
	}

	return appendFlags(cmd, opts)
}

func BuildCommandWithPrompt(cliPath string, opts *CommandOptions, prompt string) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose", "--print", prompt}
	return appendFlags(cmd, opts)
}

func appendFlags(cmd []string, opts *CommandOptions) []string {
	if opts == nil {
		return cmd
	}

	if len(opts.AllowedTools) > 0 {
		cmd = append(cmd, "--allowed-tools", strings.Join(opts.AllowedTools, ","))
	}
	if len(opts.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowed-tools", strings.Join(opts.DisallowedTools, ","))
	}
	if opts.SystemPrompt != nil {
		cmd = append(cmd, "--system-prompt", *opts.SystemPrompt)
	}
	if opts.AppendSystemPrompt != nil {
		cmd = append(cmd, "--append-system-prompt", *opts.AppendSystemPrompt)
	}
	if opts.Model != nil {
		cmd = append(cmd, "--model", *opts.Model)
	}
	if opts.FallbackModel != nil {
		cmd = append(cmd, "--fallback-model", *opts.FallbackModel)
	}
	if opts.MaxThinkingTokens != nil {
		cmd = append(cmd, "--max-thinking-tokens", fmt.Sprintf("%d", *opts.MaxThinkingTokens))
	}
	if opts.PermissionMode != nil {
		cmd = append(cmd, "--permission-mode", string(*opts.PermissionMode))
	}
	if opts.PermissionPromptToolName != nil {
		cmd = append(cmd, "--permission-prompt-tool", *opts.PermissionPromptToolName)
	}
	if opts.Continue {
		cmd = append(cmd, "--continue")
	}
	if opts.Resume != nil {
		cmd = append(cmd, "--resume", *opts.Resume)
	}
	if opts.ResumeSessionAt != nil {
		cmd = append(cmd, "--resume-at", *opts.ResumeSessionAt)
	}
	if opts.ForkSession {
		cmd = append(cmd, "--fork-session")
	}
	if opts.PersistSession != nil && !*opts.PersistSession {
		cmd = append(cmd, "--no-persist")
	}
	if opts.MaxTurns != nil {
		cmd = append(cmd, "--max-turns", fmt.Sprintf("%d", *opts.MaxTurns))
	}
	if opts.MaxBudgetUSD != nil {
		cmd = append(cmd, "--max-budget-usd", fmt.Sprintf("%.2f", *opts.MaxBudgetUSD))
	}
	if opts.Cwd != nil {
		cmd = append(cmd, "--cwd", *opts.Cwd)
	}
	for _, dir := range opts.AdditionalDirectories {
		cmd = append(cmd, "--add-dir", dir)
	}
	if len(opts.McpServers) > 0 {
		data, err := json.Marshal(opts.McpServers)
		if err == nil {
			cmd = append(cmd, "--mcp-servers", string(data))
		}
	}
	if opts.StrictMcpConfig {
		cmd = append(cmd, "--strict-mcp-config")
	}
	if opts.Agent != nil {
		cmd = append(cmd, "--agent", *opts.Agent)
	}
	if opts.EnableFileCheckpointing {
		cmd = append(cmd, "--enable-file-checkpointing")
	}
	for _, beta := range opts.Betas {
		cmd = append(cmd, "--beta", beta)
	}
	for _, source := range opts.SettingSources {
		cmd = append(cmd, "--settings-source", source)
	}
	if opts.AllowDangerouslySkipPermissions {
		cmd = append(cmd, "--dangerously-skip-permissions")
	}
	if opts.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}
	if opts.Tools != nil {
		if data, err := json.Marshal(opts.Tools); err == nil {
			cmd = append(cmd, "--tools", string(data))
		}
	}
	if opts.Sandbox != nil {
		if data, err := json.Marshal(opts.Sandbox); err == nil {
			cmd = append(cmd, "--sandbox", string(data))
		}
	}
	if opts.Plugins != nil {
		if data, err := json.Marshal(opts.Plugins); err == nil {
			cmd = append(cmd, "--plugins", string(data))
		}
	}
	if opts.OutputFormat != nil {
		if data, err := json.Marshal(opts.OutputFormat); err == nil {
			cmd = append(cmd, "--output-format-config", string(data))
		}
	}

	for flag, value := range opts.ExtraArgs {
		if value == nil {
			cmd = append(cmd, "--"+flag)
		} else {
			cmd = append(cmd, "--"+flag, *value)
		}
	}

	return cmd
}
