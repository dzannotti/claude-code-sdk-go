package claudeagent

import (
	"claudeagent/internal/cli"
)

func FindCLI() (string, error) {
	return cli.FindCLI()
}
