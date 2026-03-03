package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mobile-coder/cloud/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	UserID  int64  `json:"user_id"`
	Email   string `json:"email"`
	Token   string `json:"token"` // 简化的 token
	Message string `json:"message,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Register(req.Email, req.Password)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 生成简单的 token (实际应该用 JWT)
	token := generateToken(user.ID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		UserID:  user.ID,
		Email:   user.Email,
		Token:   token,
		Message: "registration successful",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	token := generateToken(user.ID, user.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		UserID:  user.ID,
		Email:   user.Email,
		Token:   token,
		Message: "login successful",
	})
}

// generateToken 生成简单的 token (实际应使用 JWT)
func generateToken(userID int64, email string) string {
	return fmt.Sprintf("token_%d_%s_%d", userID, email, time.Now().Unix())
}
