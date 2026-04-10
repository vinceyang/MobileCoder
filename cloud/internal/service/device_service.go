package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/mobile-coder/cloud/internal/db"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrBindCodeExpired = errors.New("bind code expired")
)

type Device struct {
	ID           int64
	UserID       int64
	DeviceID     string
	DeviceName   string
	BindCode     string
	BindCodeExp  time.Time
	Status       string
	LastActiveAt string
}

type Session struct {
	ID          int64
	DeviceID    string
	SessionName string
	ProjectPath string
	Status      string
	CreatedAt   string
}

type DeviceService struct {
	db *db.SupabaseDB
}

func NewDeviceService(database *db.SupabaseDB) *DeviceService {
	return &DeviceService{db: database}
}

func generateCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// RegisterDevice creates a device with user-provided bind code (for Desktop Agent)
func (s *DeviceService) RegisterDevice(bindCode, deviceName string) (*Device, error) {
	deviceID := generateCode(16)
	bindCodeExp := time.Now().Add(10 * time.Minute).Format(time.RFC3339)

	// UserID is 0 until H5 binds the device
	device, err := s.db.CreateDevice(0, deviceID, deviceName, bindCode, bindCodeExp)
	if err != nil {
		return nil, err
	}

	loc := time.FixedZone("UTC+8", 8*3600)
	parsedTime, _ := time.ParseInLocation("2006-01-02T15:04:05", device.BindCodeExp, loc)
	return &Device{
		ID:          device.ID,
		UserID:      device.UserID,
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		BindCode:    device.BindCode,
		BindCodeExp: parsedTime,
		Status:      device.Status,
	}, nil
}

func (s *DeviceService) CreateBindCode(userID int64, deviceName string) (*Device, error) {
	deviceID := generateCode(16)
	bindCode := generateCode(6)
	bindCodeExp := time.Now().Add(10 * time.Minute).Format(time.RFC3339)

	device, err := s.db.CreateDevice(userID, deviceID, deviceName, bindCode, bindCodeExp)
	if err != nil {
		return nil, err
	}

	// Parse time - Supabase stores as "2026-02-25T18:03:16" without timezone
	loc := time.FixedZone("UTC+8", 8*3600)
	parsedTime, _ := time.ParseInLocation("2006-01-02T15:04:05", device.BindCodeExp, loc)
	return &Device{
		ID:          device.ID,
		UserID:      device.UserID,
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		BindCode:    device.BindCode,
		BindCodeExp: parsedTime,
		Status:      device.Status,
	}, nil
}

func (s *DeviceService) BindDevice(userID int64, bindCode string) (*Device, error) {
	device, err := s.db.GetDeviceByBindCode(bindCode)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	// Parse time - Supabase stores as "2026-02-25T18:03:16" without timezone
	// Use time.Parse without timezone (local time)
	loc := time.FixedZone("UTC+8", 8*3600)
	parsedTime, err := time.ParseInLocation("2006-01-02T15:04:05", device.BindCodeExp, loc)
	if err != nil {
		log.Printf("Failed to parse bind_code_exp: %v, value: %s", err, device.BindCodeExp)
		return nil, ErrBindCodeExpired
	}
	if time.Now().After(parsedTime) {
		return nil, ErrBindCodeExpired
	}

	// Update device status
	err = s.db.UpdateDeviceBindCode(device.DeviceID)
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:          device.ID,
		UserID:      device.UserID,
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		BindCode:    "",
		BindCodeExp: time.Time{},
		Status:      "online",
	}, nil
}

func (s *DeviceService) GetUserDevices(userID int64) ([]Device, error) {
	devices, err := s.db.GetUserDevices(userID)
	if err != nil {
		return nil, err
	}

	var result []Device
	for _, d := range devices {
		result = append(result, Device{
			ID:           d.ID,
			UserID:       d.UserID,
			DeviceID:     d.DeviceID,
			DeviceName:   d.DeviceName,
			Status:       d.Status,
			LastActiveAt: d.LastActiveAt,
		})
	}
	return result, nil
}

// CreateBindCodeSimple - 简化版，无需用户登录
func (s *DeviceService) CreateBindCodeSimple(deviceName string) (*Device, error) {
	deviceID := generateCode(16)
	bindCode := generateCode(6)
	// 简化版：绑定码永久有效
	bindCodeExp := ""

	device, err := s.db.CreateDevice(0, deviceID, deviceName, bindCode, bindCodeExp)
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:          device.ID,
		UserID:      device.UserID,
		DeviceID:    device.DeviceID,
		DeviceName:  device.DeviceName,
		BindCode:    device.BindCode,
		BindCodeExp: time.Time{},
		Status:      device.Status,
	}, nil
}

// BindDeviceSimple - 简化版，无需用户登录，绑定码永久有效
func (s *DeviceService) BindDeviceSimple(bindCode string) (*Device, error) {
	device, err := s.db.GetDeviceByBindCode(bindCode)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	// 简化版：不再检查绑定码过期时间

	// 更新设备状态
	err = s.db.UpdateDeviceBindCode(device.DeviceID)
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:         device.ID,
		UserID:     device.UserID,
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		Status:     "online",
	}, nil
}

// ListAllDevices - 列出所有设备（简化版）
func (s *DeviceService) ListAllDevices() ([]Device, error) {
	devices, err := s.db.ListAllDevices()
	if err != nil {
		return nil, err
	}

	var result []Device
	for _, d := range devices {
		// Parse time - Supabase stores as "2026-02-25T18:03:16" without timezone
		loc := time.FixedZone("UTC+8", 8*3600)
		parsedTime, _ := time.ParseInLocation("2006-01-02T15:04:05", d.BindCodeExp, loc)

		result = append(result, Device{
			ID:          d.ID,
			UserID:      d.UserID,
			DeviceID:    d.DeviceID,
			DeviceName:  d.DeviceName,
			BindCode:    d.BindCode,
			BindCodeExp: parsedTime,
			Status:      d.Status,
		})
	}
	return result, nil
}

// BindDeviceByCode binds a device using just the bind code (for Desktop Agent)
func (s *DeviceService) BindDeviceByCode(bindCode string) (*Device, error) {
	device, err := s.db.GetDeviceByBindCode(bindCode)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	// Parse time - Supabase stores as "2026-02-25T18:03:16" without timezone
	loc := time.FixedZone("UTC+8", 8*3600)
	parsedTime, err := time.ParseInLocation("2006-01-02T15:04:05", device.BindCodeExp, loc)
	if err != nil {
		log.Printf("Failed to parse bind_code_exp: %v, value: %s", err, device.BindCodeExp)
		return nil, ErrBindCodeExpired
	}
	if time.Now().After(parsedTime) {
		return nil, ErrBindCodeExpired
	}

	// Update device status to online
	err = s.db.UpdateDeviceBindCode(device.DeviceID)
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:         device.ID,
		UserID:     device.UserID,
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		Status:     "online",
	}, nil
}

// BindDeviceToUser 将设备绑定到用户
func (s *DeviceService) BindDeviceToUser(bindCode string, userID int64) (*Device, error) {
	// 通过绑定码找到设备
	device, err := s.db.GetDeviceByBindCode(bindCode)
	if err != nil {
		return nil, ErrDeviceNotFound
	}

	// 检查用户已绑定设备数量
	userDevices, err := s.db.GetUserDevices(userID)
	if err == nil && len(userDevices) >= 5 {
		return nil, errors.New("max devices reached")
	}

	// 绑定用户
	err = s.db.BindDeviceToUser(device.DeviceID, userID)
	if err != nil {
		return nil, err
	}

	// 清空绑定码
	err = s.db.UpdateDeviceBindCode(device.DeviceID)
	if err != nil {
		return nil, err
	}

	return &Device{
		ID:         device.ID,
		UserID:     userID,
		DeviceID:   device.DeviceID,
		DeviceName: device.DeviceName,
		Status:     "online",
	}, nil
}

// GetDeviceByDeviceID gets a device by device_id
func (s *DeviceService) GetDeviceByDeviceID(deviceID string) (*Device, error) {
	device, err := s.db.GetDeviceByDeviceID(deviceID)
	if err != nil {
		return nil, ErrDeviceNotFound
	}
	return &Device{
		ID:           device.ID,
		UserID:       device.UserID,
		DeviceID:     device.DeviceID,
		DeviceName:   device.DeviceName,
		Status:       device.Status,
		LastActiveAt: device.LastActiveAt,
	}, nil
}

// GetDeviceSessions 获取设备的所有 Session
func (s *DeviceService) GetDeviceSessions(deviceID string) ([]Session, error) {
	sessions, err := s.db.GetSessionsByDevice(deviceID)
	if err != nil {
		return nil, err
	}

	var result []Session
	for _, ses := range sessions {
		result = append(result, Session{
			ID:          ses.ID,
			DeviceID:    ses.DeviceID,
			SessionName: ses.SessionName,
			ProjectPath: ses.ProjectPath,
			Status:      ses.Status,
			CreatedAt:   ses.CreatedAt,
		})
	}
	return result, nil
}

// CreateSession 创建设备的 Session（如果已存在则更新）
func (s *DeviceService) CreateSession(deviceID, sessionName, projectPath string) (*Session, error) {
	session, err := s.db.CreateOrUpdateSession(deviceID, sessionName, projectPath)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:          session.ID,
		DeviceID:    session.DeviceID,
		SessionName: session.SessionName,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
	}, nil
}

// UpdateSessionStatus updates session status when agent disconnects
func (s *DeviceService) UpdateSessionStatus(deviceID, sessionName, status string) error {
	return s.db.UpdateSessionStatus(deviceID, sessionName, status)
}

// GetActiveSession gets the active session for a device
func (s *DeviceService) GetActiveSession(deviceID string) (*Session, error) {
	session, err := s.db.GetActiveSession(deviceID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, nil
	}
	return &Session{
		ID:          session.ID,
		DeviceID:    session.DeviceID,
		SessionName: session.SessionName,
		ProjectPath: session.ProjectPath,
		Status:      session.Status,
		CreatedAt:   session.CreatedAt,
	}, nil
}

// DeleteSession deletes a session
func (s *DeviceService) DeleteSession(sessionID int64) error {
	return s.db.DeleteSession(sessionID)
}

// UpdateDeviceName updates the device name
func (s *DeviceService) UpdateDeviceName(deviceID, deviceName string) error {
	return s.db.UpdateDeviceName(deviceID, deviceName)
}

// DeleteDevice deletes a device
func (s *DeviceService) DeleteDevice(deviceID string) error {
	return s.db.DeleteDevice(deviceID)
}
