package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mobile-coder/cloud/internal/ws"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHubHandler struct {
	hub *ws.Hub
}

func NewWSHubHandler(hub *ws.Hub) *WSHubHandler {
	return &WSHubHandler{hub: hub}
}

func (h *WSHubHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device_id")
	token := r.URL.Query().Get("token")

	log.Printf("WS connection request: device_id=%s, token=%s", deviceID, token)

	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	// token 参数用于标识客户端类型：
	// - 有 token: H5 viewer（会收到终端输出）
	// - 无 token: Desktop Agent（发送终端输出）
	var userID int64
	if token != "" {
		// 简单的 token 解析（不验证，仅用于标识）
		// 后续可以通过其他方式实现用户识别
		log.Printf("WS: viewer connected with token")
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	log.Printf("WS: connected device_id=%s, userID=%d", deviceID, userID)

	client := &ws.Client{
		Conn:     conn,
		DeviceID: deviceID,
		UserID:   userID,
		Send:     make(chan []byte, 256),
	}

	h.hub.Register(client)

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
		h.hub.Unregister(client.DeviceID)
		client.Conn.Close()
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
			// Broadcast terminal_output only to H5 viewers (not to agents)
			h.hub.BroadcastToViewers(client.DeviceID, message)
		} else if msgType == "terminal_input" {
			// terminal_input from H5 should only go to Desktop Agents
			h.hub.SendToAgents(client.DeviceID, message)
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
