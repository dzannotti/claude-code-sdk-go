package claudecode

import (
	"context"
	"fmt"
	"sync"

	"claudecode/control"
	"claudecode/internal/transport"
	"claudecode/mcp"
	"claudecode/message"
)

type Client interface {
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	Query(ctx context.Context, prompt string) error
	QueryWithSession(ctx context.Context, prompt string, sessionID string) error

	Messages(ctx context.Context) <-chan message.Message
	Errors(ctx context.Context) <-chan error

	Interrupt(ctx context.Context) error
	SetPermissionMode(ctx context.Context, mode PermissionMode) error
	SetModel(ctx context.Context, model string) error
	RewindFiles(ctx context.Context, userMessageID string) error

	SetMcpServers(ctx context.Context, servers map[string]mcp.ServerConfig) (*mcp.SetServersResult, error)
	McpServerStatus(ctx context.Context) ([]mcp.ServerStatus, error)
	SetMaxThinkingTokens(ctx context.Context, tokens *int) error

	SupportedCommands(ctx context.Context) ([]SlashCommand, error)
	SupportedModels(ctx context.Context) ([]ModelInfo, error)
	AccountInfo(ctx context.Context) (*AccountInfo, error)

	SessionID() string
}

type clientImpl struct {
	transport    *transport.SubprocessTransport
	options      *Options
	cliPath      string
	sessionID    string
	initResponse *initResponse
	mu           sync.RWMutex
}

type initResponse struct {
	commands []SlashCommand
	models   []ModelInfo
	account  AccountInfo
}

func NewClient(opts ...Option) (Client, error) {
	options := applyOptions(opts)

	cliPath, err := resolveCLIPath(options)
	if err != nil {
		return nil, err
	}

	return &clientImpl{
		options: options,
		cliPath: cliPath,
	}, nil
}

func WithClient(ctx context.Context, fn func(Client) error, opts ...Option) error {
	client, err := NewClient(opts...)
	if err != nil {
		return err
	}

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer func() { _ = client.Disconnect() }()

	return fn(client)
}

func (c *clientImpl) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil && c.transport.IsConnected() {
		return fmt.Errorf("client already connected")
	}

	cmdOpts := buildCommandOptions(c.options)
	var opts []transport.SubprocessOption
	if c.options.Env != nil {
		opts = append(opts, transport.WithEnv(c.options.Env))
	}

	t := transport.NewSubprocessTransport(c.cliPath, cmdOpts, opts...)

	if c.options.CanUseTool != nil {
		t.Control().SetCanUseTool(c.options.CanUseTool)
	}

	if c.options.Hooks != nil {
		hookMatchers := make(map[control.HookEvent][]control.HookCallbackMatcher)
		for event, matchers := range c.options.Hooks {
			hookMatchers[event] = matchers
		}
		t.Control().SetHooks(hookMatchers)
	}

	if err := t.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.transport = t

	var jsonSchema map[string]any
	if c.options.OutputFormat != nil {
		jsonSchema = c.options.OutputFormat.Schema
	}

	var agents map[string]any
	if len(c.options.Agents) > 0 {
		agents = make(map[string]any)
		for name, def := range c.options.Agents {
			agents[name] = def
		}
	}

	initResp, err := t.Control().Initialize(ctx, nil, nil, jsonSchema, nil, nil, agents)
	if err != nil {
		t.Close()
		c.transport = nil
		return fmt.Errorf("failed to initialize: %w", err)
	}

	c.initResponse = &initResponse{
		commands: make([]SlashCommand, len(initResp.Commands)),
		models:   make([]ModelInfo, len(initResp.Models)),
		account: AccountInfo{
			Email:            initResp.Account.Email,
			Organization:     initResp.Account.Organization,
			SubscriptionType: initResp.Account.SubscriptionType,
			TokenSource:      initResp.Account.TokenSource,
			APIKeySource:     initResp.Account.APIKeySource,
		},
	}

	for i, cmd := range initResp.Commands {
		c.initResponse.commands[i] = SlashCommand{
			Name:         cmd.Name,
			Description:  cmd.Description,
			ArgumentHint: cmd.ArgumentHint,
		}
	}

	for i, model := range initResp.Models {
		c.initResponse.models[i] = ModelInfo{
			Value:       model.Value,
			DisplayName: model.DisplayName,
			Description: model.Description,
		}
	}

	return nil
}

func (c *clientImpl) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport == nil {
		return nil
	}

	err := c.transport.Close()
	c.transport = nil
	return err
}

func (c *clientImpl) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.transport != nil && c.transport.IsConnected()
}

func (c *clientImpl) Query(ctx context.Context, prompt string) error {
	return c.QueryWithSession(ctx, prompt, "default")
}

func (c *clientImpl) QueryWithSession(ctx context.Context, prompt string, sessionID string) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	msg := transport.StreamMessage{
		Type: "user",
		Message: message.UserContent{
			Role:    "user",
			Content: prompt,
		},
		SessionID: sessionID,
	}

	return t.SendMessage(ctx, msg)
}

func (c *clientImpl) Messages(ctx context.Context) <-chan message.Message {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil {
		ch := make(chan message.Message)
		close(ch)
		return ch
	}

	msgChan, _ := t.ReceiveMessages(ctx)
	return msgChan
}

func (c *clientImpl) Errors(ctx context.Context) <-chan error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil {
		ch := make(chan error)
		close(ch)
		return ch
	}

	_, errChan := t.ReceiveMessages(ctx)
	return errChan
}

func (c *clientImpl) Interrupt(ctx context.Context) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	return t.Control().Interrupt(ctx)
}

func (c *clientImpl) SetPermissionMode(ctx context.Context, mode PermissionMode) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	return t.Control().SetPermissionMode(ctx, mode)
}

func (c *clientImpl) SetModel(ctx context.Context, model string) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	return t.Control().SetModel(ctx, &model)
}

func (c *clientImpl) SetMaxThinkingTokens(ctx context.Context, tokens *int) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	return t.Control().SetMaxThinkingTokens(ctx, tokens)
}

func (c *clientImpl) RewindFiles(ctx context.Context, userMessageID string) error {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return ErrNotConnected
	}

	return t.Control().RewindFiles(ctx, userMessageID)
}

func (c *clientImpl) SetMcpServers(ctx context.Context, servers map[string]mcp.ServerConfig) (*mcp.SetServersResult, error) {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return nil, ErrNotConnected
	}

	return t.Control().SetMcpServers(ctx, servers)
}

func (c *clientImpl) McpServerStatus(ctx context.Context) ([]mcp.ServerStatus, error) {
	c.mu.RLock()
	t := c.transport
	c.mu.RUnlock()

	if t == nil || !t.IsConnected() {
		return nil, ErrNotConnected
	}

	return t.Control().McpStatus(ctx)
}

func (c *clientImpl) SupportedCommands(ctx context.Context) ([]SlashCommand, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.initResponse == nil {
		return nil, ErrNotConnected
	}
	return c.initResponse.commands, nil
}

func (c *clientImpl) SupportedModels(ctx context.Context) ([]ModelInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.initResponse == nil {
		return nil, ErrNotConnected
	}
	return c.initResponse.models, nil
}

func (c *clientImpl) AccountInfo(ctx context.Context) (*AccountInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.initResponse == nil {
		return nil, ErrNotConnected
	}
	return &c.initResponse.account, nil
}

func (c *clientImpl) SessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}
