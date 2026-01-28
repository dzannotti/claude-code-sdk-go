package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
)

// SdkServerConfig represents an in-process MCP server configuration.
// This is the Go equivalent of the TypeScript SDK's McpSdkServerConfigWithInstance.
type SdkServerConfig struct {
	Type     string    `json:"type"`
	Name     string    `json:"name"`
	Instance McpServer `json:"-"`
}

func (SdkServerConfig) serverConfig() {}

// McpServer represents an in-process MCP server that handles tool calls.
type McpServer interface {
	Name() string
	Version() string
	ListTools() []ToolDefinition
	CallTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error)
}

// ToolDefinition describes a tool that can be called via MCP.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolResult represents the result of a tool call.
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError,omitempty"`
}

// ToolResultContent represents a content item in a tool result.
type ToolResultContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ToolHandler is a function that handles a tool call.
type ToolHandler[T any] func(ctx context.Context, args T) (*ToolResult, error)

// Tool creates a tool definition with a typed handler.
// The input schema is derived from the struct type T using reflection.
func Tool[T any](name, description string, handler ToolHandler[T]) *TypedTool[T] {
	var zero T
	schema := structToSchema(reflect.TypeOf(zero))
	return &TypedTool[T]{
		Definition: ToolDefinition{
			Name:        name,
			Description: description,
			InputSchema: schema,
		},
		Handler: handler,
	}
}

// TypedTool wraps a tool definition with its typed handler.
type TypedTool[T any] struct {
	Definition ToolDefinition
	Handler    ToolHandler[T]
}

// Call invokes the tool with the given arguments.
func (t *TypedTool[T]) Call(ctx context.Context, args map[string]any) (*ToolResult, error) {
	var input T
	data, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal args: %w", err)
	}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to unmarshal args: %w", err)
	}
	return t.Handler(ctx, input)
}

// SdkMcpServer is a simple in-process MCP server implementation.
type SdkMcpServer struct {
	name    string
	version string
	tools   map[string]toolCaller
	defs    []ToolDefinition
}

type toolCaller interface {
	call(ctx context.Context, args map[string]any) (*ToolResult, error)
}

type typedToolCaller[T any] struct {
	tool *TypedTool[T]
}

func (c *typedToolCaller[T]) call(ctx context.Context, args map[string]any) (*ToolResult, error) {
	return c.tool.Call(ctx, args)
}

// CreateSdkMcpServer creates a new in-process MCP server.
func CreateSdkMcpServer(name string, opts ...SdkMcpServerOption) *SdkMcpServer {
	s := &SdkMcpServer{
		name:    name,
		version: "1.0.0",
		tools:   make(map[string]toolCaller),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SdkMcpServerOption configures an SdkMcpServer.
type SdkMcpServerOption func(*SdkMcpServer)

// WithVersion sets the server version.
func WithVersion(version string) SdkMcpServerOption {
	return func(s *SdkMcpServer) {
		s.version = version
	}
}

// AddTool registers a typed tool with the server.
func AddTool[T any](tool *TypedTool[T]) SdkMcpServerOption {
	return func(s *SdkMcpServer) {
		s.tools[tool.Definition.Name] = &typedToolCaller[T]{tool: tool}
		s.defs = append(s.defs, tool.Definition)
	}
}

// Name returns the server name.
func (s *SdkMcpServer) Name() string { return s.name }

// Version returns the server version.
func (s *SdkMcpServer) Version() string { return s.version }

// ListTools returns all registered tool definitions.
func (s *SdkMcpServer) ListTools() []ToolDefinition { return s.defs }

// CallTool invokes a tool by name with the given arguments.
func (s *SdkMcpServer) CallTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error) {
	caller, ok := s.tools[name]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
	return caller.call(ctx, args)
}

// Config returns an SdkServerConfig for this server.
func (s *SdkMcpServer) Config() SdkServerConfig {
	return SdkServerConfig{
		Type:     "sdk",
		Name:     s.name,
		Instance: s,
	}
}

// structToSchema converts a Go struct type to a JSON Schema.
func structToSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return map[string]any{"type": "object"}
	}

	properties := make(map[string]any)
	required := []string{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		name := field.Name
		omitEmpty := false
		if jsonTag != "" {
			parts := splitTag(jsonTag)
			if parts[0] != "" {
				name = parts[0]
			}
			for _, p := range parts[1:] {
				if p == "omitempty" {
					omitEmpty = true
				}
			}
		}

		propSchema := typeToSchema(field.Type)
		if desc := field.Tag.Get("description"); desc != "" {
			propSchema["description"] = desc
		}

		properties[name] = propSchema
		if !omitEmpty {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func typeToSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": typeToSchema(t.Elem()),
		}
	case reflect.Map:
		return map[string]any{
			"type":                 "object",
			"additionalProperties": typeToSchema(t.Elem()),
		}
	case reflect.Ptr:
		return typeToSchema(t.Elem())
	case reflect.Struct:
		return structToSchema(t)
	default:
		return map[string]any{}
	}
}

func splitTag(tag string) []string {
	var parts []string
	for len(tag) > 0 {
		idx := 0
		for idx < len(tag) && tag[idx] != ',' {
			idx++
		}
		parts = append(parts, tag[:idx])
		if idx < len(tag) {
			tag = tag[idx+1:]
		} else {
			break
		}
	}
	return parts
}
