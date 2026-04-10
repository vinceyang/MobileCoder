package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/service"
)

func TestGetTasksReturnsTaskPayloadForAuthorizedUser(t *testing.T) {
	fakeDB := newFakeSupabaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/v1/devices" && strings.Contains(r.URL.RawQuery, "user_id=eq.7"):
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"id":          1,
				"user_id":     7,
				"device_id":   "dev-1",
				"device_name": "MacBook",
				"status":      "online",
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/rest/v1/sessions" && strings.Contains(r.URL.RawQuery, "device_id=eq.dev-1"):
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"id":           10,
				"device_id":    "dev-1",
				"session_name": "feature-branch",
				"project_path": "/Users/me/repo-a",
				"status":       "active",
			}})
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	})
	defer fakeDB.Close()

	manager := cloudauth.NewManager("test-secret", time.Minute)
	token, err := manager.Issue(7, "user@example.com")
	if err != nil {
		t.Fatalf("Issue token: %v", err)
	}

	handler := NewTaskHandler(
		service.NewTaskService(service.NewDeviceService(db.NewSupabaseDB(&db.Config{
			APIKey:     "test-key",
			ProjectURL: fakeDB.URL,
		}))),
		manager,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()

	handler.GetTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload struct {
		Tasks []map[string]any `json:"tasks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	if len(payload.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(payload.Tasks))
	}
	if payload.Tasks[0]["state"] != string(service.TaskStateRunning) {
		t.Fatalf("state = %v, want %q", payload.Tasks[0]["state"], service.TaskStateRunning)
	}
}

func TestGetTaskReturnsTaskPayloadForAuthorizedUser(t *testing.T) {
	fakeDB := newFakeSupabaseServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/rest/v1/devices" && strings.Contains(r.URL.RawQuery, "user_id=eq.7"):
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"id":          1,
				"user_id":     7,
				"device_id":   "dev-1",
				"device_name": "MacBook",
				"status":      "online",
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/rest/v1/sessions" && strings.Contains(r.URL.RawQuery, "device_id=eq.dev-1"):
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"id":           10,
				"device_id":    "dev-1",
				"session_name": "feature-branch",
				"project_path": "/Users/me/repo-a",
				"status":       "active",
			}})
		default:
			t.Fatalf("unexpected request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
	})
	defer fakeDB.Close()

	manager := cloudauth.NewManager("test-secret", time.Minute)
	token, err := manager.Issue(7, "user@example.com")
	if err != nil {
		t.Fatalf("Issue token: %v", err)
	}

	handler := NewTaskHandler(
		service.NewTaskService(service.NewDeviceService(db.NewSupabaseDB(&db.Config{
			APIKey:     "test-key",
			ProjectURL: fakeDB.URL,
		}))),
		manager,
	)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/detail?id=dev-1:feature-branch", nil)
	req.Header.Set("Authorization", token)
	rec := httptest.NewRecorder()

	handler.GetTask(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var payload struct {
		Task map[string]any `json:"task"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}
	if payload.Task["id"] != "dev-1:feature-branch" {
		t.Fatalf("task.id = %v, want %q", payload.Task["id"], "dev-1:feature-branch")
	}
}

func newFakeSupabaseServer(t *testing.T, handler func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
}
