package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	Host       string
	Port       string
	User       string
	Password   string
	DBName     string
	APIKey     string
	ProjectURL string
}

func LoadDBConfig() *Config {
	return &Config{
		Host:       os.Getenv("DB_HOST"),
		Port:       os.Getenv("DB_PORT"),
		User:       os.Getenv("DB_USER"),
		Password:   os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		APIKey:     os.Getenv("SUPABASE_API_KEY"),
		ProjectURL: os.Getenv("SUPABASE_PROJECT_URL"),
	}
}

// SupabaseDB uses REST API to interact with Supabase
type SupabaseDB struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewSupabaseDB(cfg *Config) *SupabaseDB {
	projectURL := cfg.ProjectURL
	if projectURL == "" {
		projectURL = "https://" + cfg.Host
	}
	return &SupabaseDB{
		baseURL: projectURL + "/rest/v1",
		apiKey:  cfg.APIKey,
		client:  &http.Client{},
	}
}

func (s *SupabaseDB) do(method, endpoint string, body []byte) ([]byte, error) {
	url := s.baseURL + endpoint
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", s.apiKey)
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Prefer", "return=representation")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s %s", resp.Status, string(respBody))
	}

	// Handle empty responses
	if len(respBody) == 0 {
		return []byte("[]"), nil
	}

	return respBody, nil
}

// User operations
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (s *SupabaseDB) CreateUser(username, password, email string) (*User, error) {
	body, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	})

	log.Printf("Creating user: %s", username)
	resp, err := s.do("POST", "/users", body)
	if err != nil {
		log.Printf("CreateUser error: %v", err)
		return nil, err
	}

	log.Printf("CreateUser response: %s", string(resp))
	// Supabase returns an array
	var users []User
	json.Unmarshal(resp, &users)
	if len(users) == 0 {
		return nil, fmt.Errorf("user not created")
	}
	return &users[0], nil
}

func (s *SupabaseDB) GetUserByUsername(username string) (*User, error) {
	log.Printf("GetUserByUsername: %s", username)
	resp, err := s.do("GET", "/users?username=eq."+username, nil)
	if err != nil {
		log.Printf("GetUserByUsername error: %v", err)
		return nil, err
	}
	log.Printf("GetUserByUsername response: %s", string(resp))

	var users []User
	json.Unmarshal(resp, &users)
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &users[0], nil
}

func (s *SupabaseDB) GetUserByEmail(email string) (*User, error) {
	log.Printf("GetUserByEmail: %s", email)
	resp, err := s.do("GET", "/users?email=eq."+email, nil)
	if err != nil {
		log.Printf("GetUserByEmail error: %v", err)
		return nil, err
	}
	log.Printf("GetUserByEmail response: %s", string(resp))

	var users []User
	json.Unmarshal(resp, &users)
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return &users[0], nil
}

// Device operations
type Device struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	DeviceID     string `json:"device_id"`
	DeviceName   string `json:"device_name"`
	BindCode     string `json:"bind_code"`
	BindCodeExp  string `json:"bind_code_exp"`
	Status       string `json:"status"`
	LastActiveAt string `json:"last_active_at"`
	CreatedAt    string `json:"created_at"`
}

// Session operations
type Session struct {
	ID          int64  `json:"id"`
	DeviceID    string `json:"device_id"`
	SessionName string `json:"session_name"`
	ProjectPath string `json:"project_path"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// Notification represents a persisted control tower notification record.
type Notification struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	TaskID      string `json:"task_id"`
	DeviceID    string `json:"device_id"`
	SessionName string `json:"session_name"`
	EventType   string `json:"event_type"`
	Title       string `json:"title"`
	Body        string `json:"body"`
	DedupeKey   string `json:"dedupe_key"`
	ReadAt      string `json:"read_at"`
	CreatedAt   string `json:"created_at"`
}

func (s *SupabaseDB) CreateSession(deviceID, sessionName, projectPath string) (*Session, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"device_id":    deviceID,
		"session_name": sessionName,
		"project_path": projectPath,
		"status":       "active",
	})

	resp, err := s.do("POST", "/sessions", body)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	json.Unmarshal(resp, &sessions)
	if len(sessions) == 0 {
		return nil, fmt.Errorf("session not created")
	}
	return &sessions[0], nil
}

// UpdateSessionStatus updates session status by device_id and session_name
func (s *SupabaseDB) UpdateSessionStatus(deviceID, sessionName, status string) error {
	body, _ := json.Marshal(map[string]string{
		"status": status,
	})
	encodedSessionName := url.QueryEscape(sessionName)
	_, err := s.do("PATCH", "/sessions?device_id=eq."+deviceID+"&session_name=eq."+encodedSessionName, body)
	return err
}

// DeleteSession deletes a session by ID
func (s *SupabaseDB) DeleteSession(sessionID int64) error {
	_, err := s.do("DELETE", "/sessions?id=eq."+fmt.Sprintf("%d", sessionID), nil)
	return err
}

// GetSessionByName finds an existing session by device_id and session_name
func (s *SupabaseDB) GetSessionByName(deviceID, sessionName string) (*Session, error) {
	encodedSessionName := url.QueryEscape(sessionName)
	resp, err := s.do("GET", "/sessions?device_id=eq."+deviceID+"&session_name=eq."+encodedSessionName, nil)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	json.Unmarshal(resp, &sessions)
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// GetActiveSession finds the active session for a device
func (s *SupabaseDB) GetActiveSession(deviceID string) (*Session, error) {
	resp, err := s.do("GET", "/sessions?device_id=eq."+deviceID+"&status=eq.active", nil)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	json.Unmarshal(resp, &sessions)
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
}

// CreateOrUpdateSession creates a new session or updates existing one
func (s *SupabaseDB) CreateOrUpdateSession(deviceID, sessionName, projectPath string) (*Session, error) {
	// First try to find existing session
	existing, err := s.GetSessionByName(deviceID, sessionName)
	if err != nil {
		return nil, err
	}

	if existing != nil {
		// Update existing session
		body, _ := json.Marshal(map[string]interface{}{
			"status":       "active",
			"project_path": projectPath,
		})
		_, err := s.do("PATCH", "/sessions?id=eq."+fmt.Sprintf("%d", existing.ID), body)
		if err != nil {
			return nil, err
		}
		existing.Status = "active"
		existing.ProjectPath = projectPath
		return existing, nil
	}

	// Create new session
	return s.CreateSession(deviceID, sessionName, projectPath)
}

func (s *SupabaseDB) GetSessionsByDevice(deviceID string) ([]Session, error) {
	resp, err := s.do("GET", "/sessions?device_id=eq."+deviceID, nil)
	if err != nil {
		return nil, err
	}

	var sessions []Session
	json.Unmarshal(resp, &sessions)
	return sessions, nil
}

func (s *SupabaseDB) CreateDevice(userID int64, deviceID, deviceName, bindCode string, bindCodeExp string) (*Device, error) {
	log.Printf("CreateDevice: userID=%d, deviceID=%s, deviceName=%s, bindCode=%s, bindCodeExp=%s",
		userID, deviceID, deviceName, bindCode, bindCodeExp)

	// If userID is 0, use null (device not yet bound to user)
	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}

	body, _ := json.Marshal(map[string]interface{}{
		"user_id":       userIDPtr,
		"device_id":     deviceID,
		"device_name":   deviceName,
		"bind_code":     bindCode,
		"bind_code_exp": bindCodeExp,
		"status":        "offline",
	})

	resp, err := s.do("POST", "/devices", body)
	if err != nil {
		log.Printf("CreateDevice error: %v", err)
		return nil, err
	}
	log.Printf("CreateDevice response: %s", string(resp))

	// Supabase returns an array
	var devices []Device
	json.Unmarshal(resp, &devices)
	if len(devices) == 0 {
		return nil, fmt.Errorf("device not created")
	}
	return &devices[0], nil
}

func (s *SupabaseDB) GetDeviceByBindCode(bindCode string) (*Device, error) {
	log.Printf("GetDeviceByBindCode: %s", bindCode)
	resp, err := s.do("GET", "/devices?bind_code=eq."+bindCode, nil)
	if err != nil {
		log.Printf("GetDeviceByBindCode error: %v", err)
		return nil, err
	}
	log.Printf("GetDeviceByBindCode response: %s", string(resp))

	var devices []Device
	json.Unmarshal(resp, &devices)
	if len(devices) == 0 {
		return nil, fmt.Errorf("device not found")
	}
	log.Printf("Found device: %+v", devices[0])
	return &devices[0], nil
}

func (s *SupabaseDB) GetDeviceByDeviceID(deviceID string) (*Device, error) {
	resp, err := s.do("GET", "/devices?device_id=eq."+deviceID, nil)
	if err != nil {
		return nil, err
	}

	var devices []Device
	json.Unmarshal(resp, &devices)
	if len(devices) == 0 {
		return nil, fmt.Errorf("device not found")
	}
	return &devices[0], nil
}

func (s *SupabaseDB) UpdateDeviceBindCode(deviceID string) error {
	// Clear bind code after successful binding
	body, _ := json.Marshal(map[string]interface{}{
		"bind_code":     nil,
		"bind_code_exp": nil,
		"status":        "online",
	})

	_, err := s.do("PATCH", "/devices?device_id=eq."+deviceID, body)
	return err
}

func (s *SupabaseDB) BindDeviceToUser(deviceID string, userID int64) error {
	body, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"status":  "online",
	})
	_, err := s.do("PATCH", "/devices?device_id=eq."+deviceID, body)
	return err
}

func (s *SupabaseDB) GetUserDevices(userID int64) ([]Device, error) {
	resp, err := s.do("GET", "/devices?user_id=eq."+fmt.Sprintf("%d", userID), nil)
	if err != nil {
		return nil, err
	}

	var devices []Device
	json.Unmarshal(resp, &devices)
	return devices, nil
}

// ListAllDevices returns all devices (simplified - no user filter)
func (s *SupabaseDB) ListAllDevices() ([]Device, error) {
	resp, err := s.do("GET", "/devices?select=*&order=created_at.desc", nil)
	if err != nil {
		return nil, err
	}

	var devices []Device
	json.Unmarshal(resp, &devices)
	return devices, nil
}

// Notification operations
func (s *SupabaseDB) CreateNotification(notification *Notification) (*Notification, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"user_id":      notification.UserID,
		"task_id":      notification.TaskID,
		"device_id":    notification.DeviceID,
		"session_name": notification.SessionName,
		"event_type":   notification.EventType,
		"title":        notification.Title,
		"body":         notification.Body,
		"dedupe_key":   notification.DedupeKey,
	})

	resp, err := s.do("POST", "/notifications", body)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	json.Unmarshal(resp, &notifications)
	if len(notifications) == 0 {
		return nil, fmt.Errorf("notification not created")
	}
	return &notifications[0], nil
}

func (s *SupabaseDB) ListNotificationsByUser(userID int64, limit int, since string, unreadOnly bool) ([]Notification, error) {
	query := []string{
		"user_id=eq." + fmt.Sprintf("%d", userID),
		"order=created_at.desc",
	}
	if limit > 0 {
		query = append(query, fmt.Sprintf("limit=%d", limit))
	}
	if since != "" {
		query = append(query, "created_at=gte."+url.QueryEscape(since))
	}
	if unreadOnly {
		query = append(query, "read_at=is.null")
	}

	resp, err := s.do("GET", "/notifications?select=*&"+strings.Join(query, "&"), nil)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	json.Unmarshal(resp, &notifications)
	return notifications, nil
}

func (s *SupabaseDB) GetNotificationByID(notificationID int64) (*Notification, error) {
	resp, err := s.do("GET", "/notifications?id=eq."+fmt.Sprintf("%d", notificationID), nil)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	json.Unmarshal(resp, &notifications)
	if len(notifications) == 0 {
		return nil, nil
	}
	return &notifications[0], nil
}

func (s *SupabaseDB) GetLatestNotificationByDedupeKey(userID int64, dedupeKey string) (*Notification, error) {
	resp, err := s.do("GET", "/notifications?user_id=eq."+fmt.Sprintf("%d", userID)+"&dedupe_key=eq."+url.QueryEscape(dedupeKey)+"&order=created_at.desc&limit=1", nil)
	if err != nil {
		return nil, err
	}

	var notifications []Notification
	json.Unmarshal(resp, &notifications)
	if len(notifications) == 0 {
		return nil, nil
	}
	return &notifications[0], nil
}

func (s *SupabaseDB) MarkNotificationRead(notificationID int64, readAt string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"read_at": readAt,
	})
	_, err := s.do("PATCH", "/notifications?id=eq."+fmt.Sprintf("%d", notificationID), body)
	return err
}

func (s *SupabaseDB) MarkAllNotificationsRead(userID int64, readAt string) error {
	body, _ := json.Marshal(map[string]interface{}{
		"read_at": readAt,
	})
	_, err := s.do("PATCH", "/notifications?user_id=eq."+fmt.Sprintf("%d", userID)+"&read_at=is.null", body)
	return err
}

func (s *SupabaseDB) DeleteNotificationsBefore(userID int64, cutoff string) error {
	_, err := s.do("DELETE", "/notifications?user_id=eq."+fmt.Sprintf("%d", userID)+"&created_at=lt."+url.QueryEscape(cutoff), nil)
	return err
}

// UpdateDeviceName updates the device name
func (s *SupabaseDB) UpdateDeviceName(deviceID, deviceName string) error {
	data := map[string]string{
		"device_name": deviceName,
	}
	body, _ := json.Marshal(data)
	_, err := s.do("PATCH", "/devices?device_id=eq."+deviceID, body)
	return err
}

// DeleteDevice deletes a device by device_id
func (s *SupabaseDB) DeleteDevice(deviceID string) error {
	_, err := s.do("DELETE", "/devices?device_id=eq."+deviceID, nil)
	return err
}

// InitDB returns a simple DB wrapper that uses Supabase REST API
func InitDB(cfg *Config) (*SupabaseDB, error) {
	db := NewSupabaseDB(cfg)

	// Test connection
	_, err := db.do("GET", "/users?limit=1", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Supabase: %w", err)
	}

	log.Println("Connected to Supabase via REST API")
	return db, nil
}
