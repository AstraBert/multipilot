package shared

import (
	"testing"

	copilot "github.com/github/copilot-sdk/go"
)

func compareMcpServerKeys(expected, actual map[string]any) bool {
	if (expected == nil) != (actual == nil) {
		return false
	}
	if expected == nil {
		return true
	}
	if len(expected) != len(actual) {
		return false
	}
	for k := range expected {
		if _, ok := actual[k]; !ok {
			return false
		}
	}
	return true
}

func TestCopilotInput(t *testing.T) {
	t.Setenv("GH_TOKEN", "hello")
	testCases := []struct {
		task               CopilotInput
		expectedToken      string
		expectedLogFile    string
		expectedMcpServers map[string]any
		expectedError      bool
	}{
		{
			task: CopilotInput{
				GitHubToken: "ghp_actualtoken123",
				LogFile:     "/var/log/copilot.log",
			},
			expectedToken:      "ghp_actualtoken123",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
		{
			task: CopilotInput{
				GitHubToken: "$GH_TOKEN",
				LogFile:     "/var/log/copilot.log",
			},
			expectedToken:      "hello",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
		{
			task: CopilotInput{
				GitHubToken: "$GITHUB_TOKEN",
				LogFile:     "/var/log/copilot.log",
			},
			expectedToken:      "",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      true,
		},
		{
			task: CopilotInput{
				GitHubToken: "",
				LogFile:     "/var/log/copilot.log",
			},
			expectedToken:      "",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
		{
			task: CopilotInput{
				LogFile:     "",
				GitHubToken: "ghp_token",
			},
			expectedToken:      "ghp_token",
			expectedLogFile:    "",
			expectedMcpServers: nil,
			expectedError:      true,
		},
		{
			task: CopilotInput{
				LogFile:     "/tmp/test.log",
				GitHubToken: "ghp_token",
			},
			expectedToken:      "ghp_token",
			expectedLogFile:    "/tmp/test.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
		{
			task: CopilotInput{
				LogFile:          "/var/log/copilot.log",
				GitHubToken:      "ghp_token",
				LocalMcpServers:  nil,
				RemoteMcpServers: nil,
			},
			expectedToken:      "ghp_token",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
		{
			task: CopilotInput{
				LogFile:     "/var/log/copilot.log",
				GitHubToken: "ghp_token",
				LocalMcpServers: map[string]copilot.MCPLocalServerConfig{
					"server1": {},
				},
				RemoteMcpServers: nil,
			},
			expectedToken:   "ghp_token",
			expectedLogFile: "/var/log/copilot.log",
			expectedMcpServers: map[string]any{
				"server1": copilot.MCPLocalServerConfig{},
			},
			expectedError: false,
		},
		{
			task: CopilotInput{
				LogFile:         "/var/log/copilot.log",
				GitHubToken:     "ghp_token",
				LocalMcpServers: nil,
				RemoteMcpServers: map[string]copilot.MCPRemoteServerConfig{
					"remote1": {},
				},
			},
			expectedToken:   "ghp_token",
			expectedLogFile: "/var/log/copilot.log",
			expectedMcpServers: map[string]any{
				"remote1": copilot.MCPRemoteServerConfig{},
			},
			expectedError: false,
		},
		{
			task: CopilotInput{
				LogFile:     "/var/log/copilot.log",
				GitHubToken: "ghp_token",
				LocalMcpServers: map[string]copilot.MCPLocalServerConfig{
					"local1": {},
				},
				RemoteMcpServers: map[string]copilot.MCPRemoteServerConfig{
					"remote1": {},
				},
			},
			expectedToken:   "ghp_token",
			expectedLogFile: "/var/log/copilot.log",
			expectedMcpServers: map[string]any{
				"local1":  copilot.MCPLocalServerConfig{},
				"remote1": copilot.MCPRemoteServerConfig{},
			},
			expectedError: false,
		},
		{
			task: CopilotInput{
				LogFile:          "/var/log/copilot.log",
				GitHubToken:      "ghp_token",
				LocalMcpServers:  map[string]copilot.MCPLocalServerConfig{},
				RemoteMcpServers: map[string]copilot.MCPRemoteServerConfig{},
			},
			expectedToken:      "ghp_token",
			expectedLogFile:    "/var/log/copilot.log",
			expectedMcpServers: nil,
			expectedError:      false,
		},
	}

	for _, tc := range testCases {
		token, errTok := tc.task.GetToken()
		logFile, errLog := tc.task.GetLogFile()
		mcpServers := tc.task.GetMcpServers()
		if tc.expectedError && tc.expectedToken == "" && errTok == nil {
			t.Fatalf("Expected error when getting the token, got none")
		} else if tc.expectedError && tc.expectedLogFile == "" && errLog == nil {
			t.Fatalf("Expected error when getting the log gile, got none")
		} else if !tc.expectedError && (tc.expectedLogFile != logFile || tc.expectedToken != token) {
			t.Fatalf("Expected token to be %s and log file to be %s, got %s and %s", tc.expectedToken, tc.expectedLogFile, token, logFile)
		}
		if !compareMcpServerKeys(mcpServers, tc.expectedMcpServers) {
			t.Fatalf("Expected MCP servers %v, got %v", tc.expectedMcpServers, mcpServers)
		}
	}
}

func TestCopilotTasks(t *testing.T) {
	testCases := []struct {
		tasks         CopilotTasks
		expectedError bool
		errorMessage  string
	}{
		{
			tasks: CopilotTasks{
				Tasks: []CopilotInput{
					{
						LogFile: "hello.jsonl",
						Cwd:     "/test/dir",
					},
					{
						LogFile: "hello1.jsonl",
						Cwd:     "/test/dir1",
					},
				},
			},
			expectedError: false,
			errorMessage:  "",
		},
		{
			tasks: CopilotTasks{
				Tasks: []CopilotInput{
					{
						LogFile: "hello.jsonl",
						Cwd:     "/test/dir",
					},
					{
						LogFile: "hello.jsonl",
						Cwd:     "/test/dir1",
					},
				},
			},
			expectedError: true,
			errorMessage:  "cannot use the same log file for two or more tasks because of potential race conditions",
		},
		{
			tasks: CopilotTasks{
				Tasks: []CopilotInput{
					{
						LogFile: "hello.jsonl",
						Cwd:     "/test/dir",
					},
					{
						LogFile: "hello1.jsonl",
						Cwd:     "/test/dir",
					},
				},
			},
			expectedError: true,
			errorMessage:  "cannot use the same working directory for mulitple tasks because of potential race conditions",
		},
	}
	for _, tc := range testCases {
		err := tc.tasks.Validate()
		if tc.expectedError && err != nil {
			if tc.errorMessage != err.Error() {
				t.Fatalf("Expected error message to be %s, got %s", tc.errorMessage, err.Error())
			}
		} else if tc.expectedError && err == nil {
			t.Fatal("Expected an error, but none gotten")
		} else if !tc.expectedError && err != nil {
			t.Fatalf("No error expected, got %s", err.Error())
		}
	}
}
