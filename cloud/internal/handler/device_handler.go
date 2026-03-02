package handler

import (
	"encoding/json"
	"net/http"

	"github.com/mobile-coder/cloud/internal/service"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
}

func NewDeviceHandler(deviceService *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

type CreateBindCodeRequest struct {
	DeviceName string `json:"device_name"`
}

type BindRequest struct {
	BindCode string `json:"bind_code"`
}

type DeviceRegisterRequest struct {
	BindCode   string `json:"bind_code"`
	DeviceName string `json:"device_name"`
}

type DeviceCheckRequest struct {
	DeviceID string `json:"device_id"`
}

// Register allows Desktop Agent to register itself (no auth required)
func (h *DeviceHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req DeviceRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.BindCode == "" {
		http.Error(w, "bind_code required", http.StatusBadRequest)
		return
	}

	// Create device with the provided bind code
	device, err := h.deviceService.RegisterDevice(req.BindCode, req.DeviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_id":   device.DeviceID,
		"bind_code":   device.BindCode,
		"expires_at":  device.BindCodeExp,
	})
}

func (h *DeviceHandler) CreateBindCode(w http.ResponseWriter, r *http.Request) {
	// 简化版：无需用户登录，直接创建设备
	// 此接口保留但不使用，由 Register 接口替代

	var req CreateBindCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	device, err := h.deviceService.CreateBindCodeSimple(req.DeviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_id":  device.DeviceID,
		"bind_code":  device.BindCode,
		"expires_at": device.BindCodeExp,
	})
}

func (h *DeviceHandler) BindDevice(w http.ResponseWriter, r *http.Request) {
	// 简化版：无需用户登录，直接通过绑定码绑定设备
	var req BindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	device, err := h.deviceService.BindDeviceSimple(req.BindCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_id":   device.DeviceID,
		"device_name": device.DeviceName,
		"status":      device.Status,
	})
}

func (h *DeviceHandler) ListDevices(w http.ResponseWriter, r *http.Request) {
	// 简化版：无需用户登录，返回所有设备（或按需扩展）
	devices, err := h.deviceService.ListAllDevices()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

// BindAgent allows Desktop Agent to bind using a bind code (no auth required)
func (h *DeviceHandler) BindAgent(w http.ResponseWriter, r *http.Request) {
	var req BindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// For Desktop Agent, we need to find the device by bind code first
	// This is a simplified flow - the Desktop Agent uses bind code to get device_id
	device, err := h.deviceService.BindDeviceByCode(req.BindCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_id":   device.DeviceID,
		"device_name": device.DeviceName,
		"status":      device.Status,
	})
}

// CheckDevice checks if a device_id is valid
func (h *DeviceHandler) CheckDevice(w http.ResponseWriter, r *http.Request) {
	var req DeviceCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	device, err := h.deviceService.GetDeviceByDeviceID(req.DeviceID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"valid": false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":   true,
		"status":  device.Status,
	})
}
