package cli

import (
	"strings"
	"testing"

	"claudecode/control"
)

func TestBuildCommand_Basic(t *testing.T) {
	cmd := BuildCommand("/usr/bin/claude", nil, false)

	if cmd[0] != "/usr/bin/claude" {
		t.Errorf("expected /usr/bin/claude, got %s", cmd[0])
	}

	hasOutputFormat := false
	hasVerbose := false
	hasInputFormat := false
	for i, arg := range cmd {
		if arg == "--output-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
			hasOutputFormat = true
		}
		if arg == "--verbose" {
			hasVerbose = true
		}
		if arg == "--input-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
			hasInputFormat = true
		}
	}

	if !hasOutputFormat {
		t.Error("expected --output-format stream-json")
	}
	if !hasVerbose {
		t.Error("expected --verbose")
	}
	if !hasInputFormat {
		t.Error("expected --input-format stream-json for streaming mode")
	}
}

func TestBuildCommand_OneShot(t *testing.T) {
	cmd := BuildCommand("/usr/bin/claude", nil, true)

	hasPrint := false
	hasInputFormat := false
	for _, arg := range cmd {
		if arg == "--print" {
			hasPrint = true
		}
		if arg == "--input-format" {
			hasInputFormat = true
		}
	}

	if !hasPrint {
		t.Error("expected --print for one-shot mode")
	}
	if hasInputFormat {
		t.Error("--input-format should not be present in one-shot mode")
	}
}

func TestBuildCommand_WithOptions(t *testing.T) {
	model := "claude-3"
	mode := control.PermissionModeAcceptEdits
	maxTurns := 5
	systemPrompt := "You are helpful"

	opts := &CommandOptions{
		Model:          &model,
		PermissionMode: &mode,
		MaxTurns:       &maxTurns,
		SystemPrompt:   &systemPrompt,
		AllowedTools:   []string{"Bash", "Read"},
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--model claude-3") {
		t.Error("expected --model claude-3")
	}
	if !strings.Contains(cmdStr, "--permission-mode acceptEdits") {
		t.Error("expected --permission-mode acceptEdits")
	}
	if !strings.Contains(cmdStr, "--max-turns 5") {
		t.Error("expected --max-turns 5")
	}
	if !strings.Contains(cmdStr, "--system-prompt") {
		t.Error("expected --system-prompt")
	}
	if !strings.Contains(cmdStr, "--allowed-tools Bash,Read") {
		t.Error("expected --allowed-tools Bash,Read")
	}
}

func TestBuildCommand_WithExtraArgs(t *testing.T) {
	flagValue := "value"
	opts := &CommandOptions{
		ExtraArgs: map[string]*string{
			"custom-flag": &flagValue,
			"bool-flag":   nil,
		},
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--custom-flag value") {
		t.Error("expected --custom-flag value")
	}
	if !strings.Contains(cmdStr, "--bool-flag") {
		t.Error("expected --bool-flag")
	}
}

func TestBuildCommand_WithCwd(t *testing.T) {
	cwd := "/home/user/project"
	opts := &CommandOptions{
		Cwd: &cwd,
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--cwd /home/user/project") {
		t.Error("expected --cwd /home/user/project")
	}
}

func TestBuildCommand_WithAdditionalDirectories(t *testing.T) {
	opts := &CommandOptions{
		AdditionalDirectories: []string{"/tmp", "/var"},
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--add-dir /tmp") {
		t.Error("expected --add-dir /tmp")
	}
	if !strings.Contains(cmdStr, "--add-dir /var") {
		t.Error("expected --add-dir /var")
	}
}

func TestBuildCommand_Continue(t *testing.T) {
	opts := &CommandOptions{
		Continue: true,
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--continue") {
		t.Error("expected --continue")
	}
}

func TestBuildCommand_Resume(t *testing.T) {
	sessionID := "session-123"
	opts := &CommandOptions{
		Resume: &sessionID,
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--resume session-123") {
		t.Error("expected --resume session-123")
	}
}

func TestBuildCommandWithPrompt(t *testing.T) {
	cmd := BuildCommandWithPrompt("/usr/bin/claude", nil, "Hello world")

	if cmd[0] != "/usr/bin/claude" {
		t.Errorf("expected /usr/bin/claude, got %s", cmd[0])
	}

	foundPrompt := false
	foundPrint := false
	for i, arg := range cmd {
		if arg == "--print" {
			foundPrint = true
			if i+1 < len(cmd) && cmd[i+1] == "Hello world" {
				foundPrompt = true
			}
		}
	}

	if !foundPrint {
		t.Error("expected --print")
	}
	if !foundPrompt {
		t.Error("expected prompt after --print")
	}
}

func TestBuildCommand_DangerouslySkipPermissions(t *testing.T) {
	opts := &CommandOptions{
		AllowDangerouslySkipPermissions: true,
	}

	cmd := BuildCommand("/usr/bin/claude", opts, false)
	cmdStr := strings.Join(cmd, " ")

	if !strings.Contains(cmdStr, "--dangerously-skip-permissions") {
		t.Error("expected --dangerously-skip-permissions")
	}
}
