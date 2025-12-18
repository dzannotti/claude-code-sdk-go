// Package claudecode provides a Go SDK for programmatic control of Claude Code CLI.
//
// The SDK offers two main APIs:
//
//   - Query API: One-shot queries with automatic resource cleanup
//   - Client API: Bidirectional streaming for interactive sessions
//
// Basic usage with Query API:
//
//	iter, err := claudecode.Query(ctx, "What is 2+2?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer iter.Close()
//
//	for {
//	    msg, err := iter.Next(ctx)
//	    if errors.Is(err, claudecode.ErrDone) {
//	        break
//	    }
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    // Handle message
//	}
//
// Interactive usage with Client API:
//
//	err := claudecode.WithClient(ctx, func(c claudecode.Client) error {
//	    if err := c.Query(ctx, "Hello"); err != nil {
//	        return err
//	    }
//	    for msg := range c.Messages(ctx) {
//	        // Handle message
//	    }
//	    return nil
//	})
package claudecode
