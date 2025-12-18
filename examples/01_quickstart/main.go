// Package main demonstrates basic usage of the Claude Code SDK Query API.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"claudecode"
)

func main() {
	fmt.Println("Claude Code SDK - Query API Example")
	fmt.Println("Asking: What is 2+2?")

	ctx := context.Background()

	iterator, err := claudecode.Query(ctx, "What is 2+2?")
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	defer iterator.Close()

	fmt.Println("\nResponse:")

	for {
		message, err := iterator.Next(ctx)
		if err != nil {
			if errors.Is(err, claudecode.ErrDone) {
				break
			}
			log.Fatalf("Failed to get message: %v", err)
		}

		switch msg := message.(type) {
		case *claudecode.AssistantMessage:
			for _, block := range msg.Message.Content {
				if textBlock, ok := block.(*claudecode.TextBlock); ok {
					fmt.Print(textBlock.Text)
				}
			}
		case *claudecode.ResultMessage:
			if msg.IsError {
				log.Printf("Error: %s", msg.Result)
			}
		}
	}

	fmt.Println("\nQuery completed!")
}
