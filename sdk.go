package claudeagent

import (
	"github.com/dzannotti/claude-code-sdk-go/internal/cli"
)

func FindCLI() (string, error) {
	return cli.FindCLI()
}
