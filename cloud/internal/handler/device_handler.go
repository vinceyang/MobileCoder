package handler

import (
	"encoding/json"
	"log"
	"net/http"

	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/service"
)

type DeviceHandler struct {
	deviceService *service.DeviceService
	tokenManager  *cloudauth.Manager
}

func NewDeviceHandler(deviceService *service.DeviceService, tokenManager *cloudauth.Manager) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		tokenManager:  tokenManager,
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
	BindCode string `json:"bind_code,omitempty"`
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
		"device_id":  device.DeviceID,
		"bind_code":  device.BindCode,
		"expires_at": device.BindCodeExp,
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
	var req BindRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	device, err := h.deviceService.BindDeviceToUser(req.BindCode, claims.UserID)
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
	response := map[string]interface{}{
		"valid":  true,
		"bound":  device.UserID > 0,
		"status": device.Status,
	}
	if device.UserID > 0 && req.BindCode != "" && req.BindCode == device.BindCode {
		agentToken, err := h.tokenManager.IssueAgent(device.UserID, device.DeviceID)
		if err != nil {
			http.Error(w, "failed to issue agent token", http.StatusInternalServerError)
			return
		}
		if err := h.deviceService.UpdateDeviceBindCode(device.DeviceID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		response["agent_token"] = agentToken
	} else if device.UserID > 0 && h.tokenManager != nil && r.Header.Get("Authorization") != "" {
		claims, err := h.tokenManager.VerifyAllowExpired(r.Header.Get("Authorization"))
		if err == nil && claims.TokenType == "agent" && claims.DeviceID == device.DeviceID && claims.UserID == device.UserID {
			agentToken, err := h.tokenManager.IssueAgent(device.UserID, device.DeviceID)
			if err != nil {
				http.Error(w, "failed to issue agent token", http.StatusInternalServerError)
				return
			}
			response["agent_token"] = agentToken
		}
	}
	json.NewEncoder(w).Encode(response)
}

// GetUserDevices 获取用户的所有设备
func (h *DeviceHandler) GetUserDevices(w http.ResponseWriter, r *http.Request) {
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	devices, err := h.deviceService.GetUserDevices(claims.UserID)
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
	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	device, err := h.deviceService.GetDeviceByDeviceID(deviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := ensureDeviceAccess(device, claims, deviceID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
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

	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	device, err := h.deviceService.GetDeviceByDeviceID(req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := ensureDeviceOwnership(device, claims.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	log.Printf("CreateSession: deviceID=%s, sessionName=%s, projectPath=%s", req.DeviceID, req.SessionName, req.ProjectPath)
	session, err := h.deviceService.CreateSession(req.DeviceID, req.SessionName, req.ProjectPath)
	if err != nil {
		log.Printf("CreateSession error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("CreateSession success: id=%d, sessionName=%s, status=%s", session.ID, session.SessionName, session.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session_id":   session.ID,
		"session_name": session.SessionName,
		"status":       session.Status,
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

	if _, err := requireClaimsFromRequest(r, h.tokenManager); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	device, err := h.deviceService.GetDeviceByDeviceID(req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := ensureDeviceAccess(device, claims, req.DeviceID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = h.deviceService.UpdateDeviceName(req.DeviceID, req.DeviceName)
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

	claims, err := requireClaimsFromRequest(r, h.tokenManager)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	device, err := h.deviceService.GetDeviceByDeviceID(req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := ensureDeviceOwnership(device, claims.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = h.deviceService.DeleteDevice(req.DeviceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
