// Package main demonstrates WithClient context manager pattern.
package main

import (
	"context"
	"fmt"
	"log"

	"claudecode"
)

func main() {
	fmt.Println("Claude Code SDK - WithClient Context Manager")
	fmt.Println("Automatic resource management vs manual pattern")

	ctx := context.Background()
	question := "What are the benefits of using context managers in programming?"

	fmt.Println("\n--- WithClient Pattern (Recommended) ---")
	fmt.Println("Automatic connect/disconnect")
	fmt.Println("Guaranteed cleanup on errors")

	if err := demonstrateWithClient(ctx, question); err != nil {
		log.Printf("WithClient failed: %v", err)
	}

	fmt.Println("\n--- Manual Pattern (Still Supported) ---")
	fmt.Println("Manual connect/disconnect required")
	fmt.Println("Easy to forget cleanup")

	if err := demonstrateManualPattern(ctx, question); err != nil {
		log.Printf("Manual pattern failed: %v", err)
	}

	fmt.Println("\n--- Error Handling ---")
	if err := demonstrateErrorScenarios(ctx); err != nil {
		log.Printf("Error demo failed: %v", err)
	}

	fmt.Println("\nRecommendation: Use WithClient for automatic resource management")
}

func demonstrateWithClient(ctx context.Context, question string) error {
	fmt.Println("Using WithClient for automatic resource management...")

	return claudecode.WithClient(ctx, func(client claudecode.Client) error {
		fmt.Println("Connected! Client managed automatically")

		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}

		fmt.Println("\nResponse (first lines):")
		if err := showFirstLines(ctx, client, 3, 80); err != nil {
			return err
		}
		fmt.Println("WithClient will handle cleanup automatically")
		return nil
	})
}

func demonstrateManualPattern(ctx context.Context, question string) error {
	fmt.Println("Using manual Connect/Disconnect pattern...")

	client, err := claudecode.NewClient()
	if err != nil {
		return fmt.Errorf("create client failed: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect failed: %w", err)
	}

	defer func() {
		fmt.Println("Manual cleanup...")
		if err := client.Disconnect(); err != nil {
			log.Printf("Disconnect warning: %v", err)
		}
		fmt.Println("Manual cleanup completed")
	}()

	fmt.Println("Connected manually")

	if err := client.Query(ctx, question); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	fmt.Println("\nResponse (first lines):")
	if err := showFirstLines(ctx, client, 3, 80); err != nil {
		return err
	}
	return nil
}

func demonstrateErrorScenarios(ctx context.Context) error {
	fmt.Println("Testing WithClient error handling...")

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()

	err := claudecode.WithClient(cancelCtx, func(client claudecode.Client) error {
		return client.Query(cancelCtx, "This will be cancelled")
	})
	if err != nil {
		fmt.Printf("WithClient handled cancellation: %v\n", err)
	}

	err = claudecode.WithClient(ctx, func(client claudecode.Client) error {
		return fmt.Errorf("simulated application error")
	})
	if err != nil {
		fmt.Printf("WithClient propagated error: %v\n", err)
		fmt.Println("   Connection was still cleaned up automatically")
	}

	return nil
}

func showFirstLines(ctx context.Context, client claudecode.Client, maxLines, maxWidth int) error {
	msgChan := client.Messages(ctx)
	linesShown := 0

	for linesShown < maxLines {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return nil
			}

			switch m := msg.(type) {
			case *claudecode.AssistantMessage:
				for _, block := range m.Message.Content {
					if textBlock, ok := block.(*claudecode.TextBlock); ok {
						if linesShown < maxLines {
							text := textBlock.Text
							if len(text) > maxWidth {
								text = text[:maxWidth] + "..."
							}
							fmt.Printf("  %s\n", text)
							linesShown++
						}
					}
				}
			case *claudecode.ResultMessage:
				if m.IsError {
					return fmt.Errorf("error: %s", m.Result)
				}
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	drainMessages(msgChan)
	return nil
}

func drainMessages(msgChan <-chan claudecode.Message) {
	for {
		select {
		case msg := <-msgChan:
			if msg == nil {
				return
			}
		default:
			return
		}
	}
}
