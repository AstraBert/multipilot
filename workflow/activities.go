package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AstraBert/multipilot/shared"
	copilot "github.com/github/copilot-sdk/go"
)

func RunCopilot(ctx context.Context, task shared.CopilotInput) error {
	recordFile, err := task.GetLogFile()
	if err != nil {
		return err
	}
	options := &copilot.ClientOptions{Cwd: task.Cwd, LogLevel: task.LogLevel, Env: task.Env}
	if task.GitHubToken != "" {
		options.GithubToken = task.GitHubToken
	}
	client := copilot.NewClient(options)
	if err := client.Start(); err != nil {
		return fmt.Errorf("An error occurred while starting the client: %s", err.Error())
	}
	defer client.Stop()

	var model string
	switch task.AiModel {
	case "":
		model = shared.DefaultAiModel
	default:
		model = task.AiModel
	}

	systemPrompt := &copilot.SystemMessageAppendConfig{}

	switch task.SystemPrompt {
	case "":
		systemPrompt = nil
	default:
		systemPrompt.Content = task.SystemPrompt
	}

	servers := task.GetMcpServers()
	mcpServers := make(map[string]copilot.MCPServerConfig)
	for k, v := range servers {
		if serverMap, ok := v.(map[string]any); ok {
			mcpServers[k] = serverMap
		}
	}

	var timeout int64
	if task.Timeout <= 0 {
		timeout = shared.DefaultTimeout
	} else {
		timeout = task.Timeout
	}

	// Create session
	session, err := client.CreateSession(&copilot.SessionConfig{
		Model:            model,
		WorkingDirectory: task.Cwd,
		AvailableTools:   task.AllowedTools,
		SystemMessage:    (*copilot.SystemMessageConfig)(systemPrompt),
		MCPServers:       mcpServers,
		SkillDirectories: task.Skills,
	})

	if err != nil {
		return fmt.Errorf("An error occurred while creating a new session: %s", err.Error())
	}

	session.On(func(event copilot.SessionEvent) {
		toWrite, err := eventToLog(event)
		if err != nil {
			log.Printf("An error occurred while converting session event to log: %s\n", err.Error())
			return
		}
		f, err := os.OpenFile(recordFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("An error occurred while opening the log file: %s\n", err.Error())
			return
		}
		defer func() { _ = f.Close() }()
		if _, err = f.WriteString(toWrite + "\n"); err != nil {
			log.Printf("An error occurred while writing to the log file: %s\n", err.Error())
			return
		}
	})

	defer func() { _ = session.Destroy() }()

	response, err := session.SendAndWait(copilot.MessageOptions{Prompt: task.Prompt}, time.Duration(timeout))
	if err != nil {
		log.Printf("An error occurred while sending prompt to session: %s", err.Error())
	}

	if response != nil {
		toWrite, err := eventToLog(*response)
		if err != nil {
			log.Printf("An error occurred while converting session event to log: %s\n", err.Error())
			return err
		}
		f, err := os.OpenFile(recordFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("An error occurred while opening the log file: %s\n", err.Error())
			return err
		}
		defer func() { _ = f.Close() }()
		if _, err = f.WriteString(toWrite + "\n"); err != nil {
			log.Printf("An error occurred while writing to the log file: %s\n", err.Error())
			return err
		}
	}
	return nil
}

func dataToString(data copilot.Data) (string, error) {
	content, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	var v map[string]any
	err = json.Unmarshal(content, &v)
	if err != nil {
		return "", err
	}
	ls := make([]string, 0, len(v))
	for k := range v {
		s := fmt.Sprintf("%s: %v", k, v[k])
		ls = append(ls, s)
	}
	return strings.Join(ls, "; "), nil
}

func eventToLog(event copilot.SessionEvent) (string, error) {
	now := time.Now().Format(time.RFC1123)
	data, err := dataToString(event.Data)
	if err != nil {
		return "", nil
	}
	switch event.Type {
	// Assistant events
	case copilot.AssistantIntent:
		return fmt.Sprintf("%s - Assistant Intent: %s", now, data), nil
	case copilot.AssistantMessage:
		return fmt.Sprintf("%s - Assistant Message: %s", now, data), nil
	case copilot.AssistantMessageDelta:
		return fmt.Sprintf("%s - Assistant Message Delta: %s", now, data), nil
	case copilot.AssistantReasoning:
		return fmt.Sprintf("%s - Assistant Reasoning: %s", now, data), nil
	case copilot.AssistantReasoningDelta:
		return fmt.Sprintf("%s - Assistant Reasoning Delta: %s", now, data), nil
	case copilot.AssistantTurnStart:
		return fmt.Sprintf("%s - Assistant Turn Started", now), nil
	case copilot.AssistantTurnEnd:
		return fmt.Sprintf("%s - Assistant Turn Ended", now), nil
	case copilot.AssistantUsage:
		return fmt.Sprintf("%s - Assistant Usage: %s", now, data), nil

	// Tool execution events
	case copilot.ToolExecutionStart:
		return fmt.Sprintf("%s - Tool Execution Started: %s", now, data), nil
	case copilot.ToolExecutionProgress:
		return fmt.Sprintf("%s - Tool Execution Progress: %s", now, data), nil
	case copilot.ToolExecutionPartialResult:
		return fmt.Sprintf("%s - Tool Execution Partial Result: %s", now, data), nil
	case copilot.ToolExecutionComplete:
		return fmt.Sprintf("%s - Tool Execution Complete: %s", now, data), nil
	case copilot.ToolUserRequested:
		return fmt.Sprintf("%s - Tool User Requested: %s", now, data), nil

	// User events
	case copilot.UserMessage:
		return fmt.Sprintf("%s - User Message: %s", now, data), nil

	// Error and abort events
	case copilot.SessionError:
		return fmt.Sprintf("%s - Session Error: %s", now, data), nil
	case copilot.Abort:
		return fmt.Sprintf("%s - Aborted: %s", now, data), nil

	default:
		return fmt.Sprintf("Event [%s]: %s", event.Type, data), nil
	}
}
