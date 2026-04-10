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
	deleteNotificationsByIDsFn   func(userID int64, notificationIDs []int64) error

	createCalls      []*db.Notification
	listCalls        []listNotificationsCall
	markReadCalls    []markReadCall
	markAllReadCalls []markAllReadCall
	deleteIDCalls    []deleteIDsCall
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

type deleteIDsCall struct {
	userID int64
	ids    []int64
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

func (f *fakeNotificationStore) DeleteNotificationsByIDs(userID int64, notificationIDs []int64) error {
	ids := append([]int64(nil), notificationIDs...)
	f.deleteIDCalls = append(f.deleteIDCalls, deleteIDsCall{userID: userID, ids: ids})
	if f.deleteNotificationsByIDsFn != nil {
		return f.deleteNotificationsByIDsFn(userID, notificationIDs)
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
	if len(store.createCalls) != 1 {
		t.Fatalf("len(createCalls) = %d, want 1", len(store.createCalls))
	}
	if store.createCalls[0].TaskID != "dev-1:release-train" {
		t.Fatalf("stored TaskID = %q, want %q", store.createCalls[0].TaskID, "dev-1:release-train")
	}
}

func TestNotificationServiceListNotificationsAppliesFilters(t *testing.T) {
	listCalls := 0
	store := &fakeNotificationStore{
		listNotificationsByUserFn: func(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			listCalls++
			if listCalls == 1 {
				if limit != notificationRetentionLimit+1 {
					t.Fatalf("retention limit = %d, want %d", limit, notificationRetentionLimit+1)
				}
				if since != "" {
					t.Fatalf("retention since = %q, want empty", since)
				}
				if unreadOnly {
					t.Fatal("retention unreadOnly = true, want false")
				}
				return nil, nil
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
	if listCalls != 2 {
		t.Fatalf("listCalls = %d, want 2", listCalls)
	}
	if len(store.deleteIDCalls) != 0 {
		t.Fatalf("len(deleteIDCalls) = %d, want 1", len(store.deleteIDCalls))
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

func TestNotificationServiceRetentionKeepsLatest200Notifications(t *testing.T) {
	notifications := make([]db.Notification, 0, 201)
	for i := 201; i >= 1; i-- {
		notifications = append(notifications, db.Notification{
			ID:        int64(i),
			UserID:    7,
			CreatedAt: time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC).Add(time.Duration(-i) * time.Minute).Format(time.RFC3339),
		})
	}

	store := &fakeNotificationStore{
		listNotificationsByUserFn: func(userID int64, limit int, since string, unreadOnly bool) ([]db.Notification, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			if limit != notificationRetentionLimit+1 {
				t.Fatalf("limit = %d, want %d", limit, notificationRetentionLimit+1)
			}
			if since != "" {
				t.Fatalf("since = %q, want empty", since)
			}
			if unreadOnly {
				t.Fatal("unreadOnly = true, want false")
			}
			return notifications, nil
		},
	}

	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time { return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC) }

	if err := service.applyRetention(7, service.now()); err != nil {
		t.Fatalf("applyRetention returned error: %v", err)
	}
	if len(store.deleteIDCalls) != 1 {
		t.Fatalf("len(deleteIDCalls) = %d, want 1", len(store.deleteIDCalls))
	}
	if got := store.deleteIDCalls[0].ids; len(got) != 1 || got[0] != 1 {
		t.Fatalf("delete ids = %v, want [1]", got)
	}
}

func TestNotificationServiceDedupesByConditionOnly(t *testing.T) {
	baseNotification := &db.Notification{
		ID:          11,
		UserID:      7,
		TaskID:      "dev-1:release-train",
		DeviceID:    "dev-1",
		SessionName: "release-train",
		EventType:   string(NotificationEventTaskCompleted),
		Title:       "任务已完成",
		Body:        "release-train 已结束，可查看结果",
		CreatedAt:   "2026-04-11T09:58:30Z",
	}

	var dedupeKeys []string
	store := &fakeNotificationStore{
		getLatestNotificationByKeyFn: func(userID int64, dedupeKey string) (*db.Notification, error) {
			if userID != 7 {
				t.Fatalf("userID = %d, want 7", userID)
			}
			dedupeKeys = append(dedupeKeys, dedupeKey)
			if len(dedupeKeys) == 1 {
				return nil, nil
			}
			return baseNotification, nil
		},
		createNotificationFn: func(notification *db.Notification) (*db.Notification, error) {
			notification.ID = 11
			notification.CreatedAt = "2026-04-11T09:58:30Z"
			return notification, nil
		},
	}

	service := NewNotificationService(nil)
	service.store = store
	service.now = func() time.Time { return time.Date(2026, 4, 11, 10, 0, 0, 0, time.UTC) }

	first, err := service.CreateNotification(7, NotificationEventTaskCompleted, "dev-1:release-train", "dev-1", "release-train", "任务已完成", "release-train 已结束，可查看结果")
	if err != nil {
		t.Fatalf("first CreateNotification returned error: %v", err)
	}
	if first == nil || first.ID != 11 {
		t.Fatalf("first notification = %+v, want ID 11", first)
	}

	second, err := service.CreateNotification(7, NotificationEventTaskCompleted, "dev-1:release-train", "dev-1", "release-train", "完成啦", "你可以回来看结果了")
	if err != nil {
		t.Fatalf("second CreateNotification returned error: %v", err)
	}
	if second == nil || second.ID != 11 {
		t.Fatalf("second notification = %+v, want ID 11", second)
	}
	if len(dedupeKeys) != 2 {
		t.Fatalf("len(dedupeKeys) = %d, want 2", len(dedupeKeys))
	}
	if dedupeKeys[0] != dedupeKeys[1] {
		t.Fatalf("dedupe keys differ: %q vs %q", dedupeKeys[0], dedupeKeys[1])
	}
	if len(store.createCalls) != 1 {
		t.Fatalf("len(createCalls) = %d, want 1", len(store.createCalls))
	}
}
