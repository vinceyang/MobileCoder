package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mobile-coder/cloud/internal/db"
)

const (
	notificationRetentionLimit = 200
	notificationDedupeWindow   = 5 * time.Minute
	notificationDefaultLimit   = 50
	notificationMaxLimit       = 100
)

var (
	ErrInvalidNotificationEventType = errors.New("invalid notification event type")
	ErrNotificationNotFound         = errors.New("notification not found")
	ErrNotificationAccessDenied     = errors.New("notification access denied")
)

type NotificationEventType string

const (
	NotificationEventTaskCompleted     NotificationEventType = "task_completed"
	NotificationEventTaskWaitingInput  NotificationEventType = "task_waiting_for_input"
	NotificationEventTaskIdleTooLong   NotificationEventType = "task_idle_too_long"
	NotificationEventAgentDisconnected NotificationEventType = "agent_disconnected"
)

var allowedNotificationEventTypes = map[NotificationEventType]struct{}{
	NotificationEventTaskCompleted:     {},
	NotificationEventTaskWaitingInput:  {},
	NotificationEventTaskIdleTooLong:   {},
	NotificationEventAgentDisconnected: {},
}

func (t NotificationEventType) IsValid() bool {
	_, ok := allowedNotificationEventTypes[t]
	return ok
}

type notificationStore interface {
	CreateNotification(*db.Notification) (*db.Notification, error)
	ListNotificationsByUser(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error)
	GetNotificationByID(notificationID int64) (*db.Notification, error)
	GetLatestNotificationByDedupeKey(userID int64, dedupeKey string) (*db.Notification, error)
	MarkNotificationRead(notificationID int64, readAt string) error
	MarkAllNotificationsRead(userID int64, readAt string) error
	DeleteNotificationsByIDs(userID int64, notificationIDs []int64) error
}

type NotificationService struct {
	store notificationStore
	now   func() time.Time

	dedupeMu    sync.Mutex
	dedupeLocks map[string]*notificationLockEntry
}

type notificationLockEntry struct {
	mu   sync.Mutex
	refs int
}

func NewNotificationService(database *db.SupabaseDB) *NotificationService {
	return &NotificationService{
		store:       database,
		now:         time.Now,
		dedupeLocks: make(map[string]*notificationLockEntry),
	}
}

func (s *NotificationService) CreateNotification(userID int64, eventType NotificationEventType, taskID, deviceID, sessionName, title, body string) (*db.Notification, error) {
	if !eventType.IsValid() {
		return nil, ErrInvalidNotificationEventType
	}

	dedupeKey := buildNotificationDedupeKey(userID, eventType, taskID, deviceID, sessionName)
	release := s.acquireDedupeLock(dedupeKey)
	defer release()

	now := s.now().UTC()
	if err := s.applyRetention(userID); err != nil {
		return nil, err
	}

	if existing, err := s.store.GetLatestNotificationByDedupeKey(userID, dedupeKey); err != nil {
		return nil, err
	} else if existing != nil && notificationIsRecent(existing, now) {
		return existing, nil
	}

	created, err := s.store.CreateNotification(&db.Notification{
		UserID:      userID,
		TaskID:      taskID,
		DeviceID:    deviceID,
		SessionName: sessionName,
		EventType:   string(eventType),
		Title:       title,
		Body:        body,
		DedupeKey:   dedupeKey,
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *NotificationService) ListNotifications(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error) {
	if limit <= 0 {
		limit = notificationDefaultLimit
	}
	if limit > notificationMaxLimit {
		limit = notificationMaxLimit
	}

	if err := s.applyRetention(userID); err != nil {
		return nil, err
	}

	sinceValue := ""
	if since != nil && !since.IsZero() {
		sinceValue = since.UTC().Format(time.RFC3339)
	}

	return s.store.ListNotificationsByUser(userID, limit, sinceValue, unreadOnly)
}

func (s *NotificationService) MarkNotificationRead(userID, notificationID int64) error {
	notification, err := s.store.GetNotificationByID(notificationID)
	if err != nil {
		return err
	}
	if notification == nil {
		return ErrNotificationNotFound
	}
	if notification.UserID != userID {
		return ErrNotificationAccessDenied
	}

	return s.store.MarkNotificationRead(notificationID, s.now().UTC().Format(time.RFC3339))
}

func (s *NotificationService) MarkAllNotificationsRead(userID int64) error {
	return s.store.MarkAllNotificationsRead(userID, s.now().UTC().Format(time.RFC3339))
}

func (s *NotificationService) applyRetention(userID int64) error {
	notifications, err := s.store.ListNotificationsByUser(userID, 0, "", false)
	if err != nil {
		return err
	}
	if len(notifications) <= notificationRetentionLimit {
		return nil
	}

	overflow := notifications[notificationRetentionLimit:]
	ids := make([]int64, 0, len(overflow))
	for _, notification := range overflow {
		if notification.ID > 0 {
			ids = append(ids, notification.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	return s.store.DeleteNotificationsByIDs(userID, ids)
}

func (s *NotificationService) acquireDedupeLock(dedupeKey string) func() {
	s.dedupeMu.Lock()
	entry := s.dedupeLocks[dedupeKey]
	if entry == nil {
		entry = &notificationLockEntry{}
		s.dedupeLocks[dedupeKey] = entry
	}
	entry.refs++
	s.dedupeMu.Unlock()

	entry.mu.Lock()
	return func() {
		entry.mu.Unlock()

		s.dedupeMu.Lock()
		entry.refs--
		if entry.refs == 0 {
			delete(s.dedupeLocks, dedupeKey)
		}
		s.dedupeMu.Unlock()
	}
}

func notificationIsRecent(notification *db.Notification, now time.Time) bool {
	if notification == nil || notification.CreatedAt == "" {
		return false
	}

	createdAt, err := parseNotificationTime(notification.CreatedAt)
	if err != nil {
		return false
	}

	return now.Sub(createdAt.UTC()) <= notificationDedupeWindow
}

func buildNotificationDedupeKey(userID int64, eventType NotificationEventType, taskID, deviceID, sessionName string) string {
	parts := []string{
		fmt.Sprintf("%d", userID),
		string(eventType),
		normalizeNotificationKeyPart(taskID),
		normalizeNotificationKeyPart(deviceID),
		normalizeNotificationKeyPart(sessionName),
	}
	return strings.Join(parts, "|")
}

func normalizeNotificationKeyPart(value string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(value))), " ")
}

func parseNotificationTime(value string) (time.Time, error) {
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
