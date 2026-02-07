package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
	tok, err := task.GetToken()
	if err != nil {
		return err
	}
	if tok != "" {
		options.GithubToken = tok
	}
	client := copilot.NewClient(options)
	if err := client.Start(); err != nil {
		return fmt.Errorf("an error occurred while starting the client: %s", err.Error())
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
		ExcludedTools:    task.ExcludeTools,
		SystemMessage:    (*copilot.SystemMessageConfig)(systemPrompt),
		MCPServers:       mcpServers,
		SkillDirectories: task.Skills,
	})

	if err != nil {
		return fmt.Errorf("an error occurred while creating a new session: %s", err.Error())
	}

	seenIds := make(map[string]int8)

	session.On(func(event copilot.SessionEvent) {
		_, ok := seenIds[event.ID]
		if ok {
			return
		}
		seenIds[event.ID] = 0
		toWrite, err := serializeEvent(event)
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

	response, err := session.SendAndWait(copilot.MessageOptions{Prompt: task.Prompt}, time.Duration(timeout)*time.Second)
	if err != nil {
		log.Printf("An error occurred while sending prompt to session: %s", err.Error())
	}

	if response != nil {
		toWrite, err := serializeEvent(*response)
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

func serializeEvent(event copilot.SessionEvent) (string, error) {
	transformed := shared.CopilotEvent{ID: event.ID, Timestamp: event.Timestamp, Type: string(event.Type), Data: make(map[string]any)}
	content, err := json.Marshal(event.Data)
	if err != nil {
		return "", err
	}
	var m map[string]any
	err = json.Unmarshal(content, &m)
	if err != nil {
		return "", err
	}
	for k := range m {
		if m[k] != nil {
			transformed.Data[k] = m[k]
		}
	}
	serialized, err := json.Marshal(transformed)
	if err != nil {
		return "", err
	}
	return string(serialized), nil
}
