package transport

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"claudeagent/internal/cli"
	"claudeagent/internal/parser"
	"claudeagent/internal/protocol"
	"claudeagent/message"
)

const (
	channelBufferSize         = 10
	terminationTimeoutSeconds = 5
)

type SubprocessTransport struct {
	cliPath    string
	cmdOpts    *cli.CommandOptions
	closeStdin bool
	promptArg  *string
	entrypoint string

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr *os.File

	parser  *parser.Parser
	control *protocol.ControlHandler

	msgChan chan message.Message
	errChan chan error

	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	connected bool

	env map[string]string
	cwd *string
}

type SubprocessOption func(*SubprocessTransport)

func WithPrompt(prompt string) SubprocessOption {
	return func(t *SubprocessTransport) {
		t.promptArg = &prompt
		t.closeStdin = true
		t.entrypoint = "sdk-go"
	}
}

func WithCloseStdin(close bool) SubprocessOption {
	return func(t *SubprocessTransport) {
		t.closeStdin = close
	}
}

func WithEnv(env map[string]string) SubprocessOption {
	return func(t *SubprocessTransport) {
		t.env = env
	}
}

func WithWorkingDirectory(cwd string) SubprocessOption {
	return func(t *SubprocessTransport) {
		t.cwd = &cwd
	}
}

func NewSubprocessTransport(cliPath string, cmdOpts *cli.CommandOptions, opts ...SubprocessOption) *SubprocessTransport {
	t := &SubprocessTransport{
		cliPath:    cliPath,
		cmdOpts:    cmdOpts,
		closeStdin: false,
		entrypoint: "sdk-go-client",
		parser:     parser.New(),
	}

	for _, opt := range opts {
		opt(t)
	}

	t.control = protocol.NewControlHandler(t.sendRaw)

	return t
}

func (t *SubprocessTransport) sendRaw(ctx context.Context, data []byte) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected || t.stdin == nil {
		return fmt.Errorf("transport not connected or stdin closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

func (t *SubprocessTransport) Control() *protocol.ControlHandler {
	return t.control
}

func (t *SubprocessTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected && t.cmd != nil && t.cmd.Process != nil
}

func (t *SubprocessTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return fmt.Errorf("transport already connected")
	}

	var args []string
	if t.promptArg != nil {
		args = cli.BuildCommandWithPrompt(t.cliPath, t.cmdOpts, *t.promptArg)
	} else {
		args = cli.BuildCommand(t.cliPath, t.cmdOpts, t.closeStdin)
	}

	t.cmd = exec.CommandContext(ctx, args[0], args[1:]...)

	env := os.Environ()
	env = append(env, "CLAUDE_CODE_ENTRYPOINT="+t.entrypoint)
	for k, v := range t.env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	t.cmd.Env = env

	if t.cwd != nil {
		if err := cli.ValidateWorkingDirectory(*t.cwd); err != nil {
			return err
		}
		t.cmd.Dir = *t.cwd
	} else if t.cmdOpts != nil && t.cmdOpts.Cwd != nil {
		if err := cli.ValidateWorkingDirectory(*t.cmdOpts.Cwd); err != nil {
			return err
		}
		t.cmd.Dir = *t.cmdOpts.Cwd
	}

	var err error
	if t.promptArg == nil {
		t.stdin, err = t.cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	t.stderr, err = os.CreateTemp("", "claude_stderr_*.log")
	if err != nil {
		return fmt.Errorf("failed to create stderr file: %w", err)
	}
	t.cmd.Stderr = t.stderr

	if err := t.cmd.Start(); err != nil {
		t.cleanup()
		return fmt.Errorf("failed to start CLI: %w", err)
	}

	t.ctx, t.cancel = context.WithCancel(ctx)
	t.msgChan = make(chan message.Message, channelBufferSize)
	t.errChan = make(chan error, channelBufferSize)

	t.wg.Add(1)
	go t.handleStdout()

	t.connected = true
	return nil
}

func (t *SubprocessTransport) SendMessage(ctx context.Context, msg StreamMessage) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.promptArg != nil {
		return nil
	}

	if !t.connected || t.stdin == nil {
		return fmt.Errorf("transport not connected or stdin closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if _, err := t.stdin.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if t.closeStdin {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	return nil
}

func (t *SubprocessTransport) ReceiveMessages(_ context.Context) (<-chan message.Message, <-chan error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		msgChan := make(chan message.Message)
		errChan := make(chan error)
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	return t.msgChan, t.errChan
}

func (t *SubprocessTransport) Interrupt(_ context.Context) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected || t.cmd == nil || t.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("interrupt not supported on windows")
	}

	return t.cmd.Process.Signal(os.Interrupt)
}

func (t *SubprocessTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false

	if t.cancel != nil {
		t.cancel()
	}

	if t.stdin != nil {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(terminationTimeoutSeconds * time.Second):
	}

	var err error
	if t.cmd != nil && t.cmd.Process != nil {
		err = t.terminateProcess()
	}

	t.cleanup()
	return err
}

func (t *SubprocessTransport) handleStdout() {
	defer t.wg.Done()
	defer close(t.msgChan)
	defer close(t.errChan)

	scanner := bufio.NewScanner(t.stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		if t.isControlMessage(line) {
			resp, err := t.control.HandleIncoming(t.ctx, []byte(line))
			if err != nil {
				select {
				case t.errChan <- err:
				case <-t.ctx.Done():
					return
				}
			}
			if resp != nil {
				_ = t.sendRaw(t.ctx, resp)
			}
			continue
		}

		messages, err := t.parser.ProcessLine(line)
		if err != nil {
			select {
			case t.errChan <- err:
			case <-t.ctx.Done():
				return
			}
			continue
		}

		for _, msg := range messages {
			if msg != nil {
				select {
				case t.msgChan <- msg:
				case <-t.ctx.Done():
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errChan <- fmt.Errorf("stdout scanner error: %w", err):
		case <-t.ctx.Done():
		}
	}
}

func (t *SubprocessTransport) isControlMessage(line string) bool {
	return strings.Contains(line, `"type":"control_request"`) ||
		strings.Contains(line, `"type":"control_response"`) ||
		strings.Contains(line, `"type": "control_request"`) ||
		strings.Contains(line, `"type": "control_response"`)
}

func (t *SubprocessTransport) terminateProcess() error {
	if t.cmd == nil || t.cmd.Process == nil {
		return nil
	}

	if err := t.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		if isProcessFinished(err) {
			return nil
		}
		if killErr := t.cmd.Process.Kill(); killErr != nil && !isProcessFinished(killErr) {
			return killErr
		}
		return nil
	}

	done := make(chan error, 1)
	cmd := t.cmd
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil && strings.Contains(err.Error(), "signal:") {
			return nil
		}
		return err
	case <-time.After(terminationTimeoutSeconds * time.Second):
		if killErr := t.cmd.Process.Kill(); killErr != nil && !isProcessFinished(killErr) {
			return killErr
		}
		<-done
		return nil
	case <-t.ctx.Done():
		if killErr := t.cmd.Process.Kill(); killErr != nil && !isProcessFinished(killErr) {
			return killErr
		}
		<-done
		return nil
	}
}

func (t *SubprocessTransport) cleanup() {
	if t.stdout != nil {
		_ = t.stdout.Close()
		t.stdout = nil
	}

	if t.stderr != nil {
		_ = t.stderr.Close()
		_ = os.Remove(t.stderr.Name())
		t.stderr = nil
	}

	t.cmd = nil
}

func isProcessFinished(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "process already finished") ||
		strings.Contains(s, "process already released") ||
		strings.Contains(s, "no child processes") ||
		strings.Contains(s, "signal: killed")
}
