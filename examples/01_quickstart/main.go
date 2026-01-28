// Package main demonstrates basic usage of the Claude Code SDK Query API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"claudeagent"
)

func main() {
	fmt.Println("Claude Code SDK - Query API Example")
	fmt.Println("Asking: What is 2+2?")

	ctx := context.Background()

	iterator, err := claudeagent.Query(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\nResponse:")

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudeagent.ErrDone) {
				break
			}
			log.Fatalf("Failed to get message: %v", err)
		}

		switch msg := message.(type) {
		case *claudeagent.AssistantMessage:
			for _, block := range msg.Message.Content {
				if textBlock, ok := block.(*claudeagent.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		case *claudeagent.ResultMessage:
			if msg.IsError {
				log.Printf("Error: %s", msg.Result)
			}
		}
	}

	fmt.Println("\nQuery completed!")
}
