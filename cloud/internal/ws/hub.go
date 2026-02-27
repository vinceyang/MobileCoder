package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	Conn     *websocket.Conn
	DeviceID string
	UserID   int64
	IsAgent  bool  // true for Desktop Agent, false for H5 viewer
	Send     chan []byte
}

type Hub struct {
	clients          map[string]map[*Client]bool  // deviceID -> set of clients
	lastOutput       map[string][]byte             // deviceID -> last terminal output
	mu               sync.RWMutex
	register         chan *Client
	unregister       chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		lastOutput: make(map[string][]byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.DeviceID] == nil {
				h.clients[client.DeviceID] = make(map[*Client]bool)
			}
			h.clients[client.DeviceID][client] = true
			log.Printf("Hub: registered client, deviceID=%s, isAgent=%v, totalClients=%d",
				client.DeviceID, client.IsAgent, len(h.clients[client.DeviceID]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.DeviceID]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					log.Printf("Hub: unregistered client, deviceID=%s, isAgent=%v, remainingClients=%d",
						client.DeviceID, client.IsAgent, len(clients))
				}
				// 只在没有客户端（包括 Agent 和 Viewer）时才删除 deviceID
				if len(clients) == 0 {
					delete(h.clients, client.DeviceID)
					log.Printf("Hub: deviceID=%s has no clients, removed", client.DeviceID)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(deviceID string) {
	h.unregister <- &Client{DeviceID: deviceID}
}

func (h *Hub) SendToDevice(deviceID string, message []byte) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[deviceID]; ok {
		for client := range clients {
			select {
			case client.Send <- message:
				return true
			default:
				return false
			}
		}
	}
	return false
}

func (h *Hub) BroadcastToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.clients {
		for client := range clients {
			if client.UserID == userID {
				select {
				case client.Send <- message:
				default:
				}
			}
		}
	}
}

// BroadcastToDevice broadcasts message to all clients with the same device ID
// SendToAgents sends message only to Desktop Agent clients (not H5 viewers)
func (h *Hub) SendToAgents(deviceID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[deviceID]; ok {
		for client := range clients {
			if client.IsAgent {
				select {
				case client.Send <- message:
					log.Printf("SendToAgents: sent to agent userID=%d", client.UserID)
				default:
				}
			}
		}
	}
}

// BroadcastToViewers sends message only to the latest H5 viewer (not Desktop Agents)
// This prevents duplicate messages when multiple H5 pages are open
func (h *Hub) BroadcastToViewers(deviceID string, message []byte) {
	h.mu.Lock()
	// Save last output for new viewers
	h.lastOutput[deviceID] = make([]byte, len(message))
	copy(h.lastOutput[deviceID], message)
	h.mu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[deviceID]; ok {
		// Find all H5 viewers and send to each
		var viewers []*Client
		for client := range clients {
			if !client.IsAgent {
				viewers = append(viewers, client)
			}
		}
		log.Printf("BroadcastToViewers: deviceID=%s, viewerCount=%d, msgLen=%d", deviceID, len(viewers), len(message))
		// Send to each viewer
		for _, viewer := range viewers {
			select {
			case viewer.Send <- message:
				log.Printf("BroadcastToViewers: sent to userID=%d", viewer.UserID)
			default:
			}
		}
	}
}

// SendLastOutput sends the last terminal output to a new viewer
func (h *Hub) SendLastOutput(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if output, ok := h.lastOutput[client.DeviceID]; ok && len(output) > 0 {
		select {
		case client.Send <- output:
			log.Printf("SendLastOutput: sent to userID=%d, len=%d", client.UserID, len(output))
		default:
		}
	}
}

// BroadcastToDevice sends to all clients (for backward compatibility)
func (h *Hub) BroadcastToDevice(deviceID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[deviceID]; ok {
		log.Printf("BroadcastToDevice: deviceID=%s, message=%s, clientCount=%d", deviceID, string(message), len(clients))
		for client := range clients {
			select {
			case client.Send <- message:
				log.Printf("BroadcastToDevice: sent to client userID=%d, isAgent=%v", client.UserID, client.IsAgent)
			default:
			}
		}
	}
}
