package components

import (
	"sort"
	"testing"
	"time"

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

func TestSortEventsByTimestamp(t *testing.T) {
	events := []shared.CopilotEvent{
		{Timestamp: time.Date(2026, time.February, 7, 12, 0, 0, 0, time.UTC)},
		{Timestamp: time.Date(2026, time.February, 7, 11, 59, 0, 0, time.UTC)},
		{Timestamp: time.Date(2026, time.February, 7, 12, 1, 0, 0, time.UTC)},
		{Timestamp: time.Date(2026, time.February, 7, 11, 58, 0, 0, time.UTC)},
	}
	sortedEvents := sortEventsByTimestamp(events)
	if !sortedEvents[0].Timestamp.Equal(events[3].Timestamp) || !sortedEvents[1].Timestamp.Equal(events[1].Timestamp) || !sortedEvents[2].Timestamp.Equal(events[0].Timestamp) || !sortedEvents[3].Timestamp.Equal(events[2].Timestamp) {
		t.Fatalf("Unsorted slice: %v", sortedEvents)
	}
}
