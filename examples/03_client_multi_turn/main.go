// Package main demonstrates multi-turn conversation with context preservation using WithClient.
package main

import (
	"context"
	"fmt"
	"log"

	"claudeagent"
)

func main() {
	fmt.Println("Claude Code SDK - Multi-Turn Conversation Example")
	fmt.Println("Building context across multiple related questions")

	ctx := context.Background()

	questions := []string{
		"What is a binary search tree?",
		"Can you show me a Go implementation of inserting a node?",
		"What would be the time complexity of that insertion?",
		"How would I implement a search function for the same tree?",
	}

	err := claudeagent.WithClient(ctx, func(client claudeagent.Client) error {
		fmt.Println("\nConnected! Starting conversation...")

		for i, question := range questions {
			fmt.Printf("\n--- Turn %d ---\n", i+1)
			fmt.Printf("Q: %s\n\n", question)

			if err := client.Query(ctx, question); err != nil {
				return fmt.Errorf("turn %d failed: %w", i+1, err)
			}

			if err := streamFullResponse(ctx, client); err != nil {
				return fmt.Errorf("turn %d streaming failed: %w", i+1, err)
			}
		}

		fmt.Println("\n\nConversation completed!")
		fmt.Println("Notice: Each question built on previous responses automatically")
		return nil
	})
	if err != nil {
		log.Fatalf("Conversation failed: %v", err)
	}
}

func streamFullResponse(ctx context.Context, client claudeagent.Client) error {
	msgChan := client.Messages(ctx)
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			switch m := msg.(type) {
			case *claudeagent.AssistantMessage:
				for _, block := range m.Message.Content {
					if textBlock, ok := block.(*claudeagent.TextBlock); ok {
						fmt.Print(textBlock.Text)
					}
				}
			case *claudeagent.ResultMessage:
				if m.IsError {
					return fmt.Errorf("error: %s", m.Result)
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
