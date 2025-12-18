package claudecode

import (
	"context"
	"fmt"

	"claudecode/internal/cli"
	"claudecode/internal/transport"
	"claudecode/message"
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
		MaxTurns:                        options.MaxTurns,
		MaxBudgetUSD:                    options.MaxBudgetUSD,
		Cwd:                             options.Cwd,
		AdditionalDirectories:           options.AdditionalDirectories,
		Betas:                           options.Betas,
		ExtraArgs:                       options.ExtraArgs,
		SettingSources:                  options.SettingSources,
		AllowDangerouslySkipPermissions: options.AllowDangerouslySkipPermissions,
		IncludePartialMessages:          options.IncludePartialMessages,
		Model:                           options.Model,
		MaxThinkingTokens:               options.MaxThinkingTokens,
		PermissionMode:                  options.PermissionMode,
		PermissionPromptToolName:        options.PermissionPromptToolName,
	}

	if options.SystemPrompt != nil {
		if options.SystemPrompt.Prompt != "" {
			cmdOpts.SystemPrompt = &options.SystemPrompt.Prompt
		}
		if options.SystemPrompt.Append != "" {
			cmdOpts.AppendSystemPrompt = &options.SystemPrompt.Append
		}
	}

	return cmdOpts
}
