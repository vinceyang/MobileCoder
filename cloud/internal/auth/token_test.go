package auth

import (
	"testing"
	"time"
)

func TestManagerRoundTrip(t *testing.T) {
	manager := NewManager("test-secret", time.Minute)

	token, err := manager.Issue(42, "user@example.com")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", claims.Email)
	}
}

func TestManagerRejectsTamperedToken(t *testing.T) {
	manager := NewManager("test-secret", time.Minute)

	token, err := manager.Issue(42, "user@example.com")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := manager.Verify(tampered); err == nil {
		t.Fatal("Verify succeeded for tampered token")
	}
}

func TestManagerRejectsExpiredToken(t *testing.T) {
	manager := NewManager("test-secret", -time.Second)

	token, err := manager.Issue(42, "user@example.com")
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}

	if _, err := manager.Verify(token); err == nil {
		t.Fatal("Verify succeeded for expired token")
	}
}

func TestManagerVerifiesExpiredTokenForRefresh(t *testing.T) {
	manager := NewManager("test-secret", -time.Second)

	token, err := manager.IssueAgent(42, "device-123")
	if err != nil {
		t.Fatalf("IssueAgent returned error: %v", err)
	}

	claims, err := manager.VerifyAllowExpired(token)
	if err != nil {
		t.Fatalf("VerifyAllowExpired returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", claims.UserID)
	}
	if claims.DeviceID != "device-123" {
		t.Fatalf("DeviceID = %q, want device-123", claims.DeviceID)
	}
	if claims.TokenType != "agent" {
		t.Fatalf("TokenType = %q, want agent", claims.TokenType)
	}
}

func TestManagerIssuesAndVerifiesAgentToken(t *testing.T) {
	manager := NewManager("test-secret", time.Minute)

	token, err := manager.IssueAgent(42, "device-123")
	if err != nil {
		t.Fatalf("IssueAgent returned error: %v", err)
	}

	claims, err := manager.Verify(token)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}

	if claims.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", claims.UserID)
	}
	if claims.DeviceID != "device-123" {
		t.Fatalf("DeviceID = %q, want device-123", claims.DeviceID)
	}
	if claims.TokenType != "agent" {
		t.Fatalf("TokenType = %q, want agent", claims.TokenType)
	}
}
