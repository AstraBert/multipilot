package shared

import (
	"errors"

	copilot "github.com/github/copilot-sdk/go"
)

const DefaultAiModel string = "gpt-4.1"
const DefaultTimeout int64 = 120

type CopilotInput struct {
	LogFile          string                                   `json:"log_file"`
	Cwd              string                                   `json:"cwd"`
	LogLevel         string                                   `json:"log_level"`
	Env              []string                                 `json:"env"`
	Prompt           string                                   `json:"prompt"`
	GitHubToken      string                                   `json:"token"`
	AiModel          string                                   `json:"ai_model"`
	SystemPrompt     string                                   `json:"system_prompt"`
	AllowedTools     []string                                 `json:"allowed_tools"`
	Skills           []string                                 `json:"skills"`
	LocalMcpServers  map[string]copilot.MCPLocalServerConfig  `json:"local_mcp_servers"`
	RemoteMcpServers map[string]copilot.MCPRemoteServerConfig `json:"remote_mcp_servers"`
	Timeout          int64                                    `json:"timeout_sec"`
}

type CopilotTasks struct {
	Tasks []CopilotInput `json:"tasks"`
}

func (c CopilotInput) GetMcpServers() map[string]any {
	mcpServers := map[string]any{}
	if c.LocalMcpServers != nil {
		for k := range c.LocalMcpServers {
			mcpServers[k] = c.LocalMcpServers[k]
		}
	}
	if c.RemoteMcpServers != nil {
		for k := range c.RemoteMcpServers {
			mcpServers[k] = c.RemoteMcpServers[k]
		}
	}
	if len(mcpServers) > 0 {
		return mcpServers
	}
	return nil
}

func (c CopilotInput) GetLogFile() (string, error) {
	if c.LogFile == "" {
		return "", errors.New("log_file cannot be empty")
	}
	return c.LogFile, nil
}

func (t *CopilotTasks) Validate() error {
	logFiles := make(map[string]int)
	cwds := make(map[string]int)
	for i, task := range t.Tasks {
		if _, ok := cwds[task.Cwd]; ok {
			return errors.New("cannot use the same working directory for mulitple tasks because of potential race conditions")
		}
		if _, ok := logFiles[task.LogFile]; ok {
			return errors.New("cannot use the same log file for two or more tasks because of potential race conditions")
		}
		logFiles[task.LogFile] = i
		cwds[task.Cwd] = i
	}
	return nil
}
