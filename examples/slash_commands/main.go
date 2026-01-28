// Package main demonstrates discovering and executing custom slash commands.
package main

import (
	"context"
	"fmt"
	"os"

	"claudeagent"
)

func main() {
	fmt.Println("=== Slash Commands Example ===")

	ctx := context.Background()

	// Create client with project settings loaded (where custom commands live)
	// Without WithSettingSources, no custom commands are loaded by default
	client, err := claudeagent.NewClient(
		claudeagent.WithSettingSources("user", "project"), // Load user + project settings
	)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		os.Exit(1)
	}

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect() }()

	// List available slash commands from initialization
	commands, err := client.SupportedCommands(ctx)
	if err != nil {
		fmt.Printf("Failed to get commands: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available Slash Commands:")
	fmt.Println("─────────────────────────")
	if len(commands) == 0 {
		fmt.Println("  (none found - check .claude/commands/ in your project)")
	}
	for _, cmd := range commands {
		fmt.Printf("  /%s", cmd.Name)
		if cmd.ArgumentHint != "" {
			fmt.Printf(" %s", cmd.ArgumentHint)
		}
		fmt.Println()
		if cmd.Description != "" {
			fmt.Printf("    %s\n", cmd.Description)
		}
	}

	// Also show account info to verify connection
	account, err := client.AccountInfo(ctx)
	if err == nil && account.Email != nil {
		fmt.Printf("\nConnected as: %s\n", *account.Email)
	}

	// Example: Execute a slash command (if one exists)
	// Slash commands are just prompts prefixed with /
	if len(commands) > 0 {
		fmt.Printf("\n─────────────────────────\n")
		fmt.Printf("To execute a command, just Query it:\n")
		fmt.Printf("  client.Query(ctx, \"/%s\")\n", commands[0].Name)
	}

	// Demo: if you want to actually execute one, uncomment:
	// if err := client.Query(ctx, "/help"); err != nil {
	// 	fmt.Printf("Query error: %v\n", err)
	// }
	// for msg := range client.Messages(ctx) {
	// 	// handle response...
	// }

	fmt.Println("\nDone!")
}
