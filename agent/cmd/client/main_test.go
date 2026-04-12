package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestLoadOrCreateDeviceIDClearsStaleBindCodeForDeviceWithAgentToken(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	deviceIDPath := getDeviceIDPath()
	bindCodePath := getBindCodePath()
	agentTokenPath := getAgentTokenPath()
	if err := os.MkdirAll(filepath.Dir(deviceIDPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(deviceIDPath, []byte("dev-123"), 0o644); err != nil {
		t.Fatalf("write device-id: %v", err)
	}
	if err := os.WriteFile(bindCodePath, []byte("abc123"), 0o644); err != nil {
		t.Fatalf("write bind-code: %v", err)
	}
	if err := os.WriteFile(agentTokenPath, []byte("agent-token"), 0o644); err != nil {
		t.Fatalf("write agent-token: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/device/check" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid": true,
			"bound": true,
		})
	}))
	defer server.Close()

	deviceID, bindCode, err := loadOrCreateDeviceID(server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("loadOrCreateDeviceID: %v", err)
	}

	if deviceID != "dev-123" {
		t.Fatalf("deviceID = %q, want dev-123", deviceID)
	}
	if bindCode != "" {
		t.Fatalf("bindCode = %q, want empty for bound device", bindCode)
	}
	if _, err := os.Stat(bindCodePath); !os.IsNotExist(err) {
		t.Fatalf("bind-code file still exists, stat err = %v", err)
	}
}

func TestWaitForDeviceBindingClearsBindCodeAfterServerConfirms(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	bindCodePath := getBindCodePath()
	agentTokenPath := getAgentTokenPath()
	if err := os.MkdirAll(filepath.Dir(bindCodePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(bindCodePath, []byte("abc123"), 0o644); err != nil {
		t.Fatalf("write bind-code: %v", err)
	}

	var checks int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/device/check" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		bound := atomic.AddInt32(&checks, 1) >= 3
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid":       true,
			"bound":       bound,
			"agent_token": "agent-token",
		})
	}))
	defer server.Close()

	if err := waitForDeviceBinding(server.Listener.Addr().String(), "dev-123", "abc123", 250*time.Millisecond, 10*time.Millisecond); err != nil {
		t.Fatalf("waitForDeviceBinding: %v", err)
	}

	if atomic.LoadInt32(&checks) < 3 {
		t.Fatalf("checks = %d, want at least 3", checks)
	}
	if _, err := os.Stat(bindCodePath); !os.IsNotExist(err) {
		t.Fatalf("bind-code file still exists, stat err = %v", err)
	}
	data, err := os.ReadFile(agentTokenPath)
	if err != nil {
		t.Fatalf("read agent-token: %v", err)
	}
	if string(data) != "agent-token" {
		t.Fatalf("agent-token = %q, want agent-token", string(data))
	}
}

func TestLoadOrCreateDeviceIDRequiresRebindWithoutTokenOrBindCode(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	deviceIDPath := getDeviceIDPath()
	if err := os.MkdirAll(filepath.Dir(deviceIDPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(deviceIDPath, []byte("dev-123"), 0o644); err != nil {
		t.Fatalf("write device-id: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/device/check" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"valid": true,
			"bound": true,
		})
	}))
	defer server.Close()

	if _, _, err := loadOrCreateDeviceID(server.Listener.Addr().String()); err == nil {
		t.Fatal("loadOrCreateDeviceID succeeded without agent token or bind code")
	}
}

func TestTerminalInputToTmuxCommandsSubmitsTextWithEnterKey(t *testing.T) {
	got := terminalInputToTmuxCommands("codex-session", map[string]interface{}{
		"content": "fix the login flow\n",
	})

	want := [][]string{
		{"send-keys", "-t", "codex-session", "-l", "fix the login flow"},
		{"send-keys", "-t", "codex-session", "C-m"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestTerminalInputToTmuxCommandsUsesNativeEnterForKeyAction(t *testing.T) {
	got := terminalInputToTmuxCommands("codex-session", map[string]interface{}{
		"action":    "key",
		"key":       "Enter",
		"modifiers": []interface{}{},
	})

	want := [][]string{
		{"send-keys", "-t", "codex-session", "C-m"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestTerminalInputToTmuxCommandsUsesNativeControlKey(t *testing.T) {
	got := terminalInputToTmuxCommands("codex-session", map[string]interface{}{
		"action":    "key",
		"key":       "c",
		"modifiers": []interface{}{"ctrl"},
	})

	want := [][]string{
		{"send-keys", "-t", "codex-session", "C-c"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestIsLiteralTmuxInput(t *testing.T) {
	if !isLiteralTmuxInput([]string{"send-keys", "-t", "codex-session", "-l", "hello"}) {
		t.Fatal("literal command was not detected")
	}
	if isLiteralTmuxInput([]string{"send-keys", "-t", "codex-session", "C-m"}) {
		t.Fatal("native key command was detected as literal input")
	}
}
