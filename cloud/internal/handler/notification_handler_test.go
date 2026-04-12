package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/service"
)

type fakeNotificationService struct {
	listNotificationsFn        func(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error)
	markNotificationReadFn     func(userID, notificationID int64) error
	markAllNotificationsReadFn func(userID int64) error

	listCalls     []notificationListCall
	markReadCalls []notificationReadCall
	markAllCalls  []int64
}

type fakeNotificationRefresher struct {
	refreshFn    func(userID int64) error
	refreshCalls []int64
}

type notificationListCall struct {
	userID     int64
	unreadOnly bool
	since      *time.Time
	limit      int
}

type notificationReadCall struct {
	userID         int64
	notificationID int64
}

func (f *fakeNotificationService) ListNotifications(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error) {
	f.listCalls = append(f.listCalls, notificationListCall{
		userID:     userID,
		unreadOnly: unreadOnly,
		since:      since,
		limit:      limit,
	})
	if f.listNotificationsFn != nil {
		return f.listNotificationsFn(userID, unreadOnly, since, limit)
	}
	return nil, nil
}

func (f *fakeNotificationService) MarkNotificationRead(userID, notificationID int64) error {
	f.markReadCalls = append(f.markReadCalls, notificationReadCall{
		userID:         userID,
		notificationID: notificationID,
	})
	if f.markNotificationReadFn != nil {
		return f.markNotificationReadFn(userID, notificationID)
	}
	return nil
}

func (f *fakeNotificationService) MarkAllNotificationsRead(userID int64) error {
	f.markAllCalls = append(f.markAllCalls, userID)
	if f.markAllNotificationsReadFn != nil {
		return f.markAllNotificationsReadFn(userID)
	}
	return nil
}

func (f *fakeNotificationRefresher) RefreshNotificationsForUser(userID int64) error {
	f.refreshCalls = append(f.refreshCalls, userID)
	if f.refreshFn != nil {
		return f.refreshFn(userID)
	}
	return nil
}

func newNotificationRequest(method, target string, body []byte, token string) *http.Request {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	return req
}

func newNotificationHandlerForTest(t *testing.T, svc notificationService, refresher ...notificationRefresher) (*NotificationHandler, string) {
	t.Helper()

	manager := cloudauth.NewManager("test-secret", time.Hour)
	token, err := manager.Issue(42, "user@example.com")
	if err != nil {
		t.Fatalf("Issue token: %v", err)
	}

	return NewNotificationHandler(svc, manager, refresher...), token
}

func TestNotificationHandlerListNotifications(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		listNotificationsFn: func(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error) {
			if userID != 42 {
				t.Fatalf("userID = %d, want 42", userID)
			}
			if !unreadOnly {
				t.Fatal("unreadOnly = false, want true")
			}
			if since == nil {
				t.Fatal("since is nil")
			}
			if got, want := since.UTC().Format(time.RFC3339), "2026-04-11T10:00:00Z"; got != want {
				t.Fatalf("since = %q, want %q", got, want)
			}
			if limit != 25 {
				t.Fatalf("limit = %d, want 25", limit)
			}
			return []db.Notification{
				{ID: 7, UserID: 42, TaskID: "dev-1:ship", EventType: string(service.NotificationEventTaskCompleted), Title: "Done"},
			}, nil
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	req := newNotificationRequest(http.MethodGet, "/api/notifications?since=2026-04-11T10:00:00Z&unread=1&limit=25", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp map[string][]db.Notification
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(resp["notifications"]) != 1 {
		t.Fatalf("len(notifications) = %d, want 1", len(resp["notifications"]))
	}
}

func TestNotificationHandlerRefreshesTaskNotificationsBeforeListing(t *testing.T) {
	fakeSvc := &fakeNotificationService{}
	refresher := &fakeNotificationRefresher{
		refreshFn: func(userID int64) error {
			if userID != 42 {
				t.Fatalf("userID = %d, want 42", userID)
			}
			return nil
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc, refresher)

	req := newNotificationRequest(http.MethodGet, "/api/notifications", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(refresher.refreshCalls) != 1 {
		t.Fatalf("refreshCalls = %d, want 1", len(refresher.refreshCalls))
	}
}

func TestNotificationHandlerRejectsInvalidSince(t *testing.T) {
	handler, token := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodGet, "/api/notifications?since=not-a-time", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotificationHandlerMarkNotificationRead(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		markNotificationReadFn: func(userID, notificationID int64) error {
			if userID != 42 {
				t.Fatalf("userID = %d, want 42", userID)
			}
			if notificationID != 99 {
				t.Fatalf("notificationID = %d, want 99", notificationID)
			}
			return nil
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	body, _ := json.Marshal(map[string]int64{"notification_id": 99})
	req := newNotificationRequest(http.MethodPost, "/api/notifications/read", body, token)
	rr := httptest.NewRecorder()

	handler.MarkNotificationRead(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakeSvc.markReadCalls) != 1 {
		t.Fatalf("markReadCalls = %d, want 1", len(fakeSvc.markReadCalls))
	}
}

func TestNotificationHandlerMarkNotificationReadMapsAccessDeniedToForbidden(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		markNotificationReadFn: func(userID, notificationID int64) error {
			return service.ErrNotificationAccessDenied
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	body, _ := json.Marshal(map[string]int64{"notification_id": 99})
	req := newNotificationRequest(http.MethodPost, "/api/notifications/read", body, token)
	rr := httptest.NewRecorder()

	handler.MarkNotificationRead(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestNotificationHandlerMarkAllNotificationsRead(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		markAllNotificationsReadFn: func(userID int64) error {
			if userID != 42 {
				t.Fatalf("userID = %d, want 42", userID)
			}
			return nil
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	req := newNotificationRequest(http.MethodPost, "/api/notifications/read-all", nil, token)
	rr := httptest.NewRecorder()

	handler.MarkAllNotificationsRead(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if len(fakeSvc.markAllCalls) != 1 {
		t.Fatalf("markAllCalls = %d, want 1", len(fakeSvc.markAllCalls))
	}
}

func TestNotificationHandlerRejectsUnauthorizedRequests(t *testing.T) {
	handler, _ := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodGet, "/api/notifications", nil, "")
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestNotificationHandlerRejectsInvalidNotificationID(t *testing.T) {
	handler, token := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodPost, "/api/notifications/read", []byte(`{"notification_id":"bad"}`), token)
	rr := httptest.NewRecorder()

	handler.MarkNotificationRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotificationHandlerMapsNotFoundTo404(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		markNotificationReadFn: func(userID, notificationID int64) error {
			return service.ErrNotificationNotFound
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	body, _ := json.Marshal(map[string]int64{"notification_id": 99})
	req := newNotificationRequest(http.MethodPost, "/api/notifications/read", body, token)
	rr := httptest.NewRecorder()

	handler.MarkNotificationRead(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestNotificationHandlerInvalidLimit(t *testing.T) {
	handler, token := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodGet, "/api/notifications?limit=nope", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotificationHandlerReadAllUnauthorized(t *testing.T) {
	handler, _ := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodPost, "/api/notifications/read-all", nil, "")
	rr := httptest.NewRecorder()

	handler.MarkAllNotificationsRead(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestNotificationHandlerMarkNotificationReadRejectsMissingID(t *testing.T) {
	handler, token := newNotificationHandlerForTest(t, &fakeNotificationService{})

	req := newNotificationRequest(http.MethodPost, "/api/notifications/read", []byte(`{}`), token)
	rr := httptest.NewRecorder()

	handler.MarkNotificationRead(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestNotificationHandlerListNotificationsRejectsMalformedBodyIsIgnored(t *testing.T) {
	handler, token := newNotificationHandlerForTest(t, &fakeNotificationService{
		listNotificationsFn: func(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error) {
			return nil, nil
		},
	})

	req := newNotificationRequest(http.MethodGet, "/api/notifications?unread=true", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestNotificationHandlerMarkAllNotificationsMapsAccessDeniedToForbidden(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		markAllNotificationsReadFn: func(userID int64) error {
			return service.ErrNotificationAccessDenied
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	req := newNotificationRequest(http.MethodPost, "/api/notifications/read-all", nil, token)
	rr := httptest.NewRecorder()

	handler.MarkAllNotificationsRead(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestNotificationHandlerListNotificationsPropagatesServiceError(t *testing.T) {
	fakeSvc := &fakeNotificationService{
		listNotificationsFn: func(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error) {
			return nil, errors.New("boom")
		},
	}
	handler, token := newNotificationHandlerForTest(t, fakeSvc)

	req := newNotificationRequest(http.MethodGet, "/api/notifications", nil, token)
	rr := httptest.NewRecorder()

	handler.ListNotifications(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}
