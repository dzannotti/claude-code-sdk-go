package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"claudecode/control"
	"claudecode/mcp"
)

type CommandOptions struct {
	AllowedTools                    []string
	DisallowedTools                 []string
	SystemPrompt                    *string
	AppendSystemPrompt              *string
	Model                           *string
	MaxThinkingTokens               *int
	PermissionMode                  *control.PermissionMode
	PermissionPromptToolName        *string
	Continue                        bool
	Resume                          *string
	MaxTurns                        *int
	MaxBudgetUSD                    *float64
	Cwd                             *string
	AdditionalDirectories           []string
	McpServers                      map[string]mcp.ServerConfig
	Betas                           []string
	ExtraArgs                       map[string]*string
	SettingSources                  []string
	AllowDangerouslySkipPermissions bool
	IncludePartialMessages          bool
}

func BuildCommand(cliPath string, opts *CommandOptions, closeStdin bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	if closeStdin {
		cmd = append(cmd, "--print")
	} else {
		cmd = append(cmd, "--input-format", "stream-json")
	}

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

	for flag, value := range opts.ExtraArgs {
		if value == nil {
			cmd = append(cmd, "--"+flag)
		} else {
			cmd = append(cmd, "--"+flag, *value)
		}
	}

	return cmd
}

func BuildCommandWithPrompt(cliPath string, opts *CommandOptions, prompt string) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose", "--print", prompt}

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
	if opts.PermissionMode != nil {
		cmd = append(cmd, "--permission-mode", string(*opts.PermissionMode))
	}
	if opts.Continue {
		cmd = append(cmd, "--continue")
	}
	if opts.Resume != nil {
		cmd = append(cmd, "--resume", *opts.Resume)
	}
	if opts.MaxTurns != nil {
		cmd = append(cmd, "--max-turns", fmt.Sprintf("%d", *opts.MaxTurns))
	}
	if opts.Cwd != nil {
		cmd = append(cmd, "--cwd", *opts.Cwd)
	}
	for _, dir := range opts.AdditionalDirectories {
		cmd = append(cmd, "--add-dir", dir)
	}
	if opts.AllowDangerouslySkipPermissions {
		cmd = append(cmd, "--dangerously-skip-permissions")
	}
	if opts.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
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
