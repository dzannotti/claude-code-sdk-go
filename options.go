package claudecode

import (
	"claudecode/control"
	"claudecode/mcp"
)

type Options struct {
	Tools                           []string
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
}

type SystemPromptConfig struct {
	Prompt string
	Preset string
	Append string
}

type AgentDefinition struct {
	Description                        string   `json:"description"`
	Tools                              []string `json:"tools,omitempty"`
	DisallowedTools                    []string `json:"disallowedTools,omitempty"`
	Prompt                             string   `json:"prompt"`
	Model                              string   `json:"model,omitempty"`
	CriticalSystemReminderExperimental string   `json:"criticalSystemReminder_EXPERIMENTAL,omitempty"`
}

type SandboxSettings struct {
	Enabled                  bool              `json:"enabled,omitempty"`
	AutoAllowBashIfSandboxed bool              `json:"autoAllowBashIfSandboxed,omitempty"`
	Network                  *NetworkConfig    `json:"network,omitempty"`
	IgnoreViolations         *IgnoreViolations `json:"ignoreViolations,omitempty"`
}

type NetworkConfig struct {
	AllowLocalBinding bool     `json:"allowLocalBinding,omitempty"`
	AllowUnixSockets  []string `json:"allowUnixSockets,omitempty"`
}

type IgnoreViolations struct {
	Paths    []string `json:"paths,omitempty"`
	Networks []string `json:"networks,omitempty"`
}

type OutputFormat struct {
	Type   string         `json:"type"`
	Schema map[string]any `json:"schema"`
}

type PluginConfig struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

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
		o.Tools = tools
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
