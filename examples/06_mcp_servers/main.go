// Package main demonstrates MCP server configuration.
// MCP (Model Context Protocol) servers provide external tools to Claude.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"claudeagent"
	"claudeagent/mcp"
)

func main() {
	fmt.Println("Claude Code SDK - MCP Servers Example")
	fmt.Println("This example shows how to configure external MCP servers")
	fmt.Println()

	// Define MCP server configurations
	servers := map[string]mcp.ServerConfig{
		// Example: stdio-based MCP server (like mcp-server-time)
		"time-server": mcp.StdioServerConfig{
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "@anthropic/mcp-server-time"},
		},
	}

	ctx := context.Background()

	// Note: This example requires the MCP server to be installed.
	// Run: npm install -g @anthropic/mcp-server-time
	// Or the query will fail gracefully.

	iterator, err := claudeagent.Query(ctx, "What time is it?",
		claudeagent.WithMcpServers(servers),
		claudeagent.WithMaxTurns(3),
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
				if tool, ok := block.(*claudeagent.ToolUseBlock); ok {
					fmt.Printf("\n[Using tool: %s]\n", tool.Name)
				}
			}
		case *claudeagent.ResultMessage:
			fmt.Printf("\n\nCompleted! Cost: $%.4f\n", m.TotalCostUSD)
		}
	}
}
