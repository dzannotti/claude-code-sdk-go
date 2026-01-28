package protocol

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"claudeagent/control"
	"claudeagent/mcp"
)

type ControlRequest struct {
	Type      string         `json:"type"`
	RequestID string         `json:"request_id"`
	Request   map[string]any `json:"request"`
}

type ControlResponse struct {
	Type     string          `json:"type"`
	Response ResponsePayload `json:"response"`
}

type ResponsePayload struct {
	Subtype   string         `json:"subtype"`
	RequestID string         `json:"request_id"`
	Response  map[string]any `json:"response,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type CanUseToolRequest struct {
	Subtype               string                     `json:"subtype"`
	ToolName              string                     `json:"tool_name"`
	Input                 map[string]any             `json:"input"`
	PermissionSuggestions []control.PermissionUpdate `json:"permission_suggestions,omitempty"`
	BlockedPath           *string                    `json:"blocked_path,omitempty"`
	DecisionReason        *string                    `json:"decision_reason,omitempty"`
	ToolUseID             string                     `json:"tool_use_id"`
	AgentID               *string                    `json:"agent_id,omitempty"`
}

type HookCallbackRequest struct {
	Subtype    string  `json:"subtype"`
	CallbackID string  `json:"callback_id"`
	Input      any     `json:"input"`
	ToolUseID  *string `json:"tool_use_id,omitempty"`
}

type ControlHandler struct {
	sendFn    func(ctx context.Context, data []byte) error
	requestID atomic.Uint64
	pending   map[string]chan *ResponsePayload
	mu        sync.RWMutex

	canUseTool control.CanUseToolFunc
	hooks      map[control.HookEvent][]control.HookCallbackMatcher
	hooksByID  map[string]control.HookCallback
}

func NewControlHandler(sendFn func(ctx context.Context, data []byte) error) *ControlHandler {
	return &ControlHandler{
		sendFn:    sendFn,
		pending:   make(map[string]chan *ResponsePayload),
		hooksByID: make(map[string]control.HookCallback),
	}
}

func (h *ControlHandler) SetCanUseTool(fn control.CanUseToolFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.canUseTool = fn
}

func (h *ControlHandler) SetHooks(hooks map[control.HookEvent][]control.HookCallbackMatcher) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hooks = hooks
}

func (h *ControlHandler) RegisterHookCallback(id string, fn control.HookCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hooksByID[id] = fn
}

func (h *ControlHandler) nextRequestID() string {
	id := h.requestID.Add(1)
	return fmt.Sprintf("sdk-req-%d", id)
}

func (h *ControlHandler) SendRequest(ctx context.Context, subtype string, payload map[string]any) (*ResponsePayload, error) {
	reqID := h.nextRequestID()

	respChan := make(chan *ResponsePayload, 1)
	h.mu.Lock()
	h.pending[reqID] = respChan
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.pending, reqID)
		h.mu.Unlock()
	}()

	if payload == nil {
		payload = make(map[string]any)
	}
	payload["subtype"] = subtype

	req := ControlRequest{
		Type:      "control_request",
		RequestID: reqID,
		Request:   payload,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := h.sendFn(ctx, data); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respChan:
		if resp.Subtype == "error" {
			return nil, fmt.Errorf("control error: %s", resp.Error)
		}
		return resp, nil
	}
}

func (h *ControlHandler) HandleIncoming(ctx context.Context, data []byte) ([]byte, error) {
	var typeHolder struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeHolder); err != nil {
		return nil, fmt.Errorf("failed to parse type: %w", err)
	}

	switch typeHolder.Type {
	case "control_response":
		return nil, h.handleResponse(data)
	case "control_request":
		return h.handleRequest(ctx, data)
	default:
		return nil, nil
	}
}

func (h *ControlHandler) handleResponse(data []byte) error {
	var resp ControlResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	h.mu.RLock()
	respChan, ok := h.pending[resp.Response.RequestID]
	h.mu.RUnlock()

	if ok {
		respChan <- &resp.Response
	}

	return nil
}

func (h *ControlHandler) handleRequest(ctx context.Context, data []byte) ([]byte, error) {
	var req ControlRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("failed to parse request: %w", err)
	}

	subtype, _ := req.Request["subtype"].(string)

	var response map[string]any
	var respErr error

	switch subtype {
	case "can_use_tool":
		response, respErr = h.handleCanUseTool(ctx, req.Request)
	case "hook_callback":
		response, respErr = h.handleHookCallback(ctx, req.Request)
	default:
		respErr = fmt.Errorf("unknown request subtype: %s", subtype)
	}

	resp := ControlResponse{
		Type: "control_response",
		Response: ResponsePayload{
			RequestID: req.RequestID,
		},
	}

	if respErr != nil {
		resp.Response.Subtype = "error"
		resp.Response.Error = respErr.Error()
	} else {
		resp.Response.Subtype = "success"
		resp.Response.Response = response
	}

	return json.Marshal(resp)
}

func (h *ControlHandler) handleCanUseTool(ctx context.Context, reqData map[string]any) (map[string]any, error) {
	h.mu.RLock()
	fn := h.canUseTool
	h.mu.RUnlock()

	if fn == nil {
		return map[string]any{
			"behavior":      "allow",
			"updated_input": reqData["input"],
		}, nil
	}

	toolName, _ := reqData["tool_name"].(string)
	input, _ := reqData["input"].(map[string]any)
	toolUseID, _ := reqData["tool_use_id"].(string)

	var suggestions []control.PermissionUpdate
	if raw, ok := reqData["permission_suggestions"].([]any); ok {
		data, _ := json.Marshal(raw)
		_ = json.Unmarshal(data, &suggestions) // best effort conversion
	}

	var blockedPath, decisionReason *string
	if v, ok := reqData["blocked_path"].(string); ok {
		blockedPath = &v
	}
	if v, ok := reqData["decision_reason"].(string); ok {
		decisionReason = &v
	}
	var agentID *string
	if v, ok := reqData["agent_id"].(string); ok {
		agentID = &v
	}

	opts := control.CanUseToolOptions{
		Suggestions:    suggestions,
		BlockedPath:    blockedPath,
		DecisionReason: decisionReason,
		ToolUseID:      toolUseID,
		AgentID:        agentID,
	}

	result, err := fn(ctx, toolName, input, opts)
	if err != nil {
		return nil, err
	}

	resp := map[string]any{
		"behavior": string(result.Behavior),
	}

	if result.Behavior == control.PermissionAllow {
		resp["updated_input"] = result.UpdatedInput
		if len(result.UpdatedPermissions) > 0 {
			resp["updated_permissions"] = result.UpdatedPermissions
		}
	} else {
		resp["message"] = result.Message
		if result.Interrupt {
			resp["interrupt"] = true
		}
	}

	return resp, nil
}

func (h *ControlHandler) handleHookCallback(ctx context.Context, reqData map[string]any) (map[string]any, error) {
	callbackID, _ := reqData["callback_id"].(string)

	h.mu.RLock()
	fn, ok := h.hooksByID[callbackID]
	h.mu.RUnlock()

	if !ok {
		return map[string]any{"continue": true}, nil
	}

	var toolUseID *string
	if v, ok := reqData["tool_use_id"].(string); ok {
		toolUseID = &v
	}

	inputData, _ := json.Marshal(reqData["input"])
	var hookInput control.HookInput
	_ = json.Unmarshal(inputData, &hookInput) // best effort conversion

	output, err := fn(ctx, hookInput, toolUseID)
	if err != nil {
		return nil, err
	}

	resp := make(map[string]any)
	data, _ := json.Marshal(output)
	_ = json.Unmarshal(data, &resp) // always succeeds for map[string]any

	return resp, nil
}

func (h *ControlHandler) Interrupt(ctx context.Context) error {
	_, err := h.SendRequest(ctx, "interrupt", nil)
	return err
}

func (h *ControlHandler) SetPermissionMode(ctx context.Context, mode control.PermissionMode) error {
	_, err := h.SendRequest(ctx, "set_permission_mode", map[string]any{
		"mode": string(mode),
	})
	return err
}

func (h *ControlHandler) SetModel(ctx context.Context, model *string) error {
	payload := make(map[string]any)
	if model != nil {
		payload["model"] = *model
	}
	_, err := h.SendRequest(ctx, "set_model", payload)
	return err
}

func (h *ControlHandler) SetMaxThinkingTokens(ctx context.Context, tokens *int) error {
	payload := make(map[string]any)
	if tokens != nil {
		payload["max_thinking_tokens"] = *tokens
	} else {
		payload["max_thinking_tokens"] = nil
	}
	_, err := h.SendRequest(ctx, "set_max_thinking_tokens", payload)
	return err
}

func (h *ControlHandler) RewindFiles(ctx context.Context, userMessageID string) error {
	_, err := h.SendRequest(ctx, "rewind_files", map[string]any{
		"user_message_id": userMessageID,
	})
	return err
}

type RewindFilesResult struct {
	CanRewind    bool     `json:"canRewind"`
	Error        *string  `json:"error,omitempty"`
	FilesChanged []string `json:"filesChanged,omitempty"`
	Insertions   *int     `json:"insertions,omitempty"`
	Deletions    *int     `json:"deletions,omitempty"`
}

func (h *ControlHandler) RewindFilesWithOptions(ctx context.Context, userMessageID string, dryRun bool) (*RewindFilesResult, error) {
	payload := map[string]any{
		"user_message_id": userMessageID,
	}
	if dryRun {
		payload["dry_run"] = true
	}
	resp, err := h.SendRequest(ctx, "rewind_files", payload)
	if err != nil {
		return nil, err
	}

	result := &RewindFilesResult{}
	if canRewind, ok := resp.Response["canRewind"].(bool); ok {
		result.CanRewind = canRewind
	}
	if errStr, ok := resp.Response["error"].(string); ok {
		result.Error = &errStr
	}
	if files, ok := resp.Response["filesChanged"].([]any); ok {
		for _, f := range files {
			if s, ok := f.(string); ok {
				result.FilesChanged = append(result.FilesChanged, s)
			}
		}
	}
	if insertions, ok := resp.Response["insertions"].(float64); ok {
		i := int(insertions)
		result.Insertions = &i
	}
	if deletions, ok := resp.Response["deletions"].(float64); ok {
		d := int(deletions)
		result.Deletions = &d
	}
	return result, nil
}

func (h *ControlHandler) ReconnectMcpServer(ctx context.Context, serverName string) error {
	_, err := h.SendRequest(ctx, "mcp_reconnect", map[string]any{
		"server_name": serverName,
	})
	return err
}

func (h *ControlHandler) ToggleMcpServer(ctx context.Context, serverName string, enabled bool) error {
	_, err := h.SendRequest(ctx, "mcp_toggle", map[string]any{
		"server_name": serverName,
		"enabled":     enabled,
	})
	return err
}

func (h *ControlHandler) McpStatus(ctx context.Context) ([]mcp.ServerStatus, error) {
	resp, err := h.SendRequest(ctx, "mcp_status", nil)
	if err != nil {
		return nil, err
	}

	var statuses []mcp.ServerStatus
	if raw, ok := resp.Response["statuses"].([]any); ok {
		data, err := json.Marshal(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal mcp statuses: %w", err)
		}
		if err := json.Unmarshal(data, &statuses); err != nil {
			return nil, fmt.Errorf("failed to parse mcp statuses: %w", err)
		}
	}

	return statuses, nil
}

func (h *ControlHandler) SetMcpServers(ctx context.Context, servers map[string]mcp.ServerConfig) (*mcp.SetServersResult, error) {
	payload := map[string]any{
		"servers": servers,
	}
	resp, err := h.SendRequest(ctx, "mcp_set_servers", payload)
	if err != nil {
		return nil, err
	}

	result := &mcp.SetServersResult{}
	if added, ok := resp.Response["added"].([]any); ok {
		for _, a := range added {
			if s, ok := a.(string); ok {
				result.Added = append(result.Added, s)
			}
		}
	}
	if removed, ok := resp.Response["removed"].([]any); ok {
		for _, r := range removed {
			if s, ok := r.(string); ok {
				result.Removed = append(result.Removed, s)
			}
		}
	}
	if errors, ok := resp.Response["errors"].(map[string]any); ok {
		result.Errors = make(map[string]string)
		for k, v := range errors {
			if s, ok := v.(string); ok {
				result.Errors[k] = s
			}
		}
	}

	return result, nil
}

func (h *ControlHandler) Initialize(ctx context.Context, hooks map[control.HookEvent][]HookCallbackMatcher, sdkMcpServers []string, jsonSchema map[string]any, systemPrompt, appendSystemPrompt *string, agents map[string]any) (*InitializeResponse, error) {
	payload := make(map[string]any)

	if len(hooks) > 0 {
		hookData := make(map[string]any)
		for event, matchers := range hooks {
			hookData[string(event)] = matchers
		}
		payload["hooks"] = hookData
	}
	if len(sdkMcpServers) > 0 {
		payload["sdkMcpServers"] = sdkMcpServers
	}
	if len(jsonSchema) > 0 {
		payload["jsonSchema"] = jsonSchema
	}
	if systemPrompt != nil {
		payload["systemPrompt"] = *systemPrompt
	}
	if appendSystemPrompt != nil {
		payload["appendSystemPrompt"] = *appendSystemPrompt
	}
	if len(agents) > 0 {
		payload["agents"] = agents
	}

	resp, err := h.SendRequest(ctx, "initialize", payload)
	if err != nil {
		return nil, err
	}

	result := &InitializeResponse{}
	data, err := json.Marshal(resp.Response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal init response: %w", err)
	}
	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("failed to parse init response: %w", err)
	}

	return result, nil
}

type HookCallbackMatcher struct {
	Matcher         *string  `json:"matcher,omitempty"`
	HookCallbackIDs []string `json:"hookCallbackIds"`
	Timeout         *int     `json:"timeout,omitempty"`
}

type InitializeResponse struct {
	Commands              []SlashCommand `json:"commands"`
	OutputStyle           string         `json:"output_style"`
	AvailableOutputStyles []string       `json:"available_output_styles"`
	Models                []ModelInfo    `json:"models"`
	Account               AccountInfo    `json:"account"`
}

type SlashCommand struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	ArgumentHint string `json:"argumentHint"`
}

type ModelInfo struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}

type AccountInfo struct {
	Email            *string `json:"email,omitempty"`
	Organization     *string `json:"organization,omitempty"`
	SubscriptionType *string `json:"subscriptionType,omitempty"`
	TokenSource      *string `json:"tokenSource,omitempty"`
	APIKeySource     *string `json:"apiKeySource,omitempty"`
}
