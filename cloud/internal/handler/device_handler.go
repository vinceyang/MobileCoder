package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/mobile-coder/cloud/internal/service"
)

// parseToken 解析 token 获取 userID
// 格式: "token_userID_email_timestamp"
func parseToken(token string) (int64, error) {
	// 格式: "token_userID_email_timestamp"
	parts := strings.Split(token, "_")
	if len(parts) < 3 {
		return 0, errors.New("invalid token format")
	}
	// parts[0] = "token", parts[1] = userID, parts[2] = email
	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.New("invalid token format")
	}
	return userID, nil
}

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
	// 检查是否有 Authorization header，如果有则绑定到用户
	token := r.Header.Get("Authorization")

	var req BindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 如果有 token，绑定到对应用户
	if token != "" {
		userID, err := parseToken(token)
		if err != nil {
			// 如果 token 无效，回退到简化版
			device, err := h.deviceService.BindDeviceSimple(req.BindCode)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"device_id":   device.DeviceID,
				"device_name": device.DeviceName,
				"status":      device.Status,
			})
			return
		}

		// 绑定到指定用户
		device, err := h.deviceService.BindDeviceToUser(req.BindCode, userID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"device_id":   device.DeviceID,
			"device_name": device.DeviceName,
			"status":      device.Status,
		})
		return
	}

	// 无 token，回退到简化版
	device, err := h.deviceService.BindDeviceSimple(req.BindCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
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

// GetUserDevices 获取用户的所有设备
func (h *DeviceHandler) GetUserDevices(w http.ResponseWriter, r *http.Request) {
	// 从 header 获取 token，解析 userID
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := parseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	devices, err := h.deviceService.GetUserDevices(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"devices": devices,
	})
}

// GetDeviceSessions 获取设备的所有 Session
func (h *DeviceHandler) GetDeviceSessions(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	sessions, err := h.deviceService.GetDeviceSessions(deviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"sessions": sessions,
	})
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	DeviceID    string `json:"device_id"`
	SessionName string `json:"session_name"`
	ProjectPath string `json:"project_path"`
}

// CreateSession 创建新 Session
func (h *DeviceHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" || req.SessionName == "" {
		http.Error(w, "device_id and session_name required", http.StatusBadRequest)
		return
	}

	session, err := h.deviceService.CreateSession(req.DeviceID, req.SessionName, req.ProjectPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id": session.ID,
		"session_name": session.SessionName,
		"status": session.Status,
	})
}

type DeleteSessionRequest struct {
	SessionID int64 `json:"session_id"`
}

// DeleteSession deletes a session
func (h *DeviceHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == 0 {
		http.Error(w, "session_id required", http.StatusBadRequest)
		return
	}

	err := h.deviceService.DeleteSession(req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

type UpdateDeviceRequest struct {
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

// UpdateDevice updates device info (like name)
func (h *DeviceHandler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "PUT" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	err := h.deviceService.UpdateDeviceName(req.DeviceID, req.DeviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

type DeleteDeviceRequest struct {
	DeviceID string `json:"device_id"`
}

// DeleteDevice deletes a device
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" && r.Method != "DELETE" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.DeviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	err := h.deviceService.DeleteDevice(req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
