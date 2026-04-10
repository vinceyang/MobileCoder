package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mobile-coder/cloud/internal/db"
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

type taskNotificationEmitter interface {
	CreateNotification(userID int64, eventType NotificationEventType, taskID, deviceID, sessionName, title, body string) (*db.Notification, error)
}

type TaskService struct {
	source              taskDeviceSource
	eventSource         taskEventSource
	notificationEmitter taskNotificationEmitter
	now                 func() time.Time

	mu                    sync.Mutex
	lastNotificationState map[string]string
}

var ErrTaskNotFound = errors.New("task not found")

func NewTaskService(source taskDeviceSource, eventSource ...taskEventSource) *TaskService {
	service := &TaskService{source: source}
	service.now = time.Now
	service.lastNotificationState = make(map[string]string)
	if len(eventSource) > 0 {
		service.eventSource = eventSource[0]
	}
	if deviceService, ok := source.(*DeviceService); ok && deviceService != nil && deviceService.db != nil {
		service.notificationEmitter = NewNotificationService(deviceService.db)
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
			s.emitTaskNotification(userID, task)
			tasks = append(tasks, task)
		}
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return taskStateRank(tasks[i].State) < taskStateRank(tasks[j].State)
	})

	return tasks, nil
}

func (s *TaskService) RefreshNotificationsForUser(userID int64) error {
	_, err := s.ListTasksForUser(userID)
	return err
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

func (s *TaskService) emitTaskNotification(userID int64, task Task) {
	if s.notificationEmitter == nil {
		return
	}

	eventType, fingerprint, ok := taskNotificationCandidate(task, s.now())
	if !ok {
		return
	}

	key := task.ID + "|" + string(eventType)
	s.mu.Lock()
	previous := s.lastNotificationState[key]
	if previous == fingerprint {
		s.mu.Unlock()
		return
	}
	s.lastNotificationState[key] = fingerprint
	s.mu.Unlock()

	title, body := taskNotificationMessage(task, eventType)
	if title == "" && body == "" {
		return
	}

	_, _ = s.notificationEmitter.CreateNotification(
		userID,
		eventType,
		task.ID,
		task.DeviceID,
		task.SessionName,
		title,
		body,
	)
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

func taskNotificationCandidate(task Task, now time.Time) (NotificationEventType, string, bool) {
	lastActivityAt, err := parseTaskTime(task.LastActivityAt)
	if err != nil {
		lastActivityAt = time.Time{}
	}

	switch {
	case task.State == TaskStateCompleted:
		return NotificationEventTaskCompleted, "completed|" + task.LastActivityAt, true
	case task.State == TaskStateWaiting:
		return NotificationEventTaskWaitingInput, "waiting|" + task.LastActivityAt, true
	case task.State == TaskStateAttention && taskLooksDisconnected(task):
		return NotificationEventAgentDisconnected, "disconnected|" + task.LastActivityAt, true
	case task.State == TaskStateRunning && !lastActivityAt.IsZero() && now.Sub(lastActivityAt.UTC()) >= 15*time.Minute:
		return NotificationEventTaskIdleTooLong, "idle|" + task.LastActivityAt, true
	default:
		return "", "", false
	}
}

func taskNotificationMessage(task Task, eventType NotificationEventType) (string, string) {
	title := task.Title
	if title == "" {
		title = task.SessionName
	}

	switch eventType {
	case NotificationEventTaskCompleted:
		return "任务已完成", fmt.Sprintf("%s 已完成，可以查看结果。", title)
	case NotificationEventTaskWaitingInput:
		return "需要你确认", fmt.Sprintf("%s 正在等待你的输入。", title)
	case NotificationEventTaskIdleTooLong:
		return "任务可能卡住了", fmt.Sprintf("%s 超过 15 分钟没有新输出。", title)
	case NotificationEventAgentDisconnected:
		return "设备已断开", fmt.Sprintf("%s 已离线，任务需要关注。", task.DeviceName)
	default:
		return "", ""
	}
}

func taskLooksDisconnected(task Task) bool {
	lowerReason := strings.ToLower(task.StateReason)
	if strings.Contains(lowerReason, "offline") {
		return true
	}

	lowerRecentEvent := strings.ToLower(task.RecentEvent)
	return strings.Contains(lowerRecentEvent, "offline")
}

func parseTaskTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
	}

	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}

	return time.Time{}, lastErr
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
