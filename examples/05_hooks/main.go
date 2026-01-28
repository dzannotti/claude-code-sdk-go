// Package main demonstrates hook callbacks for intercepting tool usage.
// Hooks let you observe or modify Claude's behavior during execution.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"claudeagent"
	"claudeagent/control"
)

func main() {
	fmt.Println("Claude Code SDK - Hooks Example")
	fmt.Println("This example logs tool usage via pre/post tool hooks")
	fmt.Println()

	ctx := context.Background()

	iterator, err := claudeagent.Query(ctx, "Read the go.mod file and tell me what the module name is",
		claudeagent.WithMaxTurns(3),
		claudeagent.WithHooks(control.HookPreToolUse, control.HookCallbackMatcher{
			Hooks: []control.HookCallback{preToolUseHook},
		}),
		claudeagent.WithHooks(control.HookPostToolUse, control.HookCallbackMatcher{
			Hooks: []control.HookCallback{postToolUseHook},
		}),
	)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer iterator.Close()

	for {
		msg, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudeagent.ErrDone) {
				break
			}
			log.Fatalf("Error: %v", err)
		}

		switch m := msg.(type) {
		case *claudeagent.AssistantMessage:
			for _, block := range m.Message.Content {
				if text, ok := block.(*claudeagent.TextBlock); ok {
					fmt.Print(text.Text)
				}
			}
		case *claudeagent.ResultMessage:
			fmt.Printf("\n\nCompleted! Cost: $%.4f\n", m.TotalCostUSD)
		}
	}
}

func preToolUseHook(ctx context.Context, input control.HookInput, toolUseID *string) (control.HookOutput, error) {
	if pre, ok := input.(*control.PreToolUseHookInput); ok {
		fmt.Printf("\n[HOOK] Pre-tool: %s (id: %s)\n", pre.ToolName, pre.ToolUseID)
	}
	return control.HookOutput{}, nil
}

func postToolUseHook(ctx context.Context, input control.HookInput, toolUseID *string) (control.HookOutput, error) {
	if post, ok := input.(*control.PostToolUseHookInput); ok {
		fmt.Printf("[HOOK] Post-tool: %s completed\n\n", post.ToolName)
	}
	return control.HookOutput{}, nil
}
