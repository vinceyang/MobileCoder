package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"github.com/coder/agentapi/cloud/internal/db"
)

var (
	ErrDeviceNotFound  = errors.New("device not found")
	ErrBindCodeExpired = errors.New("bind code expired")
)

type Device struct {
	ID          int64
	UserID      int64
	DeviceID    string
	DeviceName  string
	BindCode    string
	BindCodeExp time.Time
	Status      string
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
			ID:         d.ID,
			UserID:     d.UserID,
			DeviceID:   d.DeviceID,
			DeviceName: d.DeviceName,
			Status:     d.Status,
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
		result = append(result, Device{
			ID:         d.ID,
			UserID:     d.UserID,
			DeviceID:   d.DeviceID,
			DeviceName: d.DeviceName,
			Status:     d.Status,
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
