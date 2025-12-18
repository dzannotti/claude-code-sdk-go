package mcp

type ServerStatus struct {
	Name       string      `json:"name"`
	Status     string      `json:"status"`
	ServerInfo *ServerInfo `json:"serverInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors"`
}
