package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func FindCLI() (string, error) {
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	locations := getCommonLocations()
	for _, loc := range locations {
		if info, err := os.Stat(loc); err == nil && !info.IsDir() {
			if runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
				continue
			}
			return loc, nil
		}
	}

	if _, err := exec.LookPath("node"); err != nil {
		return "", fmt.Errorf("claude CLI not found and Node.js is not installed; install Node.js from https://nodejs.org/ then run: npm install -g @anthropic-ai/claude-code")
	}

	return "", fmt.Errorf("claude CLI not found; install with: npm install -g @anthropic-ai/claude-code")
}

func getCommonLocations() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	if runtime.GOOS == "windows" {
		return []string{
			filepath.Join(home, "AppData", "Roaming", "npm", "claude.cmd"),
			filepath.Join("C:", "Program Files", "nodejs", "claude.cmd"),
			filepath.Join(home, ".npm-global", "claude.cmd"),
			filepath.Join(home, "node_modules", ".bin", "claude.cmd"),
		}
	}

	return []string{
		filepath.Join(home, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		filepath.Join(home, "node_modules", ".bin", "claude"),
		filepath.Join(home, ".yarn", "bin", "claude"),
		"/opt/homebrew/bin/claude",
		"/usr/local/homebrew/bin/claude",
	}
}

func ValidateWorkingDirectory(cwd string) error {
	if cwd == "" {
		return nil
	}

	info, err := os.Stat(cwd)
	if os.IsNotExist(err) {
		return fmt.Errorf("working directory does not exist: %s", cwd)
	}
	if err != nil {
		return fmt.Errorf("failed to check working directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("working directory is not a directory: %s", cwd)
	}

	return nil
}
