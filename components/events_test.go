package components

import (
	"sort"
	"testing"

	"github.com/AstraBert/multipilot/shared"
)

func TestEventTypeToTitle(t *testing.T) {
	testCases := []struct {
		event         shared.CopilotEvent
		expectedTitle string
	}{
		{
			event:         shared.CopilotEvent{Type: "abort"},
			expectedTitle: "Abort",
		},
		{
			event:         shared.CopilotEvent{Type: "assistant.intent"},
			expectedTitle: "Assistant - Intent",
		},
		{
			event:         shared.CopilotEvent{Type: "tool.execution_complete"},
			expectedTitle: "Tool - Execution Complete",
		},
	}
	for _, tc := range testCases {
		title := eventTypeToTitle(tc.event)
		if tc.expectedTitle != title {
			t.Fatalf("Expected %s to be the title, got %s", tc.expectedTitle, title)
		}
	}
}

func TestEventToColor(t *testing.T) {
	for c := range eventTypeColors {
		color := getEventColor(c)
		if eventTypeColors[c] != color {
			t.Fatalf("Expected %s, got %s", eventTypeColors[c], color)
		}
	}
	noColor := getEventColor("not.an.event")
	if noColor != "bg-gray-100 border-gray-300" {
		t.Fatalf("Expected bg-gray-100 border-gray-300, got %s", noColor)
	}
}

func TestEventDataToList(t *testing.T) {
	testCases := []struct {
		name     string
		event    shared.CopilotEvent
		expected []string
	}{
		{
			name: "single key-value pair",
			event: shared.CopilotEvent{
				Data: map[string]any{
					"turnId": "1",
				},
			},
			expected: []string{"turnId: 1"},
		},
		{
			name: "empty data map",
			event: shared.CopilotEvent{
				Data: map[string]any{},
			},
			expected: []string{},
		},
		{
			name: "multiple key-value pairs",
			event: shared.CopilotEvent{
				Data: map[string]any{
					"status": "completed",
					"count":  42,
					"valid":  true,
				},
			},
			expected: []string{"status: completed", "count: 42", "valid: true"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := eventDataToList(tc.event)

			if len(result) != len(tc.expected) {
				t.Fatalf("length mismatch: got %d, want %d", len(result), len(tc.expected))
			}

			sort.Strings(result)
			sort.Strings(tc.expected)

			for i := range result {
				if result[i] != tc.expected[i] {
					t.Errorf("element %d mismatch: got %q, want %q", i, result[i], tc.expected[i])
				}
			}
		})
	}
}
