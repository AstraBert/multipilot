package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/AstraBert/multipilot/shared"
	"github.com/AstraBert/multipilot/workflow"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func ReadConfigToTasks(configFile string) (*shared.CopilotTasks, error) {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var tasks shared.CopilotTasks
	err = json.Unmarshal(content, &tasks)
	if err != nil {
		return nil, err
	}
	err = tasks.Validate()
	if err != nil {
		return nil, err
	}
	return &tasks, nil
}

func RunCopilotWorkflow(input shared.CopilotInput) error {
	c, err := client.Dial(client.Options{})

	if err != nil {
		log.Println("Unable to create Temporal client:", err)
		return err
	}

	defer c.Close()

	workflowId := "multipilot-" + uuid.New().String()

	options := client.StartWorkflowOptions{
		ID:        workflowId,
		TaskQueue: workflow.CopilotTaskQueue,
	}

	log.Printf("Assigning task with prompt %s and cwd %s to workflow with ID %s", input.Prompt, input.Cwd, workflowId)

	we, err := c.ExecuteWorkflow(context.Background(), options, workflow.CopilotWorkflow, input)
	if err != nil {
		log.Println("Unable to start the Workflow:", err)
		return err
	}

	log.Printf("Workflow Run ID: %s\n", we.GetRunID())

	var result error

	err = we.Get(context.Background(), &result)

	if err != nil {
		log.Println("Unable to get Workflow result:", err)
		return err
	}

	if result != nil {
		log.Println("Error during Copilot execution:", result)
		return result
	}
	return nil
}

func LoadEvents(logFile string) ([]shared.CopilotEvent, error) {
	content, err := os.ReadFile(logFile)
	if err != nil {
		return nil, err
	}
	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")
	events := make([]shared.CopilotEvent, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\n")
		if line != "" {
			var event shared.CopilotEvent
			err := json.Unmarshal([]byte(line), &event)
			if err != nil {
				return nil, err
			}
			events = append(events, event)
		}
	}
	return events, nil
}
