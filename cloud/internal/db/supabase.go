package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
