// Package main demonstrates custom permission handling via CanUseTool.
// This lets you programmatically allow, deny, or modify tool calls.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"claudeagent"
	"claudeagent/control"
)

func main() {
	fmt.Println("Claude Code SDK - Permissions Example")
	fmt.Println("This example shows custom tool permission handling")
	fmt.Println()

	ctx := context.Background()

	iterator, err := claudeagent.Query(ctx, "List files in the current directory",
		claudeagent.WithMaxTurns(3),
		claudeagent.WithCanUseTool(permissionHandler),
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

func permissionHandler(
	ctx context.Context,
	toolName string,
	input map[string]any,
	opts control.CanUseToolOptions,
) (control.PermissionResult, error) {
	fmt.Printf("[PERMISSION] Tool: %s\n", toolName)

	// Always allow read operations
	if toolName == "Read" || toolName == "Glob" || toolName == "LS" {
		fmt.Printf("[PERMISSION] Auto-allowing %s\n", toolName)
		return control.PermissionResult{Behavior: control.PermissionAllow}, nil
	}

	// Block any writes to sensitive files
	if toolName == "Write" || toolName == "Edit" {
		if path, ok := input["file_path"].(string); ok {
			if strings.Contains(path, ".env") || strings.Contains(path, "secret") {
				fmt.Printf("[PERMISSION] DENIED: Cannot modify sensitive file %s\n", path)
				return control.PermissionResult{
					Behavior: control.PermissionDeny,
					Message:  "Cannot modify sensitive files",
				}, nil
			}
		}
	}

	// Block dangerous bash commands
	if toolName == "Bash" {
		if cmd, ok := input["command"].(string); ok {
			dangerous := []string{"rm -rf", "sudo", "chmod 777", "> /dev/"}
			for _, d := range dangerous {
				if strings.Contains(cmd, d) {
					fmt.Printf("[PERMISSION] DENIED: Dangerous command blocked\n")
					return control.PermissionResult{
						Behavior: control.PermissionDeny,
						Message:  "Dangerous commands are not allowed",
					}, nil
				}
			}
		}
	}

	// Default: allow other tools
	return control.PermissionResult{Behavior: control.PermissionAllow}, nil
}
