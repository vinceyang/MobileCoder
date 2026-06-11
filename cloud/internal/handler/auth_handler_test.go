package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
)

func TestAuthHandlerRefreshCurrentUserTokenReturnsNewUserToken(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Hour)
	token, err := manager.Issue(123, "user@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	response := refreshToken(t, manager, token)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body AuthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.UserID != 123 {
		t.Fatalf("user_id = %d, want 123", body.UserID)
	}
	if body.Email != "user@example.com" {
		t.Fatalf("email = %q, want user@example.com", body.Email)
	}
	if body.Message != "refresh successful" {
		t.Fatalf("message = %q, want refresh successful", body.Message)
	}
	assertUserRefreshToken(t, manager, body.Token, 123, "user@example.com")
}

func TestAuthHandlerRefreshExpiredUserTokenWithinRenewalWindowReturnsNewUserToken(t *testing.T) {
	issuer := cloudauth.NewManager("test-secret", -time.Hour)
	token, err := issuer.Issue(456, "expired@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	manager := cloudauth.NewManager("test-secret", time.Hour)

	response := refreshToken(t, manager, token)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body AuthResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.UserID != 456 {
		t.Fatalf("user_id = %d, want 456", body.UserID)
	}
	if body.Email != "expired@example.com" {
		t.Fatalf("email = %q, want expired@example.com", body.Email)
	}
	if body.Message != "refresh successful" {
		t.Fatalf("message = %q, want refresh successful", body.Message)
	}
	assertUserRefreshToken(t, manager, body.Token, 456, "expired@example.com")
}

func TestAuthHandlerRefreshUserTokenExpiredBeyondRenewalWindowReturnsUnauthorized(t *testing.T) {
	issuer := cloudauth.NewManager("test-secret", -(31 * 24 * time.Hour))
	token, err := issuer.Issue(789, "stale@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	manager := cloudauth.NewManager("test-secret", time.Hour)

	response := refreshToken(t, manager, token)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestAuthHandlerRefreshAgentTokenReturnsUnauthorized(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Hour)
	token, err := manager.IssueAgent(123, "device-1")
	if err != nil {
		t.Fatalf("issue agent token: %v", err)
	}

	response := refreshToken(t, manager, token)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestAuthHandlerRefreshMalformedTokenReturnsUnauthorized(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Hour)

	response := refreshToken(t, manager, "not-a-token")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusUnauthorized, response.Body.String())
	}
}

func TestAuthHandlerRefreshGetReturnsMethodNotAllowed(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Hour)
	token, err := manager.Issue(123, "user@example.com")
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	handler := NewAuthHandler(nil, manager)
	request := httptest.NewRequest(http.MethodGet, "/api/auth/refresh", nil)
	request.Header.Set("Authorization", token)
	response := httptest.NewRecorder()

	handler.Refresh(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusMethodNotAllowed, response.Body.String())
	}
}

func refreshToken(t *testing.T, manager *cloudauth.Manager, token string) *httptest.ResponseRecorder {
	t.Helper()

	handler := NewAuthHandler(nil, manager)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	request.Header.Set("Authorization", token)
	response := httptest.NewRecorder()

	handler.Refresh(response, request)

	return response
}

func assertUserRefreshToken(t *testing.T, manager *cloudauth.Manager, token string, wantUserID int64, wantEmail string) {
	t.Helper()

	if token == "" {
		t.Fatal("token is empty")
	}
	claims, err := manager.VerifyAllowExpired(token)
	if err != nil {
		t.Fatalf("verify returned token: %v", err)
	}
	if claims.TokenType != "user" {
		t.Fatalf("token_type = %q, want user", claims.TokenType)
	}
	if claims.UserID != wantUserID {
		t.Fatalf("user_id = %d, want %d", claims.UserID, wantUserID)
	}
	if claims.Email != wantEmail {
		t.Fatalf("email = %q, want %s", claims.Email, wantEmail)
	}
}
