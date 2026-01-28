package claudeagent

import (
	"context"
	"encoding/json"
	"io"

	"claudeagent/control"
	"claudeagent/mcp"
)

type Options struct {
	Tools                           *ToolsConfig
	AllowedTools                    []string
	DisallowedTools                 []string
	SystemPrompt                    *SystemPromptConfig
	Model                           *string
	FallbackModel                   *string
	MaxThinkingTokens               *int
	MaxTurns                        *int
	MaxBudgetUSD                    *float64
	PermissionMode                  *control.PermissionMode
	CanUseTool                      control.CanUseToolFunc
	PermissionPromptToolName        *string
	AllowDangerouslySkipPermissions bool
	Continue                        bool
	Resume                          *string
	ResumeSessionAt                 *string
	ForkSession                     bool
	PersistSession                  *bool
	Hooks                           map[control.HookEvent][]control.HookCallbackMatcher
	McpServers                      map[string]mcp.ServerConfig
	StrictMcpConfig                 bool
	Cwd                             *string
	AdditionalDirectories           []string
	EnableFileCheckpointing         bool
	Agent                           *string
	Agents                          map[string]AgentDefinition
	Sandbox                         *SandboxSettings
	IncludePartialMessages          bool
	OutputFormat                    *OutputFormat
	Plugins                         []PluginConfig
	SettingSources                  []string
	Betas                           []string
	Executable                      *string
	ExecutableArgs                  []string
	Env                             map[string]string
	CLIPath                         *string
	ExtraArgs                       map[string]*string
	Stderr                          func(data string)
	SpawnClaudeCodeProcess          SpawnFunc
}

type SystemPromptConfig struct {
	Prompt string
	Preset string
	Append string
}

type AgentDefinition struct {
	Description                        string               `json:"description"`
	Tools                              []string             `json:"tools,omitempty"`
	DisallowedTools                    []string             `json:"disallowedTools,omitempty"`
	Prompt                             string               `json:"prompt"`
	Model                              string               `json:"model,omitempty"`
	McpServers                         []AgentMcpServerSpec `json:"mcpServers,omitempty"`
	Skills                             []string             `json:"skills,omitempty"`
	MaxTurns                           *int                 `json:"maxTurns,omitempty"`
	CriticalSystemReminderExperimental string               `json:"criticalSystemReminder_EXPERIMENTAL,omitempty"`
}

type AgentMcpServerSpec struct {
	Name    string                      `json:"-"`
	Servers map[string]mcp.ServerConfig `json:"-"`
}

func (a AgentMcpServerSpec) MarshalJSON() ([]byte, error) {
	if a.Name != "" {
		return json.Marshal(a.Name)
	}
	return json.Marshal(a.Servers)
}

func (a *AgentMcpServerSpec) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if data[0] == '"' {
		return json.Unmarshal(data, &a.Name)
	}
	return json.Unmarshal(data, &a.Servers)
}

type SandboxSettings struct {
	Enabled                   bool             `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed  bool             `json:"autoAllowBashIfSandboxed,omitempty"`
	AllowUnsandboxedCommands  bool             `json:"allowUnsandboxedCommands,omitempty"`
	ExcludedCommands          []string         `json:"excludedCommands,omitempty"`
	EnableWeakerNestedSandbox bool             `json:"enableWeakerNestedSandbox,omitempty"`
	Network                   *NetworkConfig   `json:"network,omitempty"`
	IgnoreViolations          IgnoreViolations `json:"ignoreViolations,omitempty"`
	Ripgrep                   *RipgrepConfig   `json:"ripgrep,omitempty"`
}

type RipgrepConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

type NetworkConfig struct {
	AllowedDomains      []string `json:"allowedDomains,omitempty"`
	AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
	AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
	AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
	HTTPProxyPort       *int     `json:"httpProxyPort,omitempty"`
	SOCKSProxyPort      *int     `json:"socksProxyPort,omitempty"`
}

type IgnoreViolations map[string][]string

type OutputFormat struct {
	Type   string         `json:"type"`
	Schema map[string]any `json:"schema"`
}

type PluginConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type ToolsConfig struct {
	Tools  []string
	Preset *ToolsPreset
}

type ToolsPreset struct {
	Type   string `json:"type"`
	Preset string `json:"preset"`
}

type SpawnOptions struct {
	Command string
	Args    []string
	Cwd     string
	Env     map[string]*string
	Signal  context.Context
}

type SpawnedProcess interface {
	Stdin() io.WriteCloser
	Stdout() io.ReadCloser
	Killed() bool
	ExitCode() *int
	Kill(signal string) bool
	OnExit(fn func(code *int, signal *string))
	OnError(fn func(err error))
}

type SpawnFunc func(options SpawnOptions) SpawnedProcess

type Option func(*Options)

func WithModel(model string) Option {
	return func(o *Options) {
		o.Model = &model
	}
}

func WithSystemPrompt(prompt string) Option {
	return func(o *Options) {
		o.SystemPrompt = &SystemPromptConfig{Prompt: prompt}
	}
}

func WithSystemPromptPreset(preset, append string) Option {
	return func(o *Options) {
		o.SystemPrompt = &SystemPromptConfig{Preset: preset, Append: append}
	}
}

func WithTools(tools ...string) Option {
	return func(o *Options) {
		o.Tools = &ToolsConfig{Tools: tools}
	}
}

func WithToolsPreset(preset string) Option {
	return func(o *Options) {
		o.Tools = &ToolsConfig{Preset: &ToolsPreset{Type: "preset", Preset: preset}}
	}
}

func WithSpawnClaudeCodeProcess(fn SpawnFunc) Option {
	return func(o *Options) {
		o.SpawnClaudeCodeProcess = fn
	}
}

func WithAllowedTools(tools ...string) Option {
	return func(o *Options) {
		o.AllowedTools = tools
	}
}

func WithDisallowedTools(tools ...string) Option {
	return func(o *Options) {
		o.DisallowedTools = tools
	}
}

func WithPermissionMode(mode control.PermissionMode) Option {
	return func(o *Options) {
		o.PermissionMode = &mode
	}
}

func WithCanUseTool(fn control.CanUseToolFunc) Option {
	return func(o *Options) {
		o.CanUseTool = fn
	}
}

func WithAllowDangerouslySkipPermissions() Option {
	return func(o *Options) {
		o.AllowDangerouslySkipPermissions = true
	}
}

func WithHooks(event control.HookEvent, matchers ...control.HookCallbackMatcher) Option {
	return func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = make(map[control.HookEvent][]control.HookCallbackMatcher)
		}
		o.Hooks[event] = append(o.Hooks[event], matchers...)
	}
}

func WithMcpServers(servers map[string]mcp.ServerConfig) Option {
	return func(o *Options) {
		o.McpServers = servers
	}
}

func WithCwd(cwd string) Option {
	return func(o *Options) {
		o.Cwd = &cwd
	}
}

func WithAdditionalDirectories(dirs ...string) Option {
	return func(o *Options) {
		o.AdditionalDirectories = dirs
	}
}

func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = &sessionID
	}
}

func WithContinue() Option {
	return func(o *Options) {
		o.Continue = true
	}
}

func WithMaxTurns(turns int) Option {
	return func(o *Options) {
		o.MaxTurns = &turns
	}
}

func WithMaxBudgetUSD(budget float64) Option {
	return func(o *Options) {
		o.MaxBudgetUSD = &budget
	}
}

func WithMaxThinkingTokens(tokens int) Option {
	return func(o *Options) {
		o.MaxThinkingTokens = &tokens
	}
}

func WithIncludePartialMessages() Option {
	return func(o *Options) {
		o.IncludePartialMessages = true
	}
}

func WithOutputFormat(format OutputFormat) Option {
	return func(o *Options) {
		o.OutputFormat = &format
	}
}

func WithAgent(agent string) Option {
	return func(o *Options) {
		o.Agent = &agent
	}
}

func WithAgents(agents map[string]AgentDefinition) Option {
	return func(o *Options) {
		o.Agents = agents
	}
}

func WithSandbox(settings SandboxSettings) Option {
	return func(o *Options) {
		o.Sandbox = &settings
	}
}

func WithEnableFileCheckpointing() Option {
	return func(o *Options) {
		o.EnableFileCheckpointing = true
	}
}

func WithPlugins(plugins ...PluginConfig) Option {
	return func(o *Options) {
		o.Plugins = plugins
	}
}

func WithSettingSources(sources ...string) Option {
	return func(o *Options) {
		o.SettingSources = sources
	}
}

func WithBetas(betas ...string) Option {
	return func(o *Options) {
		o.Betas = betas
	}
}

func WithEnv(env map[string]string) Option {
	return func(o *Options) {
		o.Env = env
	}
}

func WithEnvVar(key, value string) Option {
	return func(o *Options) {
		if o.Env == nil {
			o.Env = make(map[string]string)
		}
		o.Env[key] = value
	}
}

func WithCLIPath(path string) Option {
	return func(o *Options) {
		o.CLIPath = &path
	}
}

func WithExecutable(executable string, args ...string) Option {
	return func(o *Options) {
		o.Executable = &executable
		o.ExecutableArgs = args
	}
}

func WithStderr(fn func(string)) Option {
	return func(o *Options) {
		o.Stderr = fn
	}
}

func WithExtraArg(name string, value *string) Option {
	return func(o *Options) {
		if o.ExtraArgs == nil {
			o.ExtraArgs = make(map[string]*string)
		}
		o.ExtraArgs[name] = value
	}
}

func WithForkSession() Option {
	return func(o *Options) {
		o.ForkSession = true
	}
}

func WithPersistSession(persist bool) Option {
	return func(o *Options) {
		o.PersistSession = &persist
	}
}

func WithResumeSessionAt(messageID string) Option {
	return func(o *Options) {
		o.ResumeSessionAt = &messageID
	}
}

func WithStrictMcpConfig() Option {
	return func(o *Options) {
		o.StrictMcpConfig = true
	}
}

func WithPermissionPromptToolName(toolName string) Option {
	return func(o *Options) {
		o.PermissionPromptToolName = &toolName
	}
}

func WithFallbackModel(model string) Option {
	return func(o *Options) {
		o.FallbackModel = &model
	}
}

func applyOptions(opts []Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
