package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	cloudauth "github.com/mobile-coder/cloud/internal/auth"
	"github.com/mobile-coder/cloud/internal/service"
	"github.com/mobile-coder/cloud/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHubHandler struct {
	hub           *ws.Hub
	deviceService *service.DeviceService
	tokenManager  *cloudauth.Manager
}

func NewWSHubHandler(hub *ws.Hub, deviceService *service.DeviceService, tokenManager *cloudauth.Manager) *WSHubHandler {
	return &WSHubHandler{
		hub:           hub,
		deviceService: deviceService,
		tokenManager:  tokenManager,
	}
}

func (h *WSHubHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	token := r.URL.Query().Get("token")
	sessionName := r.URL.Query().Get("session_name")

	log.Printf("WS connection request: device_id=%s, token=%s, session_name=%s", deviceID, token, sessionName)

	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	// token 参数用于标识客户端类型：
	// - 有 token: H5 viewer（会收到终端输出）
	// - 无 token: Desktop Agent（发送终端输出）
	var userID int64
	if token != "" {
		claims, err := requireClaims(token, h.tokenManager)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		device, err := h.deviceService.GetDeviceByDeviceID(deviceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err := ensureDeviceOwnership(device, claims.UserID); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		userID = claims.UserID
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	log.Printf("WS: connected device_id=%s, userID=%d, session_name=%s", deviceID, userID, sessionName)

	// If no token, it's a Desktop Agent; if token exists, it's an H5 viewer
	isAgent := (token == "")

	client := &ws.Client{
		Conn:        conn,
		DeviceID:    deviceID,
		UserID:      userID,
		IsAgent:     isAgent,
		SessionName: sessionName,
		Send:        make(chan []byte, 256),
	}

	h.hub.Register(client)

	// Desktop Agent (no token) - immediately update session status to active
	// This handles reconnection scenarios where terminal_output might not be sent immediately
	if token == "" && sessionName != "" {
		go func() {
			h.deviceService.UpdateSessionStatus(deviceID, sessionName, "active")
			log.Printf("WS: Desktop Agent connected, updated session status to active")
		}()
	}

	// If it's not an agent (i.e., it's an H5 viewer), send the last terminal output
	if token != "" {
		go func() {
			h.hub.SendLastOutput(client)
		}()
	}

	go h.writePump(client)
	go h.readPump(client)
}

func (h *WSHubHandler) readPump(client *ws.Client) {
	defer func() {
		h.hub.Unregister(client)
		client.Conn.Close()

		// If agent disconnects, update session status to inactive
		if client.IsAgent {
			// Get the active session and update its status
			session, err := h.deviceService.GetActiveSession(client.DeviceID)
			if err == nil && session != nil {
				log.Printf("Agent disconnected, updating session status to inactive for deviceID=%s, session=%s", client.DeviceID, session.SessionName)
				h.deviceService.UpdateSessionStatus(client.DeviceID, session.SessionName, "inactive")
			}
		}
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse message to determine type
		var msg map[string]interface{}
		json.Unmarshal(message, &msg)
		msgType, _ := msg["type"].(string)
		log.Printf("readPump: received msgType=%s from userID=%d", msgType, client.UserID)

		// If client sends terminal_output, it's a Desktop Agent
		if msgType == "terminal_output" {
			client.IsAgent = true
			// Update session status to active when agent connects
			if client.SessionName != "" {
				h.deviceService.UpdateSessionStatus(client.DeviceID, client.SessionName, "active")
			}
			// Broadcast terminal_output only to H5 viewers (not to agents)
			h.hub.BroadcastToViewers(client.DeviceID, client.SessionName, message)
		} else if msgType == "terminal_input" {
			// terminal_input from H5 should only go to Desktop Agents
			// Use sessionName for routing if available
			h.hub.SendToAgents(client.DeviceID, client.SessionName, message)
		} else {
			// Forward other messages to all clients
			h.hub.BroadcastToDevice(client.DeviceID, message)
		}
	}
}

func (h *WSHubHandler) writePump(client *ws.Client) {
	defer client.Conn.Close()

	for {
		message, ok := <-client.Send
		if !ok {
			client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		client.Conn.WriteMessage(websocket.TextMessage, message)
	}
}
