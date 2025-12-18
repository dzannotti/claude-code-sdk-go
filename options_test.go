package claudecode

import (
	"context"
	"testing"

	"claudecode/control"
	"claudecode/mcp"
)

func TestWithModel(t *testing.T) {
	opts := applyOptions([]Option{WithModel("claude-3")})
	if opts.Model == nil || *opts.Model != "claude-3" {
		t.Errorf("expected model 'claude-3', got %v", opts.Model)
	}
}

func TestWithSystemPrompt(t *testing.T) {
	opts := applyOptions([]Option{WithSystemPrompt("You are helpful")})
	if opts.SystemPrompt == nil || opts.SystemPrompt.Prompt != "You are helpful" {
		t.Errorf("expected prompt 'You are helpful', got %v", opts.SystemPrompt)
	}
}

func TestWithSystemPromptPreset(t *testing.T) {
	opts := applyOptions([]Option{WithSystemPromptPreset("claude_code", "Be concise")})
	if opts.SystemPrompt == nil {
		t.Fatal("expected SystemPrompt to be set")
	}
	if opts.SystemPrompt.Preset != "claude_code" {
		t.Errorf("expected preset 'claude_code', got %q", opts.SystemPrompt.Preset)
	}
	if opts.SystemPrompt.Append != "Be concise" {
		t.Errorf("expected append 'Be concise', got %q", opts.SystemPrompt.Append)
	}
}

func TestWithTools(t *testing.T) {
	opts := applyOptions([]Option{WithTools("Bash", "Read", "Write")})
	if len(opts.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(opts.Tools))
	}
}

func TestWithAllowedTools(t *testing.T) {
	opts := applyOptions([]Option{WithAllowedTools("Read", "Write")})
	if len(opts.AllowedTools) != 2 {
		t.Errorf("expected 2 allowed tools, got %d", len(opts.AllowedTools))
	}
}

func TestWithDisallowedTools(t *testing.T) {
	opts := applyOptions([]Option{WithDisallowedTools("Bash")})
	if len(opts.DisallowedTools) != 1 {
		t.Errorf("expected 1 disallowed tool, got %d", len(opts.DisallowedTools))
	}
}

func TestWithPermissionMode(t *testing.T) {
	opts := applyOptions([]Option{WithPermissionMode(control.PermissionModeAcceptEdits)})
	if opts.PermissionMode == nil || *opts.PermissionMode != control.PermissionModeAcceptEdits {
		t.Errorf("expected permission mode 'acceptEdits', got %v", opts.PermissionMode)
	}
}

func TestWithCanUseTool(t *testing.T) {
	fn := func(ctx context.Context, toolName string, input map[string]any, opts control.CanUseToolOptions) (control.PermissionResult, error) {
		return control.PermissionResult{Behavior: control.PermissionAllow}, nil
	}
	opts := applyOptions([]Option{WithCanUseTool(fn)})
	if opts.CanUseTool == nil {
		t.Error("expected CanUseTool to be set")
	}
}

func TestWithAllowDangerouslySkipPermissions(t *testing.T) {
	opts := applyOptions([]Option{WithAllowDangerouslySkipPermissions()})
	if !opts.AllowDangerouslySkipPermissions {
		t.Error("expected AllowDangerouslySkipPermissions to be true")
	}
}

func TestWithHooks(t *testing.T) {
	matcher := control.HookCallbackMatcher{
		Hooks: []control.HookCallback{},
	}
	opts := applyOptions([]Option{WithHooks(control.HookPreToolUse, matcher)})
	if len(opts.Hooks[control.HookPreToolUse]) != 1 {
		t.Errorf("expected 1 hook matcher, got %d", len(opts.Hooks[control.HookPreToolUse]))
	}
}

func TestWithMcpServers(t *testing.T) {
	servers := map[string]mcp.ServerConfig{
		"test": mcp.StdioServerConfig{Command: "node", Args: []string{"server.js"}},
	}
	opts := applyOptions([]Option{WithMcpServers(servers)})
	if len(opts.McpServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(opts.McpServers))
	}
}

func TestWithCwd(t *testing.T) {
	opts := applyOptions([]Option{WithCwd("/home/user")})
	if opts.Cwd == nil || *opts.Cwd != "/home/user" {
		t.Errorf("expected cwd '/home/user', got %v", opts.Cwd)
	}
}

func TestWithAdditionalDirectories(t *testing.T) {
	opts := applyOptions([]Option{WithAdditionalDirectories("/tmp", "/var")})
	if len(opts.AdditionalDirectories) != 2 {
		t.Errorf("expected 2 directories, got %d", len(opts.AdditionalDirectories))
	}
}

func TestWithResume(t *testing.T) {
	opts := applyOptions([]Option{WithResume("session-123")})
	if opts.Resume == nil || *opts.Resume != "session-123" {
		t.Errorf("expected resume 'session-123', got %v", opts.Resume)
	}
}

func TestWithContinue(t *testing.T) {
	opts := applyOptions([]Option{WithContinue()})
	if !opts.Continue {
		t.Error("expected Continue to be true")
	}
}

func TestWithMaxTurns(t *testing.T) {
	opts := applyOptions([]Option{WithMaxTurns(10)})
	if opts.MaxTurns == nil || *opts.MaxTurns != 10 {
		t.Errorf("expected max turns 10, got %v", opts.MaxTurns)
	}
}

func TestWithMaxBudgetUSD(t *testing.T) {
	opts := applyOptions([]Option{WithMaxBudgetUSD(1.5)})
	if opts.MaxBudgetUSD == nil || *opts.MaxBudgetUSD != 1.5 {
		t.Errorf("expected max budget 1.5, got %v", opts.MaxBudgetUSD)
	}
}

func TestWithMaxThinkingTokens(t *testing.T) {
	opts := applyOptions([]Option{WithMaxThinkingTokens(1000)})
	if opts.MaxThinkingTokens == nil || *opts.MaxThinkingTokens != 1000 {
		t.Errorf("expected max thinking tokens 1000, got %v", opts.MaxThinkingTokens)
	}
}

func TestWithIncludePartialMessages(t *testing.T) {
	opts := applyOptions([]Option{WithIncludePartialMessages()})
	if !opts.IncludePartialMessages {
		t.Error("expected IncludePartialMessages to be true")
	}
}

func TestWithOutputFormat(t *testing.T) {
	format := OutputFormat{Type: "json_schema", Schema: map[string]any{"type": "object"}}
	opts := applyOptions([]Option{WithOutputFormat(format)})
	if opts.OutputFormat == nil || opts.OutputFormat.Type != "json_schema" {
		t.Errorf("expected output format 'json_schema', got %v", opts.OutputFormat)
	}
}

func TestWithAgents(t *testing.T) {
	agents := map[string]AgentDefinition{
		"reviewer": {Description: "Reviews code", Prompt: "You are a reviewer"},
	}
	opts := applyOptions([]Option{WithAgents(agents)})
	if len(opts.Agents) != 1 {
		t.Errorf("expected 1 agent, got %d", len(opts.Agents))
	}
}

func TestWithSandbox(t *testing.T) {
	settings := SandboxSettings{Enabled: true}
	opts := applyOptions([]Option{WithSandbox(settings)})
	if opts.Sandbox == nil || !opts.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}
}

func TestWithEnableFileCheckpointing(t *testing.T) {
	opts := applyOptions([]Option{WithEnableFileCheckpointing()})
	if !opts.EnableFileCheckpointing {
		t.Error("expected EnableFileCheckpointing to be true")
	}
}

func TestWithPlugins(t *testing.T) {
	plugins := []PluginConfig{{Type: "local", Path: "./plugin"}}
	opts := applyOptions([]Option{WithPlugins(plugins...)})
	if len(opts.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(opts.Plugins))
	}
}

func TestWithSettingSources(t *testing.T) {
	opts := applyOptions([]Option{WithSettingSources("user", "project")})
	if len(opts.SettingSources) != 2 {
		t.Errorf("expected 2 setting sources, got %d", len(opts.SettingSources))
	}
}

func TestWithBetas(t *testing.T) {
	opts := applyOptions([]Option{WithBetas("context-1m-2025-08-07")})
	if len(opts.Betas) != 1 {
		t.Errorf("expected 1 beta, got %d", len(opts.Betas))
	}
}

func TestWithEnv(t *testing.T) {
	env := map[string]string{"FOO": "bar"}
	opts := applyOptions([]Option{WithEnv(env)})
	if opts.Env["FOO"] != "bar" {
		t.Errorf("expected env FOO='bar', got %v", opts.Env["FOO"])
	}
}

func TestWithEnvVar(t *testing.T) {
	opts := applyOptions([]Option{WithEnvVar("KEY", "value")})
	if opts.Env["KEY"] != "value" {
		t.Errorf("expected env KEY='value', got %v", opts.Env["KEY"])
	}
}

func TestWithCLIPath(t *testing.T) {
	opts := applyOptions([]Option{WithCLIPath("/usr/bin/claude")})
	if opts.CLIPath == nil || *opts.CLIPath != "/usr/bin/claude" {
		t.Errorf("expected CLI path '/usr/bin/claude', got %v", opts.CLIPath)
	}
}

func TestWithExecutable(t *testing.T) {
	opts := applyOptions([]Option{WithExecutable("bun", "--flag")})
	if opts.Executable == nil || *opts.Executable != "bun" {
		t.Errorf("expected executable 'bun', got %v", opts.Executable)
	}
	if len(opts.ExecutableArgs) != 1 || opts.ExecutableArgs[0] != "--flag" {
		t.Errorf("expected args ['--flag'], got %v", opts.ExecutableArgs)
	}
}

func TestWithStderr(t *testing.T) {
	var called bool
	fn := func(s string) { called = true }
	opts := applyOptions([]Option{WithStderr(fn)})
	if opts.Stderr == nil {
		t.Error("expected Stderr to be set")
	}
	opts.Stderr("test")
	if !called {
		t.Error("expected Stderr function to be called")
	}
}

func TestWithExtraArg(t *testing.T) {
	value := "value"
	opts := applyOptions([]Option{WithExtraArg("flag", &value)})
	if opts.ExtraArgs["flag"] == nil || *opts.ExtraArgs["flag"] != "value" {
		t.Errorf("expected extra arg flag='value', got %v", opts.ExtraArgs["flag"])
	}
}

func TestWithForkSession(t *testing.T) {
	opts := applyOptions([]Option{WithForkSession()})
	if !opts.ForkSession {
		t.Error("expected ForkSession to be true")
	}
}

func TestWithPersistSession(t *testing.T) {
	opts := applyOptions([]Option{WithPersistSession(false)})
	if opts.PersistSession == nil || *opts.PersistSession != false {
		t.Errorf("expected PersistSession to be false, got %v", opts.PersistSession)
	}
}

func TestMultipleOptions(t *testing.T) {
	opts := applyOptions([]Option{
		WithModel("claude-3"),
		WithMaxTurns(5),
		WithPermissionMode(control.PermissionModePlan),
		WithIncludePartialMessages(),
	})

	if opts.Model == nil || *opts.Model != "claude-3" {
		t.Error("model not set correctly")
	}
	if opts.MaxTurns == nil || *opts.MaxTurns != 5 {
		t.Error("max turns not set correctly")
	}
	if opts.PermissionMode == nil || *opts.PermissionMode != control.PermissionModePlan {
		t.Error("permission mode not set correctly")
	}
	if !opts.IncludePartialMessages {
		t.Error("include partial messages not set correctly")
	}
}
