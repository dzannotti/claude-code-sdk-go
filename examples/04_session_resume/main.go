// Package main demonstrates session resume with history loading.
// This example shows how to load conversation history from a previous session
// and continue the conversation.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"claudeagent"
	"claudeagent/message"
	"claudeagent/session"
)

func main() {
	if len(os.Args) < 2 {
		listSessions()
		return
	}

	sessionID := os.Args[1]
	if len(os.Args) > 2 && os.Args[2] == "continue" {
		continueSession(sessionID)
	} else {
		showHistory(sessionID)
	}
}

func listSessions() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	projectDir, err := session.ProjectDir(cwd)
	if err != nil {
		fmt.Println("No sessions found for current directory")
		fmt.Println("Run some claude queries first to create sessions")
		return
	}

	sessions, err := session.ListSessions(projectDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Available sessions (most recent first):")
	fmt.Println()
	for _, s := range sessions {
		fmt.Printf("  %s  (%s, %d bytes)\n", s.ID, s.ModTime.Format("2006-01-02 15:04"), s.SizeBytes)
	}
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  go run . <session-id>           # Show history")
	fmt.Println("  go run . <session-id> continue  # Continue conversation")
}

func showHistory(sessionID string) {
	cwd, _ := os.Getwd()
	projectDir, err := session.ProjectDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := session.LoadByID(projectDir, sessionID)
	if err != nil {
		log.Fatalf("Failed to load session: %v", err)
	}

	fmt.Printf("Session: %s\n", sessionID)
	fmt.Printf("Messages: %d\n\n", len(msgs))

	for _, msg := range msgs {
		printMessage(msg)
	}
}

func continueSession(sessionID string) {
	cwd, _ := os.Getwd()
	projectDir, err := session.ProjectDir(cwd)
	if err != nil {
		log.Fatal(err)
	}

	// Load and display history first
	msgs, err := session.LoadByID(projectDir, sessionID)
	if err != nil {
		log.Fatalf("Failed to load session: %v", err)
	}

	fmt.Println("=== Previous conversation ===")
	for _, msg := range msgs {
		printMessage(msg)
	}
	fmt.Println("=== Continuing... ===")

	// Resume the session
	ctx := context.Background()
	iter, err := claudeagent.Query(ctx, "What were we just talking about? Give a brief summary.",
		claudeagent.WithResume(sessionID),
		claudeagent.WithMaxTurns(1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Close()

	for {
		msg, err := iter.Next(ctx)
		if errors.Is(err, claudeagent.ErrDone) {
			break
		}
		if err != nil {
			log.Printf("Error: %v", err)
			break
		}
		printMessage(msg)
	}
}

func printMessage(msg message.Message) {
	switch m := msg.(type) {
	case *message.UserMessage:
		content := ""
		if s, ok := m.Message.Content.(string); ok {
			content = s
		}
		fmt.Printf("USER: %s\n\n", truncate(content, 200))

	case *message.AssistantMessage:
		for _, block := range m.Message.Content {
			if text, ok := block.(*message.TextBlock); ok {
				fmt.Printf("ASSISTANT: %s\n\n", truncate(text.Text, 500))
			}
			if tool, ok := block.(*message.ToolUseBlock); ok {
				fmt.Printf("TOOL USE: %s\n\n", tool.Name)
			}
		}

	case *message.ResultMessage:
		fmt.Printf("[Session complete - Cost: $%.4f]\n", m.TotalCostUSD)
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
