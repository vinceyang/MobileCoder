package service

import (
	"errors"
	"testing"
	"time"

	"github.com/mobile-coder/cloud/internal/db"
)

type fakeNotificationStore struct {
	createNotificationFn         func(*db.Notification) (*db.Notification, error)
	listNotificationsByUserFn    func(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error)
	getNotificationByIDFn        func(notificationID int64) (*db.Notification, error)
	getLatestNotificationByKeyFn func(userID int64, dedupeKey string) (*db.Notification, error)
	markNotificationReadFn       func(notificationID int64, readAt string) error
	markAllNotificationsReadFn   func(userID int64, readAt string) error
	deleteNotificationsBeforeFn  func(userID int64, cutoff string) error

	createCalls       []*db.Notification
	listCalls         []listNotificationsCall
	markReadCalls     []markReadCall
	markAllReadCalls  []markAllReadCall
	deleteBeforeCalls []deleteBeforeCall
}

type listNotificationsCall struct {
	userID     int64
	limit      int
	since      string
	unreadOnly bool
}

type markReadCall struct {
	notificationID int64
	readAt         string
}

type markAllReadCall struct {
	userID int64
	readAt string
}

type deleteBeforeCall struct {
	userID int64
	cutoff string
}

func (f *fakeNotificationStore) CreateNotification(notification *db.Notification) (*db.Notification, error) {
	f.createCalls = append(f.createCalls, notification)
	if f.createNotificationFn != nil {
		return f.createNotificationFn(notification)
	}
	return notification, nil
}

func (f *fakeNotificationStore) ListNotificationsByUser(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error) {
	f.listCalls = append(f.listCalls, listNotificationsCall{userID: userID, limit: limit, since: since, unreadOnly: unreadOnly})
	if f.listNotificationsByUserFn != nil {
		return f.listNotificationsByUserFn(userID, limit, since, unreadOnly)
	}
	return nil, nil
}

func (f *fakeNotificationStore) GetNotificationByID(notificationID int64) (*db.Notification, error) {
	if f.getNotificationByIDFn != nil {
		return f.getNotificationByIDFn(notificationID)
	}
	return nil, nil
}

func (f *fakeNotificationStore) GetLatestNotificationByDedupeKey(userID int64, dedupeKey string) (*db.Notification, error) {
	if f.getLatestNotificationByKeyFn != nil {
		return f.getLatestNotificationByKeyFn(userID, dedupeKey)
	}
	return nil, nil
}

func (f *fakeNotificationStore) MarkNotificationRead(notificationID int64, readAt string) error {
	f.markReadCalls = append(f.markReadCalls, markReadCall{notificationID: notificationID, readAt: readAt})
	if f.markNotificationReadFn != nil {
		return f.markNotificationReadFn(notificationID, readAt)
	}
	return nil
}

func (f *fakeNotificationStore) MarkAllNotificationsRead(userID int64, readAt string) error {
	f.markAllReadCalls = append(f.markAllReadCalls, markAllReadCall{userID: userID, readAt: readAt})
	if f.markAllNotificationsReadFn != nil {
		return f.markAllNotificationsReadFn(userID, readAt)
	}
	return nil
}

func (f *fakeNotificationStore) DeleteNotificationsBefore(userID int64, cutoff string) error {
	f.deleteBeforeCalls = append(f.deleteBeforeCalls, deleteBeforeCall{userID: userID, cutoff: cutoff})
	if f.deleteNotificationsBeforeFn != nil {
		return f.deleteNotificationsBeforeFn(userID, cutoff)
	}
	return nil
}

func TestNotificationServiceCreateNotificationPersistsAndRetains(t *testing.T) {
	store := &fakeNotificationStore{
		createNotificationFn: func(notification *db.Notification) (*db.Notification, error) {
			notification.ID = 99
			notification.CreatedAt = "2026-04-11T10:00:00Z"
			return notification, nil
		},
	}
	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time {
		return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	}

	notification, err := service.CreateNotification(
		7,
		NotificationEventTaskCompleted,
		"dev-1:release-train",
		"dev-1",
		"release-train",
		"任务已完成",
		"release-train 已结束，可查看结果",
	)
	if err != nil {
		t.Fatalf("CreateNotification returned error: %v", err)
	}
	if notification == nil {
		t.Fatal("CreateNotification returned nil notification")
	}
	if notification.ID != 99 {
		t.Fatalf("notification.ID = %d, want 99", notification.ID)
	}
	if notification.EventType != string(NotificationEventTaskCompleted) {
		t.Fatalf("notification.EventType = %q, want %q", notification.EventType, NotificationEventTaskCompleted)
	}
	if notification.DedupeKey == "" {
		t.Fatal("notification.DedupeKey should not be empty")
	}
	if len(store.deleteBeforeCalls) != 1 {
		t.Fatalf("len(deleteBeforeCalls) = %d, want 1", len(store.deleteBeforeCalls))
	}
	if store.deleteBeforeCalls[0].userID != 7 {
		t.Fatalf("deleteBeforeCalls[0].userID = %d, want 7", store.deleteBeforeCalls[0].userID)
	}
	if len(store.createCalls) != 1 {
		t.Fatalf("len(createCalls) = %d, want 1", len(store.createCalls))
	}
	if store.createCalls[0].TaskID != "dev-1:release-train" {
		t.Fatalf("stored TaskID = %q, want %q", store.createCalls[0].TaskID, "dev-1:release-train")
	}
}

func TestNotificationServiceCreateNotificationDedupesRecentRecord(t *testing.T) {
	existing := &db.Notification{
		ID:        11,
		UserID:    7,
		TaskID:    "dev-1:release-train",
		EventType: string(NotificationEventTaskCompleted),
		Title:     "任务已完成",
		Body:      "release-train 已结束，可查看结果",
		DedupeKey: buildNotificationDedupeKey(7, NotificationEventTaskCompleted, "dev-1:release-train", "dev-1", "release-train", "任务已完成", "release-train 已结束，可查看结果"),
		CreatedAt: "2026-04-11T09:58:30Z",
	}
	store := &fakeNotificationStore{
		getLatestNotificationByKeyFn: func(userID int64, dedupeKey string) (*db.Notification, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			if dedupeKey != existing.DedupeKey {
				t.Fatalf("dedupeKey = %q, want %q", dedupeKey, existing.DedupeKey)
			}
			return existing, nil
		},
	}
	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time {
		return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	}

	notification, err := service.CreateNotification(
		7,
		NotificationEventTaskCompleted,
		"dev-1:release-train",
		"dev-1",
		"release-train",
		"任务已完成",
		"release-train 已结束，可查看结果",
	)
	if err != nil {
		t.Fatalf("CreateNotification returned error: %v", err)
	}
	if notification == nil {
		t.Fatal("CreateNotification returned nil notification")
	}
	if notification.ID != existing.ID {
		t.Fatalf("notification.ID = %d, want %d", notification.ID, existing.ID)
	}
	if len(store.createCalls) != 0 {
		t.Fatalf("len(createCalls) = %d, want 0", len(store.createCalls))
	}
}

func TestNotificationServiceListNotificationsAppliesFilters(t *testing.T) {
	store := &fakeNotificationStore{
		listNotificationsByUserFn: func(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			if limit != 100 {
				t.Fatalf("limit = %d, want 100", limit)
			}
			if since != "2026-04-11T09:00:00Z" {
				t.Fatalf("since = %q, want %q", since, "2026-04-11T09:00:00Z")
			}
			if !unreadOnly {
				t.Fatal("unreadOnly = false, want true")
			}
			return []db.Notification{{ID: 1, UserID: 7}}, nil
		},
	}
	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time {
		return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	}
	since := time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC)

	notifications, err := service.ListNotifications(7, true, &since, 250)
	if err != nil {
		t.Fatalf("ListNotifications returned error: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(notifications))
	}
	if len(store.deleteBeforeCalls) != 1 {
		t.Fatalf("len(deleteBeforeCalls) = %d, want 1", len(store.deleteBeforeCalls))
	}
}

func TestNotificationServiceMarkNotificationRead(t *testing.T) {
	store := &fakeNotificationStore{
		getNotificationByIDFn: func(notificationID int64) (*db.Notification, error) {
			if notificationID != 88 {
				t.Fatalf("notificationID = %d, want 88", notificationID)
			}
			return &db.Notification{ID: 88, UserID: 7}, nil
		},
		markNotificationReadFn: func(notificationID int64, readAt string) error {
			if notificationID != 88 {
				t.Fatalf("notificationID = %d, want 88", notificationID)
			}
			if readAt != "2026-04-11T10:00:00Z" {
				t.Fatalf("readAt = %q, want %q", readAt, "2026-04-11T10:00:00Z")
			}
			return nil
		},
	}
	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time {
		return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC)
	}

	if err := service.MarkNotificationRead(7, 88); err != nil {
		t.Fatalf("MarkNotificationRead returned error: %v", err)
	}
	if len(store.markReadCalls) != 1 {
		t.Fatalf("len(markReadCalls) = %d, want 1", len(store.markReadCalls))
	}
}

func TestNotificationServiceMarkNotificationReadRejectsForeignNotification(t *testing.T) {
	store := &fakeNotificationStore{
		getNotificationByIDFn: func(notificationID int64) (*db.Notification, error) {
			return &db.Notification{ID: notificationID, UserID: 99}, nil
		},
	}
	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time { return time.Unix(0, 0).UTC() }

	err := service.MarkNotificationRead(7, 88)
	if !errors.Is(err, ErrNotificationAccessDenied) {
		t.Fatalf("error = %v, want ErrNotificationAccessDenied", err)
	}
}

func TestNotificationServiceRejectsInvalidEventType(t *testing.T) {
	store := &fakeNotificationStore{}
	service := NewNotificationService(nil)
	service.store = store

	_, err := service.CreateNotification(7, NotificationEventType("bogus"), "dev-1:tests", "dev-1", "tests", "标题", "正文")
	if !errors.Is(err, ErrInvalidNotificationEventType) {
		t.Fatalf("error = %v, want ErrInvalidNotificationEventType", err)
	}
	if len(store.createCalls) != 0 {
		t.Fatalf("len(createCalls) = %d, want 0", len(store.createCalls))
	}
}
