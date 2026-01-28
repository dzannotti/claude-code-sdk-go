// Package main demonstrates an interactive chat with streaming, tools, and conversation memory.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"claudeagent"
)

var debugLog *os.File

func debugf(format string, args ...any) {
	if debugLog != nil {
		fmt.Fprintf(debugLog, "[%s] %s\n", time.Now().Format("15:04:05.000"), fmt.Sprintf(format, args...))
		_ = debugLog.Sync()
	}
}

func main() {
	// Create debug log file
	var err error
	debugLog, err = os.Create("debug.log")
	if err != nil {
		fmt.Printf("Warning: couldn't create debug.log: %v\n", err)
	} else {
		defer debugLog.Close()
		fmt.Println("Debug logging to: debug.log")
	}

	fmt.Println("=== Interactive Chat Example ===")
	fmt.Println("Type your message and press Enter. Type 'quit' to exit.")
	fmt.Println("AI has access to Read/Write/Bash tools and remembers conversation.")

	ctx := context.Background()

	client, err := claudeagent.NewClient(
		claudeagent.WithAllowedTools("Read", "Write", "Bash", "Glob", "Grep"),
		claudeagent.WithSystemPrompt("You are a helpful assistant. When using tools, briefly explain what you're doing."),
		claudeagent.WithIncludePartialMessages(), // Enable token streaming
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

	fmt.Println("Connected to Claude CLI!")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if err := client.Query(ctx, input); err != nil {
			fmt.Printf("Query error: %v\n", err)
			continue
		}

		fmt.Print("\nClaude: ")
		if err := streamResponse(ctx, client); err != nil {
			fmt.Printf("\nStream error: %v\n", err)
		}
		fmt.Print("\n\n")
	}
}

func streamResponse(ctx context.Context, client claudeagent.Client) error {
	msgChan := client.Messages(ctx)

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				debugf("Channel closed")
				return nil
			}

			// Log all message types
			msgJSON, _ := json.Marshal(msg)
			debugf("MSG [%T]: %s", msg, string(msgJSON))

			switch m := msg.(type) {
			case *claudeagent.StreamEvent:
				// Handle streaming events for real-time output
				debugf("  StreamEvent type: %v", m.Event)
				handleStreamEvent(m.Event)

			case *claudeagent.AssistantMessage:
				// Full message (fallback if streaming events aren't sent)
				for _, block := range m.Message.Content {
					switch b := block.(type) {
					case *claudeagent.TextBlock:
						fmt.Print(b.Text)

					case *claudeagent.ThinkingBlock:
						fmt.Printf("\n[Thinking: %s...]\n", truncate(b.Thinking, 50))

					case *claudeagent.ToolUseBlock:
						fmt.Printf("\n[Tool: %s", b.Name)
						if cmd, ok := b.Input["command"].(string); ok {
							fmt.Printf(" → %s", truncate(cmd, 40))
						} else if path, ok := b.Input["file_path"].(string); ok {
							fmt.Printf(" → %s", path)
						} else if pattern, ok := b.Input["pattern"].(string); ok {
							fmt.Printf(" → %s", pattern)
						}
						fmt.Print("]\n")
					}
				}

			case *claudeagent.UserMessage:
				if blocks, ok := m.Message.Content.([]any); ok {
					for _, block := range blocks {
						if blockMap, ok := block.(map[string]any); ok {
							if blockMap["type"] == "tool_result" {
								isError, _ := blockMap["is_error"].(bool)
								content := blockMap["content"]
								if str, ok := content.(string); ok {
									preview := truncate(str, 100)
									if isError {
										fmt.Printf("[Tool Error: %s]\n", preview)
									} else {
										fmt.Printf("[Tool Result: %s]\n", preview)
									}
								}
							}
						}
					}
				}

			case *claudeagent.ResultMessage:
				if m.IsError {
					return fmt.Errorf("error: %s", m.Result)
				}
				return nil

			case *claudeagent.SystemMessage:
				// Ignore system messages (init, etc.)
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func handleStreamEvent(event any) {
	eventMap, ok := event.(map[string]any)
	if !ok {
		return
	}

	eventType, _ := eventMap["type"].(string)

	switch eventType {
	case "content_block_delta":
		// Extract text delta from streaming event
		if delta, ok := eventMap["delta"].(map[string]any); ok {
			if deltaType, _ := delta["type"].(string); deltaType == "text_delta" {
				if text, ok := delta["text"].(string); ok {
					fmt.Print(text) // Print streaming text immediately
				}
			} else if deltaType == "thinking_delta" {
				if thinking, ok := delta["thinking"].(string); ok {
					fmt.Print(thinking) // Print thinking tokens
				}
			}
		}

	case "content_block_start":
		// Optionally handle block start (e.g., for tool use indication)
		if contentBlock, ok := eventMap["content_block"].(map[string]any); ok {
			if blockType, _ := contentBlock["type"].(string); blockType == "tool_use" {
				name, _ := contentBlock["name"].(string)
				fmt.Printf("\n[Tool: %s ...]\n", name)
			}
		}
	}
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
