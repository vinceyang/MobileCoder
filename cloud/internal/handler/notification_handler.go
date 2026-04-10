package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/service"
)

type notificationService interface {
	ListNotifications(userID int64, unreadOnly bool, since *time.Time, limit int) ([]db.Notification, error)
	MarkNotificationRead(userID, notificationID int64) error
	MarkAllNotificationsRead(userID int64) error
}

type NotificationHandler struct {
	service      notificationService
	refresher    notificationRefresher
	tokenManager *cloudauth.Manager
}

type notificationRefresher interface {
	RefreshNotificationsForUser(userID int64) error
}

func NewNotificationHandler(service notificationService, tokenManager *cloudauth.Manager, refresher ...notificationRefresher) *NotificationHandler {
	handler := &NotificationHandler{
		service:      service,
		tokenManager: tokenManager,
	}
	if len(refresher) > 0 {
		handler.refresher = refresher[0]
	}
	return handler
}

func (h *NotificationHandler) refreshNotifications(userID int64) error {
	if h.refresher == nil {
		return nil
	}
	return h.refresher.RefreshNotificationsForUser(userID)
}

func (h *NotificationHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	unreadOnly, err := parseNotificationBoolQuery(r.URL.Query().Get("unread"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit, err := parseNotificationLimitQuery(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.refreshNotifications(claims.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	since, err := parseNotificationSinceQuery(r.URL.Query().Get("since"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	notifications, err := h.service.ListNotifications(claims.UserID, unreadOnly, since, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"notifications": notifications,
	})
}

func (h *NotificationHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req notificationReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.NotificationID <= 0 {
		http.Error(w, "notification_id required", http.StatusBadRequest)
		return
	}

	if err := h.service.MarkNotificationRead(claims.UserID, req.NotificationID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationAccessDenied):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, service.ErrNotificationNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}

func (h *NotificationHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.service.MarkAllNotificationsRead(claims.UserID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotificationAccessDenied):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status": "ok",
	})
}

type notificationReadRequest struct {
	NotificationID int64 `json:"notification_id"`
}

func parseNotificationSinceQuery(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, errors.New("since must be RFC3339")
	}
	return &parsed, nil
}

func parseNotificationLimitQuery(value string) (int, error) {
	if value == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New("limit must be a number")
	}
	if limit < 0 {
		return 0, errors.New("limit must be non-negative")
	}
	return limit, nil
}

func parseNotificationBoolQuery(value string) (bool, error) {
	if value == "" {
		return false, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, errors.New("unread must be a boolean")
	}
	return parsed, nil
}
