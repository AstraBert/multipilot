package cmd

import (
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/AstraBert/multipilot/shared"
	copilot "github.com/github/copilot-sdk/go"
)

func compareConfigs(cfg1, cfg2 *shared.CopilotTasks) bool {
	if (cfg1 == nil) != (cfg2 == nil) {
		return false
	}
	if cfg1 == nil && cfg2 == nil {
		return true
	}
	if len(cfg1.Tasks) != len(cfg2.Tasks) {
		return false
	}
	for i := range cfg1.Tasks {
		// compare comparables
		if cfg1.Tasks[i].LogFile != cfg2.Tasks[i].LogFile || cfg1.Tasks[i].Cwd != cfg2.Tasks[i].Cwd || cfg1.Tasks[i].LogLevel != cfg2.Tasks[i].LogLevel || cfg1.Tasks[i].Prompt != cfg2.Tasks[i].Prompt || cfg1.Tasks[i].AiModel != cfg2.Tasks[i].AiModel || cfg1.Tasks[i].SystemPrompt != cfg2.Tasks[i].SystemPrompt || cfg1.Tasks[i].Timeout != cfg2.Tasks[i].Timeout {
			return false
		}
	}
	return true
}

func TestReadConfigToTasks(t *testing.T) {
	testCases := []struct {
		configFile      string
		expectedError   bool
		validationError string
		expectedConfig  *shared.CopilotTasks
	}{
		{
			configFile:      "../testfiles/configs/correct.json",
			expectedError:   false,
			validationError: "",
			expectedConfig: &shared.CopilotTasks{
				Tasks: []shared.CopilotInput{
					{
						LogFile:          "copilot-session-multipilot.jsonl",
						Cwd:              "/Users/user/code-projects/multipilot",
						LogLevel:         "info",
						Env:              []string{},
						Prompt:           "What is the workflow engine that the current project is using?",
						GitHubToken:      "$GITHUB_TOKEN",
						AiModel:          "gpt-4.1",
						SystemPrompt:     "You are a helpful assistant that performs exploratory tasks within Go codebases.",
						ExcludeTools:     []string{"shell(rm)", "write", "shell(rmdir)"},
						Skills:           []string{},
						LocalMcpServers:  map[string]copilot.MCPLocalServerConfig{},
						RemoteMcpServers: map[string]copilot.MCPRemoteServerConfig{},
						Timeout:          300,
					},
					{
						LogFile:          "copilot-session-workflowsacp.jsonl",
						Cwd:              "/Users/user/code-projects/workflows-acp",
						LogLevel:         "info",
						Env:              []string{},
						Prompt:           "What is the workflow engine that the current project is using?",
						GitHubToken:      "$GITHUB_TOKEN",
						AiModel:          "gpt-4.1",
						SystemPrompt:     "You are a helpful assistant that performs exploratory tasks within python codebases managed with uv.",
						ExcludeTools:     []string{"shell(rm)", "write", "shell(rmdir)"},
						Skills:           []string{},
						LocalMcpServers:  map[string]copilot.MCPLocalServerConfig{},
						RemoteMcpServers: map[string]copilot.MCPRemoteServerConfig{},
						Timeout:          300,
					},
				},
			},
		},
		{
			configFile:      "../testfiles/configs/invalid.json",
			expectedError:   true,
			validationError: "cannot use the same log file for two or more tasks because of potential race conditions",
			expectedConfig:  nil,
		},
		{
			configFile:      "../testfiles/configs/notjson.txt",
			expectedError:   true,
			validationError: "",
			expectedConfig:  nil,
		},
	}

	for _, tc := range testCases {
		config, err := ReadConfigToTasks(tc.configFile)
		if tc.expectedError && err != nil {
			if tc.validationError != "" && err.Error() != tc.validationError {
				t.Fatalf("Expected a validation error to occur with message %s, got %s", tc.validationError, err.Error())
			}
		} else if tc.expectedError && err == nil {
			t.Fatal("Expected an error to occur, but got none")
		} else if !tc.expectedError && err != nil {
			t.Fatalf("Not expecting an error, got %s", err.Error())
		}
		if !compareConfigs(config, tc.expectedConfig) {
			t.Fatalf("Expected config to be %v, got %v", tc.expectedConfig, config)
		}
	}
}

func compareEvents(ev1, ev2 shared.CopilotEvent) bool {
	return ev1.ID == ev2.ID && ev1.Type == ev2.Type && maps.Equal(ev1.Data, ev2.Data) && ev1.Timestamp.Equal(ev2.Timestamp)
}

func TestLoadEvents(t *testing.T) {
	testCases := []struct {
		logFile        string
		expectedEvents []shared.CopilotEvent
		expectedError  bool
	}{
		{
			logFile: "../testfiles/logs/valid.logs",
			expectedEvents: []shared.CopilotEvent{
				{
					Timestamp: time.Date(2026, 2, 6, 11, 7, 11, 610000000, time.UTC),
					ID:        "abd2d4e9-d68b-41b1-9ab2-01ea62e622d6",
					Data: map[string]any{
						"turnId": "1",
					},
					Type: "assistant.turn_end",
				},
				{
					Timestamp: time.Date(2026, 2, 6, 11, 7, 11, 610000000, time.UTC),
					ID:        "503b8270-43a6-4922-8571-dc2b1a5418ae",
					Data:      map[string]any{},
					Type:      "session.idle",
				},
			},
			expectedError: false,
		},
		{
			logFile:        "../testfiles/logs/invalid.logs",
			expectedEvents: nil,
			expectedError:  true,
		},
	}
	for _, tc := range testCases {
		events, err := LoadEvents(tc.logFile)
		if tc.expectedError && err == nil {
			t.Fatal("An error was expected, got none")
		} else if !tc.expectedError && err != nil {
			t.Fatalf("Not expecting any error, got %s", err.Error())
		}
		if events == nil && tc.expectedEvents != nil {
			t.Fatal("Expecting events to be a non-null slice")
		} else if events != nil && tc.expectedEvents == nil {
			t.Fatalf("Expecting events to be a null slice, got %v", events)
		} else {
			if !slices.EqualFunc(events, tc.expectedEvents, compareEvents) {
				t.Fatalf("Expecting events to be %v, got %v", tc.expectedEvents, events)
			}
		}
	}
}
