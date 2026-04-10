package service

import (
	"errors"
	"path/filepath"
	"sort"
	"strings"
)

type TaskState string

const (
	TaskStateRunning   TaskState = "running"
	TaskStateWaiting   TaskState = "waiting"
	TaskStateCompleted TaskState = "completed"
	TaskStateAttention TaskState = "attention"
)

type Task struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	DeviceID       string      `json:"device_id"`
	DeviceName     string      `json:"device_name"`
	SessionName    string      `json:"session_name"`
	ProjectPath    string      `json:"project_path"`
	Tool           string      `json:"tool"`
	State          TaskState   `json:"state"`
	Summary        string      `json:"summary"`
	StateReason    string      `json:"state_reason"`
	RecentEvent    string      `json:"recent_event"`
	LastActivityAt string      `json:"last_activity_at"`
	Timeline       []TaskEvent `json:"timeline,omitempty"`
}

type TaskEvent struct {
	Summary   string        `json:"summary"`
	Timestamp string        `json:"timestamp"`
	Kind      TaskEventKind `json:"kind"`
}

type TaskEventKind string

const (
	TaskEventKindInfo       TaskEventKind = "info"
	TaskEventKindNeedsInput TaskEventKind = "needs_input"
	TaskEventKindError      TaskEventKind = "error"
	TaskEventKindTestResult TaskEventKind = "test_result"
	TaskEventKindCompleted  TaskEventKind = "completed"
	TaskEventKindToolStep   TaskEventKind = "tool_step"
)

type taskDeviceSource interface {
	GetUserDevices(userID int64) ([]Device, error)
	GetDeviceSessions(deviceID string) ([]Session, error)
}

type taskEventSource interface {
	GetRecentEvents(taskID string) []TaskEvent
}

type TaskService struct {
	source      taskDeviceSource
	eventSource taskEventSource
}

var ErrTaskNotFound = errors.New("task not found")

func NewTaskService(source taskDeviceSource, eventSource ...taskEventSource) *TaskService {
	service := &TaskService{source: source}
	if len(eventSource) > 0 {
		service.eventSource = eventSource[0]
	}
	return service
}

func (s *TaskService) ListTasksForUser(userID int64) ([]Task, error) {
	devices, err := s.source.GetUserDevices(userID)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, device := range devices {
		sessions, err := s.source.GetDeviceSessions(device.DeviceID)
		if err != nil {
			return nil, err
		}
		for _, session := range sessions {
			task := s.mapSessionToTask(device, session)
			tasks = append(tasks, task)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return taskStateRank(tasks[i].State) < taskStateRank(tasks[j].State)
	})

	return tasks, nil
}

func (s *TaskService) GetTaskForUser(userID int64, taskID string) (*Task, error) {
	tasks, err := s.ListTasksForUser(userID)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == taskID {
			taskCopy := task
			return &taskCopy, nil
		}
	}

	return nil, ErrTaskNotFound
}

func taskStateRank(state TaskState) int {
	switch state {
	case TaskStateRunning:
		return 0
	case TaskStateWaiting:
		return 1
	case TaskStateAttention:
		return 2
	default:
		return 3
	}
}

func mapSessionToTask(device Device, session Session) Task {
	taskID := session.DeviceID + ":" + session.SessionName
	state := deriveTaskState(device, session)
	task := Task{
		ID:             taskID,
		Title:          deriveTaskTitle(device, session),
		DeviceID:       device.DeviceID,
		DeviceName:     device.DeviceName,
		SessionName:    session.SessionName,
		ProjectPath:    session.ProjectPath,
		Tool:           "claude",
		State:          state,
		Summary:        deriveTaskSummary(state, session),
		StateReason:    deriveTaskStateReason(state),
		RecentEvent:    deriveTaskRecentEvent(state, device, session),
		LastActivityAt: deriveLastActivityAt(device, session),
	}
	return task
}

func (s *TaskService) enrichTask(task Task) Task {
	if s.eventSource == nil {
		return task
	}

	timeline := s.eventSource.GetRecentEvents(task.ID)
	if len(timeline) == 0 {
		return task
	}

	task.Timeline = timeline
	task.RecentEvent = timeline[0].Summary
	if timeline[0].Timestamp != "" {
		task.LastActivityAt = timeline[0].Timestamp
	}
	task.State = deriveTaskStateFromEvent(task.State, timeline[0])
	task.StateReason = deriveTaskStateReasonFromEvent(task.StateReason, timeline[0])
	return task
}

func (s *TaskService) mapSessionToTask(device Device, session Session) Task {
	return s.enrichTask(mapSessionToTask(device, session))
}

func deriveTaskState(device Device, session Session) TaskState {
	if device.Status != "online" {
		return TaskStateAttention
	}

	if strings.Contains(strings.ToLower(session.Status), "waiting") {
		return TaskStateWaiting
	}

	if strings.EqualFold(session.Status, "active") {
		return TaskStateRunning
	}

	return TaskStateCompleted
}

func deriveTaskTitle(device Device, session Session) string {
	if session.SessionName != "" {
		return session.SessionName
	}
	if session.ProjectPath != "" {
		return filepath.Base(session.ProjectPath)
	}
	return device.DeviceName
}

func deriveTaskSummary(state TaskState, session Session) string {
	switch state {
	case TaskStateRunning:
		if session.ProjectPath != "" {
			return "Running in " + session.ProjectPath
		}
		return "Task is currently running"
	case TaskStateWaiting:
		return "Waiting for user input"
	case TaskStateAttention:
		return "Device offline, check connection"
	default:
		return "Session inactive"
	}
}

func deriveTaskStateReason(state TaskState) string {
	switch state {
	case TaskStateRunning:
		return "Session is actively running"
	case TaskStateWaiting:
		return "Agent is waiting for input"
	case TaskStateAttention:
		return "Device is offline"
	default:
		return "Session is inactive"
	}
}

func deriveTaskRecentEvent(state TaskState, device Device, session Session) string {
	switch state {
	case TaskStateRunning:
		return "Session " + session.SessionName + " is active on " + device.DeviceName
	case TaskStateWaiting:
		return "Agent paused and is waiting for the next instruction"
	case TaskStateAttention:
		return device.DeviceName + " went offline during this task"
	default:
		return "Session " + session.SessionName + " is no longer active"
	}
}

func deriveLastActivityAt(device Device, session Session) string {
	if device.LastActiveAt != "" {
		return device.LastActiveAt
	}
	return session.CreatedAt
}

func deriveTaskStateFromEvent(current TaskState, event TaskEvent) TaskState {
	switch event.Kind {
	case TaskEventKindNeedsInput:
		return TaskStateWaiting
	case TaskEventKindError:
		return TaskStateAttention
	case TaskEventKindCompleted:
		return TaskStateCompleted
	default:
		return current
	}
}

func deriveTaskStateReasonFromEvent(current string, event TaskEvent) string {
	switch event.Kind {
	case TaskEventKindNeedsInput:
		return "Recent output indicates user input is required"
	case TaskEventKindError:
		return "Recent output indicates a failure or blocked step"
	case TaskEventKindCompleted:
		return "Recent output indicates the task has finished"
	default:
		return current
	}
}
