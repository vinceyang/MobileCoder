package service

import "testing"

type fakeTaskDeviceSource struct {
	devicesByUser    map[int64][]Device
	sessionsByDevice map[string][]Session
}

type fakeTaskEventSource struct {
	eventsByTask map[string][]TaskEvent
}

func (f *fakeTaskDeviceSource) GetUserDevices(userID int64) ([]Device, error) {
	return f.devicesByUser[userID], nil
}

func (f *fakeTaskDeviceSource) GetDeviceSessions(deviceID string) ([]Session, error) {
	return f.sessionsByDevice[deviceID], nil
}

func (f *fakeTaskEventSource) GetRecentEvents(taskID string) []TaskEvent {
	return f.eventsByTask[taskID]
}

func TestTaskServiceListTasksForUser(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online", LastActiveAt: "2026-04-10T09:30:00Z"},
				{DeviceID: "dev-2", DeviceName: "Mac mini", Status: "offline", LastActiveAt: "2026-04-10T08:00:00Z"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "feature-branch", ProjectPath: "/Users/me/repo-a", Status: "active", CreatedAt: "2026-04-10T09:00:00Z"},
			},
			"dev-2": {
				{ID: 11, DeviceID: "dev-2", SessionName: "nightly", ProjectPath: "/Users/me/repo-b", Status: "inactive", CreatedAt: "2026-04-10T07:30:00Z"},
			},
		},
	}, &fakeTaskEventSource{
		eventsByTask: map[string][]TaskEvent{
			"dev-1:feature-branch": {
				{Summary: "all green", Timestamp: "2026-04-10T09:31:00Z", Kind: TaskEventKindTestResult},
			},
		},
	})

	tasks, err := service.ListTasksForUser(7)
	if err != nil {
		t.Fatalf("ListTasksForUser returned error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("len(tasks) = %d, want 2", len(tasks))
	}

	if tasks[0].State != TaskStateRunning {
		t.Fatalf("tasks[0].State = %q, want %q", tasks[0].State, TaskStateRunning)
	}
	if tasks[0].Title == "" {
		t.Fatal("tasks[0].Title should not be empty")
	}
	if tasks[0].Summary == "" {
		t.Fatal("tasks[0].Summary should not be empty")
	}
	if tasks[0].LastActivityAt != "2026-04-10T09:31:00Z" {
		t.Fatalf("tasks[0].LastActivityAt = %q, want %q", tasks[0].LastActivityAt, "2026-04-10T09:31:00Z")
	}
	if tasks[0].StateReason == "" {
		t.Fatal("tasks[0].StateReason should not be empty")
	}
	if tasks[0].RecentEvent == "" {
		t.Fatal("tasks[0].RecentEvent should not be empty")
	}
	if len(tasks[0].Timeline) != 1 {
		t.Fatalf("len(tasks[0].Timeline) = %d, want 1", len(tasks[0].Timeline))
	}
	if tasks[0].RecentEvent != "all green" {
		t.Fatalf("tasks[0].RecentEvent = %q, want %q", tasks[0].RecentEvent, "all green")
	}

	if tasks[1].State != TaskStateAttention {
		t.Fatalf("tasks[1].State = %q, want %q", tasks[1].State, TaskStateAttention)
	}
	if tasks[1].StateReason != "Device is offline" {
		t.Fatalf("tasks[1].StateReason = %q, want %q", tasks[1].StateReason, "Device is offline")
	}
}

func TestTaskServiceGetTaskForUser(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "feature-branch", ProjectPath: "/Users/me/repo-a", Status: "active"},
			},
		},
	}, &fakeTaskEventSource{})

	task, err := service.GetTaskForUser(7, "dev-1:feature-branch")
	if err != nil {
		t.Fatalf("GetTaskForUser returned error: %v", err)
	}
	if task == nil {
		t.Fatal("GetTaskForUser returned nil task")
	}
	if task.ID != "dev-1:feature-branch" {
		t.Fatalf("task.ID = %q, want %q", task.ID, "dev-1:feature-branch")
	}
}

func TestTaskServiceDerivesWaitingStateAndFallbackActivity(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "review", ProjectPath: "/Users/me/repo-a", Status: "waiting_input", CreatedAt: "2026-04-10T09:00:00Z"},
			},
		},
	}, &fakeTaskEventSource{})

	tasks, err := service.ListTasksForUser(7)
	if err != nil {
		t.Fatalf("ListTasksForUser returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].State != TaskStateWaiting {
		t.Fatalf("tasks[0].State = %q, want %q", tasks[0].State, TaskStateWaiting)
	}
	if tasks[0].LastActivityAt != "2026-04-10T09:00:00Z" {
		t.Fatalf("tasks[0].LastActivityAt = %q, want %q", tasks[0].LastActivityAt, "2026-04-10T09:00:00Z")
	}
	if tasks[0].StateReason != "Agent is waiting for input" {
		t.Fatalf("tasks[0].StateReason = %q, want %q", tasks[0].StateReason, "Agent is waiting for input")
	}
}

func TestTaskServiceUsesRecentEventKindToOverrideState(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "review", ProjectPath: "/Users/me/repo-a", Status: "active", CreatedAt: "2026-04-10T09:00:00Z"},
			},
		},
	}, &fakeTaskEventSource{
		eventsByTask: map[string][]TaskEvent{
			"dev-1:review": {
				{Summary: "Waiting for confirmation", Timestamp: "2026-04-10T09:05:00Z", Kind: TaskEventKindNeedsInput},
			},
		},
	})

	tasks, err := service.ListTasksForUser(7)
	if err != nil {
		t.Fatalf("ListTasksForUser returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].State != TaskStateWaiting {
		t.Fatalf("tasks[0].State = %q, want %q", tasks[0].State, TaskStateWaiting)
	}
	if tasks[0].StateReason != "Recent output indicates user input is required" {
		t.Fatalf("tasks[0].StateReason = %q, want %q", tasks[0].StateReason, "Recent output indicates user input is required")
	}
}

func TestTaskServiceUsesErrorEventKindForAttentionState(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "tests", ProjectPath: "/Users/me/repo-a", Status: "active", CreatedAt: "2026-04-10T09:00:00Z"},
			},
		},
	}, &fakeTaskEventSource{
		eventsByTask: map[string][]TaskEvent{
			"dev-1:tests": {
				{Summary: "Error: test command failed", Timestamp: "2026-04-10T09:06:00Z", Kind: TaskEventKindError},
			},
		},
	})

	tasks, err := service.ListTasksForUser(7)
	if err != nil {
		t.Fatalf("ListTasksForUser returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].State != TaskStateAttention {
		t.Fatalf("tasks[0].State = %q, want %q", tasks[0].State, TaskStateAttention)
	}
	if tasks[0].StateReason != "Recent output indicates a failure or blocked step" {
		t.Fatalf("tasks[0].StateReason = %q, want %q", tasks[0].StateReason, "Recent output indicates a failure or blocked step")
	}
}

func TestTaskServiceUsesCompletedEventKindForCompletedState(t *testing.T) {
	service := NewTaskService(&fakeTaskDeviceSource{
		devicesByUser: map[int64][]Device{
			7: {
				{DeviceID: "dev-1", DeviceName: "MacBook", Status: "online"},
			},
		},
		sessionsByDevice: map[string][]Session{
			"dev-1": {
				{ID: 10, DeviceID: "dev-1", SessionName: "ship", ProjectPath: "/Users/me/repo-a", Status: "active", CreatedAt: "2026-04-10T09:00:00Z"},
			},
		},
	}, &fakeTaskEventSource{
		eventsByTask: map[string][]TaskEvent{
			"dev-1:ship": {
				{Summary: "Task completed successfully", Timestamp: "2026-04-10T09:10:00Z", Kind: TaskEventKindCompleted},
			},
		},
	})

	tasks, err := service.ListTasksForUser(7)
	if err != nil {
		t.Fatalf("ListTasksForUser returned error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(tasks))
	}
	if tasks[0].State != TaskStateCompleted {
		t.Fatalf("tasks[0].State = %q, want %q", tasks[0].State, TaskStateCompleted)
	}
	if tasks[0].StateReason != "Recent output indicates the task has finished" {
		t.Fatalf("tasks[0].StateReason = %q, want %q", tasks[0].StateReason, "Recent output indicates the task has finished")
	}
}
