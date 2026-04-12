package ws

import (
	"testing"

	"github.com/mobile-coder/cloud/internal/service"
)

func TestRecordTerminalOutputStoresRecentEventForTask(t *testing.T) {
	hub := NewHub()

	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"running tests\nall green\n"}}`))

	events := hub.GetRecentEvents("dev-1:feature")
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].Summary != "all green" {
		t.Fatalf("events[0].Summary = %q, want %q", events[0].Summary, "all green")
	}
	if events[0].Kind != service.TaskEventKindTestResult {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, service.TaskEventKindTestResult)
	}
}

func TestRecordTerminalOutputIgnoresDuplicateLatestLine(t *testing.T) {
	hub := NewHub()
	message := []byte(`{"type":"terminal_output","payload":{"content":"step 1\nstep 2\n"}}`)

	hub.RecordTerminalOutput("dev-1", "feature", message)
	hub.RecordTerminalOutput("dev-1", "feature", message)

	events := hub.GetRecentEvents("dev-1:feature")
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
}

func TestRecordTerminalOutputStripsAnsiAndKeepsLatestEvents(t *testing.T) {
	hub := NewHub()

	hub.RecordTerminalOutput("dev-1", "feature", []byte("{\"type\":\"terminal_output\",\"payload\":{\"content\":\"\\u001b[32mcompile ok\\u001b[0m\\n\"}}"))
	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"tests passed\n"}}`))

	events := hub.GetRecentEvents("dev-1:feature")
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Summary != "tests passed" {
		t.Fatalf("events[0].Summary = %q, want %q", events[0].Summary, "tests passed")
	}
	if events[1].Summary != "compile ok" {
		t.Fatalf("events[1].Summary = %q, want %q", events[1].Summary, "compile ok")
	}
}

func TestRecordTerminalOutputClassifiesWaitingAndErrorSignals(t *testing.T) {
	hub := NewHub()

	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"Waiting for confirmation before deleting files\n"}}`))
	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"Error: test command failed\n"}}`))

	events := hub.GetRecentEvents("dev-1:feature")
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Kind != service.TaskEventKindError {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, service.TaskEventKindError)
	}
	if events[1].Kind != service.TaskEventKindNeedsInput {
		t.Fatalf("events[1].Kind = %q, want %q", events[1].Kind, service.TaskEventKindNeedsInput)
	}
}

func TestRecordTerminalOutputClassifiesCompletedAndToolStepSignals(t *testing.T) {
	hub := NewHub()

	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"Updating dependencies...\n"}}`))
	hub.RecordTerminalOutput("dev-1", "feature", []byte(`{"type":"terminal_output","payload":{"content":"Task completed successfully\n"}}`))

	events := hub.GetRecentEvents("dev-1:feature")
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Kind != service.TaskEventKindCompleted {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, service.TaskEventKindCompleted)
	}
	if events[1].Kind != service.TaskEventKindToolStep {
		t.Fatalf("events[1].Kind = %q, want %q", events[1].Kind, service.TaskEventKindToolStep)
	}
}
