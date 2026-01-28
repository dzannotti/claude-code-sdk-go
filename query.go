package claudeagent

import (
	"context"
	"fmt"

	"claudeagent/internal/cli"
	"claudeagent/internal/transport"
	"claudeagent/message"
)

func Query(ctx context.Context, prompt string, opts ...Option) (MessageIterator, error) {
	options := applyOptions(opts)

	cliPath, err := resolveCLIPath(options)
	if err != nil {
		return nil, err
	}

	cmdOpts := buildCommandOptions(options)
	t := transport.NewSubprocessTransport(cliPath, cmdOpts, transport.WithPrompt(prompt))

	if err := t.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	msgChan, errChan := t.ReceiveMessages(ctx)

	return newChannelIterator(msgChan, errChan, t.Close), nil
}

func QueryWithInput(ctx context.Context, input <-chan message.UserMessage, opts ...Option) (MessageIterator, error) {
	options := applyOptions(opts)

	cliPath, err := resolveCLIPath(options)
	if err != nil {
		return nil, err
	}

	cmdOpts := buildCommandOptions(options)
	t := transport.NewSubprocessTransport(cliPath, cmdOpts)

	if err := t.Connect(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	go func() {
		for msg := range input {
			streamMsg := transport.StreamMessage{
				Type: "user",
				Message: message.UserContent{
					Role:    "user",
					Content: msg.Message.Content,
				},
				SessionID: msg.SessionID,
			}
			if err := t.SendMessage(ctx, streamMsg); err != nil {
				break
			}
		}
	}()

	msgChan, errChan := t.ReceiveMessages(ctx)
	return newChannelIterator(msgChan, errChan, t.Close), nil
}

func resolveCLIPath(options *Options) (string, error) {
	if options.CLIPath != nil {
		return *options.CLIPath, nil
	}
	return cli.FindCLI()
}

func buildCommandOptions(options *Options) *cli.CommandOptions {
	cmdOpts := &cli.CommandOptions{
		AllowedTools:                    options.AllowedTools,
		DisallowedTools:                 options.DisallowedTools,
		Continue:                        options.Continue,
		Resume:                          options.Resume,
		ResumeSessionAt:                 options.ResumeSessionAt,
		ForkSession:                     options.ForkSession,
		PersistSession:                  options.PersistSession,
		MaxTurns:                        options.MaxTurns,
		MaxBudgetUSD:                    options.MaxBudgetUSD,
		Cwd:                             options.Cwd,
		AdditionalDirectories:           options.AdditionalDirectories,
		McpServers:                      options.McpServers,
		StrictMcpConfig:                 options.StrictMcpConfig,
		Agent:                           options.Agent,
		EnableFileCheckpointing:         options.EnableFileCheckpointing,
		Betas:                           options.Betas,
		ExtraArgs:                       options.ExtraArgs,
		SettingSources:                  options.SettingSources,
		AllowDangerouslySkipPermissions: options.AllowDangerouslySkipPermissions,
		IncludePartialMessages:          options.IncludePartialMessages,
		Model:                           options.Model,
		FallbackModel:                   options.FallbackModel,
		MaxThinkingTokens:               options.MaxThinkingTokens,
		PermissionMode:                  options.PermissionMode,
		PermissionPromptToolName:        options.PermissionPromptToolName,
	}

	// When CanUseTool callback is set, tell CLI to send permission prompts via stdio
	if options.CanUseTool != nil && options.PermissionPromptToolName == nil {
		stdio := "stdio"
		cmdOpts.PermissionPromptToolName = &stdio
	}

	if options.SystemPrompt != nil {
		if options.SystemPrompt.Prompt != "" {
			cmdOpts.SystemPrompt = &options.SystemPrompt.Prompt
		}
		if options.SystemPrompt.Append != "" {
			cmdOpts.AppendSystemPrompt = &options.SystemPrompt.Append
		}
	}

	if options.Tools != nil {
		cmdOpts.Tools = options.Tools
	}
	if options.Sandbox != nil {
		cmdOpts.Sandbox = options.Sandbox
	}
	if options.Plugins != nil {
		cmdOpts.Plugins = options.Plugins
	}
	if options.OutputFormat != nil {
		cmdOpts.OutputFormat = options.OutputFormat
	}

	return cmdOpts
}
