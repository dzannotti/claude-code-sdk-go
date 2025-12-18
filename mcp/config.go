package mcp

type ServerConfig interface {
	serverConfig()
}

type StdioServerConfig struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (StdioServerConfig) serverConfig() {}

type SSEServerConfig struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (SSEServerConfig) serverConfig() {}

type HTTPServerConfig struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (HTTPServerConfig) serverConfig() {}
