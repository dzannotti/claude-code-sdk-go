package mcp

type ServerStatus struct {
	Name       string              `json:"name"`
	Status     string              `json:"status"`
	ServerInfo *ServerInfo         `json:"serverInfo,omitempty"`
	Error      *string             `json:"error,omitempty"`
	Config     *ServerStatusConfig `json:"config,omitempty"`
	Scope      *string             `json:"scope,omitempty"`
	Tools      []ToolInfo          `json:"tools,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServerStatusConfig struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type ToolInfo struct {
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

type ToolAnnotations struct {
	ReadOnly    *bool `json:"readOnly,omitempty"`
	Destructive *bool `json:"destructive,omitempty"`
	OpenWorld   *bool `json:"openWorld,omitempty"`
}

type SetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors"`
}
