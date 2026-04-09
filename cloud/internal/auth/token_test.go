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
