package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/db"
	"github.com/mobile-coder/cloud/internal/service"
)

func TestRequireClaimsAcceptsSignedToken(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Minute)
	token, err := manager.Issue(99, "user@example.com")
	if err != nil {
		t.Fatalf("Issue token: %v", err)
	}

	claims, err := requireClaims(token, manager)
	if err != nil {
		t.Fatalf("requireClaims returned error: %v", err)
	}

	if claims.UserID != 99 {
		t.Fatalf("UserID = %d, want 99", claims.UserID)
	}
}

func TestRequireClaimsRejectsMissingToken(t *testing.T) {
	manager := cloudauth.NewManager("test-secret", time.Minute)

	if _, err := requireClaims("", manager); err == nil {
		t.Fatal("requireClaims succeeded with empty token")
	}
}

func TestEnsureDeviceOwnershipRejectsOtherUsersDevice(t *testing.T) {
	device := &service.Device{
		UserID:   7,
		DeviceID: "dev-1",
	}

	if err := ensureDeviceOwnership(device, 99); err == nil {
		t.Fatal("ensureDeviceOwnership succeeded for foreign device")
	}
}

func TestEnsureDeviceAccessAcceptsMatchingAgentToken(t *testing.T) {
	device := &service.Device{
		UserID:   7,
		DeviceID: "dev-1",
	}
	claims := &cloudauth.Claims{
		UserID:    7,
		DeviceID:  "dev-1",
		TokenType: "agent",
	}

	if err := ensureDeviceAccess(device, claims, "dev-1"); err != nil {
		t.Fatalf("ensureDeviceAccess returned error: %v", err)
	}
}

func TestEnsureDeviceAccessRejectsForeignAgentToken(t *testing.T) {
	device := &service.Device{
		UserID:   7,
		DeviceID: "dev-1",
	}
	claims := &cloudauth.Claims{
		UserID:    7,
		DeviceID:  "dev-2",
		TokenType: "agent",
	}

	if err := ensureDeviceAccess(device, claims, "dev-1"); err == nil {
		t.Fatal("ensureDeviceAccess succeeded for mismatched agent token")
	}
}

func TestCheckDeviceRefreshesExpiredMatchingAgentToken(t *testing.T) {
	supabase := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/rest/v1/devices" || r.URL.Query().Get("device_id") != "eq.dev-1" {
			t.Fatalf("unexpected supabase request: %s %s?%s", r.Method, r.URL.Path, r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{
			"id":        1,
			"user_id":   7,
			"device_id": "dev-1",
			"status":    "online",
		}})
	}))
	defer supabase.Close()

	database := db.NewSupabaseDB(&db.Config{ProjectURL: supabase.URL, APIKey: "test-key"})
	deviceService := service.NewDeviceService(database)
	manager := cloudauth.NewManager("test-secret", time.Hour)
	expiredManager := cloudauth.NewManager("test-secret", -time.Second)
	expiredToken, err := expiredManager.IssueAgent(7, "dev-1")
	if err != nil {
		t.Fatalf("IssueAgent expired token: %v", err)
	}

	handler := NewDeviceHandler(deviceService, manager)
	req := httptest.NewRequest(http.MethodPost, "/api/device/check", bytes.NewReader([]byte(`{"device_id":"dev-1"}`)))
	req.Header.Set("Authorization", expiredToken)
	rec := httptest.NewRecorder()

	handler.CheckDevice(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Valid      bool   `json:"valid"`
		Bound      bool   `json:"bound"`
		AgentToken string `json:"agent_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Valid || !payload.Bound {
		t.Fatalf("valid/bound = %v/%v, want true/true", payload.Valid, payload.Bound)
	}
	if payload.AgentToken == "" {
		t.Fatal("agent_token is empty")
	}
	claims, err := manager.Verify(payload.AgentToken)
	if err != nil {
		t.Fatalf("refreshed token does not verify: %v", err)
	}
	if claims.TokenType != "agent" || claims.UserID != 7 || claims.DeviceID != "dev-1" {
		t.Fatalf("claims = %+v, want matching agent token", claims)
	}
}
