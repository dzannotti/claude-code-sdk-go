// Package main demonstrates structured output using OutputFormat.
// This constrains Claude's response to match a JSON schema.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"claudeagent"
)

func main() {
	fmt.Println("Claude Code SDK - Structured Output Example")
	fmt.Println("This example requests structured JSON output")
	fmt.Println()

	// Define the output schema
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"title": map[string]any{
				"type":        "string",
				"description": "A descriptive title for the answer",
			},
			"steps": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"step": map[string]any{
							"type":        "integer",
							"description": "Step number",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "What this step does",
						},
					},
					"required": []string{"step", "description"},
				},
			},
			"result": map[string]any{
				"type":        "string",
				"description": "The final answer",
			},
		},
		"required": []string{"title", "steps", "result"},
	}

	ctx := context.Background()

	iterator, err := claudeagent.Query(ctx, "Explain how to calculate 15% of 80",
		claudeagent.WithMaxTurns(1),
		claudeagent.WithOutputFormat(claudeagent.OutputFormat{
			Type:   "json_schema",
			Schema: schema,
		}),
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
					fmt.Println("Structured response:")
					fmt.Println(text.Text)
				}
			}
		case *claudeagent.ResultMessage:
			fmt.Printf("\nCompleted! Cost: $%.4f\n", m.TotalCostUSD)
		}
	}
}
