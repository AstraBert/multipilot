package workflow

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/AstraBert/multipilot/shared"
	copilot "github.com/github/copilot-sdk/go"
)

func TestSerializeEvent(t *testing.T) {
	content := "hello"
	event := copilot.SessionEvent{
		Ephemeral: nil,
		ID:        "123",
		ParentID:  nil,
		Timestamp: time.Date(2026, time.February, 7, 12, 0, 0, 0, time.UTC),
		Type:      "assistant.intent",
		Data: copilot.Data{
			Content:        &content,
			CopilotVersion: nil,
		},
	}
	result, err := serializeEvent(event)
	if err != nil {
		t.Fatalf("Not expecting an error, got %s", err.Error())
	}
	var transformed shared.CopilotEvent
	err = json.Unmarshal([]byte(result), &transformed)
	if err != nil {
		t.Fatalf("Not expecting an error, got %s", err.Error())
	}
	if transformed.ID != "123" || !transformed.Timestamp.Equal(time.Date(2026, time.February, 7, 12, 0, 0, 0, time.UTC)) || transformed.Type != "assistant.intent" {
		t.Fatal("One or more fields of CopilotEvent do not match the original SessionEvent")
	}
	val, ok := transformed.Data["content"]
	if !ok {
		t.Fatal("Expected 'content' to be in the transformed event data, but it is not")
	}
	if val != content {
		t.Fatalf("Expected %s as content, got %v", content, val)
	}
	_, ok = transformed.Data["copilotVersion"]
	if ok {
		t.Fatal("Expected copilotVersion not to be in the transformed event data because is null, but it is")
	}
}
