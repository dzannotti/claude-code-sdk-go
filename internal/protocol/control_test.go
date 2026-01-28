package protocol

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"claudeagent/control"
	"claudeagent/mcp"
)

type mockSender struct {
	mu      sync.Mutex
	sent    [][]byte
	sendErr error
}

func (m *mockSender) send(_ context.Context, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sendErr != nil {
		return m.sendErr
	}
	m.sent = append(m.sent, data)
	return nil
}

func (m *mockSender) getSent() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sent
}

func TestControlHandler_SendRequest_Success(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
				Response:  map[string]any{"result": "ok"},
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	resp, err := handler.SendRequest(ctx, "test_subtype", map[string]any{"foo": "bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Subtype != "success" {
		t.Errorf("expected subtype 'success', got %q", resp.Subtype)
	}

	if resp.Response["result"] != "ok" {
		t.Errorf("expected result 'ok', got %v", resp.Response["result"])
	}
}

func TestControlHandler_SendRequest_Error(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "error",
				Error:     "something went wrong",
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	_, err := handler.SendRequest(ctx, "test_subtype", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "control error: something went wrong" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestControlHandler_SendRequest_Timeout(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := handler.SendRequest(ctx, "test_subtype", nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestControlHandler_SendRequest_SendError(t *testing.T) {
	expectedErr := errors.New("send failed")
	sender := &mockSender{sendErr: expectedErr}
	handler := NewControlHandler(sender.send)

	ctx := context.Background()
	_, err := handler.SendRequest(ctx, "test_subtype", nil)
	if err == nil || !errors.Is(err, expectedErr) {
		t.Errorf("expected send error, got %v", err)
	}
}

func TestControlHandler_HandleCanUseTool_NoCallback(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "ls"},
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	if resp.Response.Subtype != "success" {
		t.Errorf("expected success, got %q", resp.Response.Subtype)
	}

	behavior, _ := resp.Response.Response["behavior"].(string)
	if behavior != "allow" {
		t.Errorf("expected behavior 'allow', got %q", behavior)
	}
}

func TestControlHandler_HandleCanUseTool_WithCallback(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	handler.SetCanUseTool(func(ctx context.Context, toolName string, input map[string]any, opts control.CanUseToolOptions) (control.PermissionResult, error) {
		if toolName == "Bash" {
			return control.PermissionResult{
				Behavior: control.PermissionDeny,
				Message:  "Bash not allowed",
			}, nil
		}
		return control.PermissionResult{
			Behavior:     control.PermissionAllow,
			UpdatedInput: input,
		}, nil
	})

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "rm -rf /"},
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	behavior, _ := resp.Response.Response["behavior"].(string)
	if behavior != "deny" {
		t.Errorf("expected behavior 'deny', got %q", behavior)
	}

	msg, _ := resp.Response.Response["message"].(string)
	if msg != "Bash not allowed" {
		t.Errorf("expected message 'Bash not allowed', got %q", msg)
	}
}

func TestControlHandler_HandleCanUseTool_AllowWithUpdatedInput(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	handler.SetCanUseTool(func(ctx context.Context, toolName string, input map[string]any, opts control.CanUseToolOptions) (control.PermissionResult, error) {
		input["sanitized"] = true
		return control.PermissionResult{
			Behavior:     control.PermissionAllow,
			UpdatedInput: input,
			UpdatedPermissions: []control.PermissionUpdate{
				{Type: "allow_tool", Rules: []control.PermissionRule{{ToolName: "Bash", RuleContent: "ls:*"}}},
			},
		}, nil
	})

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "ls"},
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	if resp.Response.Subtype != "success" {
		t.Errorf("expected success, got %q", resp.Response.Subtype)
	}

	updatedInput, _ := resp.Response.Response["updated_input"].(map[string]any)
	if updatedInput["sanitized"] != true {
		t.Errorf("expected sanitized=true in updated input")
	}

	perms, ok := resp.Response.Response["updated_permissions"]
	if !ok {
		t.Error("expected updated_permissions in response")
	}
	permsList, ok := perms.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", perms)
	}
	if len(permsList) != 1 {
		t.Errorf("expected 1 permission update, got %d", len(permsList))
	}
}

func TestControlHandler_HandleHookCallback(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	callbackCalled := false
	cont := true
	handler.RegisterHookCallback("hook-123", func(ctx context.Context, input control.HookInput, toolUseID *string) (control.HookOutput, error) {
		callbackCalled = true
		return control.HookOutput{
			Continue: &cont,
		}, nil
	})

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook-123",
			"input":       map[string]any{},
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !callbackCalled {
		t.Error("expected callback to be called")
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	if resp.Response.Subtype != "success" {
		t.Errorf("expected success, got %q", resp.Response.Subtype)
	}
}

func TestControlHandler_HandleHookCallback_UnknownID(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "unknown-id",
			"input":       map[string]any{},
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	cont, _ := resp.Response.Response["continue"].(bool)
	if !cont {
		t.Error("expected continue=true for unknown callback")
	}
}

func TestControlHandler_Interrupt(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		if req.Request["subtype"] != "interrupt" {
			t.Errorf("expected subtype 'interrupt', got %v", req.Request["subtype"])
		}

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	err := handler.Interrupt(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestControlHandler_SetPermissionMode(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		if req.Request["subtype"] != "set_permission_mode" {
			t.Errorf("expected subtype 'set_permission_mode', got %v", req.Request["subtype"])
		}
		if req.Request["mode"] != "acceptEdits" {
			t.Errorf("expected mode 'acceptEdits', got %v", req.Request["mode"])
		}

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	err := handler.SetPermissionMode(ctx, control.PermissionModeAcceptEdits)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestControlHandler_SetModel(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		if req.Request["subtype"] != "set_model" {
			t.Errorf("expected subtype 'set_model', got %v", req.Request["subtype"])
		}
		if req.Request["model"] != "claude-sonnet-4-20250514" {
			t.Errorf("expected model 'claude-sonnet-4-20250514', got %v", req.Request["model"])
		}

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	model := "claude-sonnet-4-20250514"
	err := handler.SetModel(ctx, &model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestControlHandler_McpStatus(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
				Response: map[string]any{
					"statuses": []any{
						map[string]any{
							"name":   "test-server",
							"status": "connected",
						},
					},
				},
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	statuses, err := handler.McpStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}

	if statuses[0].Name != "test-server" {
		t.Errorf("expected name 'test-server', got %q", statuses[0].Name)
	}
}

func TestControlHandler_SetMcpServers(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		time.Sleep(10 * time.Millisecond)
		sent := sender.getSent()
		if len(sent) == 0 {
			return
		}

		var req ControlRequest
		json.Unmarshal(sent[0], &req)

		resp := ControlResponse{
			Type: "control_response",
			Response: ResponsePayload{
				RequestID: req.RequestID,
				Subtype:   "success",
				Response: map[string]any{
					"added":   []any{"new-server"},
					"removed": []any{"old-server"},
					"errors":  map[string]any{"bad-server": "connection failed"},
				},
			},
		}
		respBytes, _ := json.Marshal(resp)
		handler.HandleIncoming(ctx, respBytes)
	}()

	servers := map[string]mcp.ServerConfig{
		"new-server": mcp.StdioServerConfig{
			Type:    "stdio",
			Command: "node",
			Args:    []string{"server.js"},
		},
	}

	result, err := handler.SetMcpServers(ctx, servers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "new-server" {
		t.Errorf("expected added=['new-server'], got %v", result.Added)
	}

	if len(result.Removed) != 1 || result.Removed[0] != "old-server" {
		t.Errorf("expected removed=['old-server'], got %v", result.Removed)
	}

	if result.Errors["bad-server"] != "connection failed" {
		t.Errorf("expected error for bad-server, got %v", result.Errors)
	}
}

func TestControlHandler_UnknownMessageType(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx := context.Background()

	data := []byte(`{"type":"unknown_type"}`)
	resp, err := handler.HandleIncoming(ctx, data)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp != nil {
		t.Errorf("expected nil response for unknown type, got %v", resp)
	}
}

func TestControlHandler_UnknownRequestSubtype(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	ctx := context.Background()

	req := ControlRequest{
		Type:      "control_request",
		RequestID: "req-1",
		Request: map[string]any{
			"subtype": "unknown_subtype",
		},
	}
	reqBytes, _ := json.Marshal(req)

	respBytes, err := handler.HandleIncoming(ctx, reqBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var resp ControlResponse
	json.Unmarshal(respBytes, &resp)

	if resp.Response.Subtype != "error" {
		t.Errorf("expected error subtype, got %q", resp.Response.Subtype)
	}

	if resp.Response.Error == "" {
		t.Error("expected error message")
	}
}

func TestControlHandler_RequestIDIncrement(t *testing.T) {
	sender := &mockSender{}
	handler := NewControlHandler(sender.send)

	id1 := handler.nextRequestID()
	id2 := handler.nextRequestID()
	id3 := handler.nextRequestID()

	if id1 == id2 || id2 == id3 || id1 == id3 {
		t.Errorf("request IDs should be unique: %s, %s, %s", id1, id2, id3)
	}
}
