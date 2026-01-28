// Package main demonstrates streaming with Client API using automatic resource management.
package main

import (
	"context"
	"fmt"
	"log"

	"claudeagent"
)

func main() {
	fmt.Println("Claude Code SDK - Client Streaming Example")
	fmt.Println("Asking: Explain Go goroutines with a simple example")

	ctx := context.Background()
	question := "Explain what Go goroutines are and show a simple example"

	err := claudeagent.WithClient(ctx, func(client claudeagent.Client) error {
		fmt.Println("\nConnected! Streaming response:")

		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

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
	})
	if err != nil {
		log.Fatalf("Streaming failed: %v", err)
	}

	fmt.Println("\n\nStreaming completed!")
}
