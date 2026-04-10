package handler

import (
	"encoding/json"
	"net/http"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/service"
)

type TaskHandler struct {
	taskService  *service.TaskService
	tokenManager *cloudauth.Manager
}

func NewTaskHandler(taskService *service.TaskService, tokenManager *cloudauth.Manager) *TaskHandler {
	return &TaskHandler{
		taskService:  taskService,
		tokenManager: tokenManager,
	}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	tasks, err := h.taskService.ListTasksForUser(claims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"tasks": tasks,
	})
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "task id is required", http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTaskForUser(claims.UserID, taskID)
	if err != nil {
		if err == service.ErrTaskNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"task": task,
	})
}
