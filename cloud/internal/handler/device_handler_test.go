package handler

import (
	"testing"
	"time"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
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
