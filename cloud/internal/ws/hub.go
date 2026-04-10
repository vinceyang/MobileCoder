package ws

import (
	"encoding/json"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mobile-coder/cloud/internal/service"
)

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	Conn        *websocket.Conn
	DeviceID    string
	UserID      int64
	IsAgent     bool   // true for Desktop Agent, false for H5 viewer
	SessionName string // current session name for agent
	Send        chan []byte
}

type Hub struct {
	clients       map[string]map[*Client]bool // key can be deviceID or sessionName
	lastOutput    map[string][]byte           // deviceID -> last terminal output
	recentEvents  map[string][]service.TaskEvent
	lastEventLine map[string]string
	mu            sync.RWMutex
	register      chan *Client
	unregister    chan *Client
}

var ansiSequencePattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func NewHub() *Hub {
	return &Hub{
		clients:       make(map[string]map[*Client]bool),
		lastOutput:    make(map[string][]byte),
		recentEvents:  make(map[string][]service.TaskEvent),
		lastEventLine: make(map[string]string),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Use sessionName as key if available, otherwise use deviceID
			key := client.DeviceID
			if client.SessionName != "" {
				key = client.SessionName
			}
			if h.clients[key] == nil {
				h.clients[key] = make(map[*Client]bool)
			}
			h.clients[key][client] = true
			log.Printf("Hub: registered client, key=%s, isAgent=%v, totalClients=%d",
				key, client.IsAgent, len(h.clients[key]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			// Use sessionName as key if available, otherwise use deviceID
			key := client.DeviceID
			if client.SessionName != "" {
				key = client.SessionName
			}
			if clients, ok := h.clients[key]; ok {
				if _, exists := clients[client]; exists {
					delete(clients, client)
					close(client.Send)
					log.Printf("Hub: unregistered client, key=%s, isAgent=%v, remainingClients=%d",
						key, client.IsAgent, len(clients))
				}
				// 只在没有客户端时删除 key
				if len(clients) == 0 {
					delete(h.clients, key)
					log.Printf("Hub: key=%s has no clients, removed", key)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
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

// SendToAgents sends message only to Desktop Agent clients
// Uses sessionName if provided, otherwise falls back to deviceID
func (h *Hub) SendToAgents(deviceID string, sessionName string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Use sessionName as key if available
	key := deviceID
	if sessionName != "" {
		key = sessionName
	}

	log.Printf("SendToAgents: looking for agents for key=%s, total clients=%d", key, len(h.clients[key]))
	if clients, ok := h.clients[key]; ok {
		for client := range clients {
			log.Printf("SendToAgents: client IsAgent=%v, deviceID=%s, sessionName=%s", client.IsAgent, client.DeviceID, client.SessionName)
			if client.IsAgent {
				select {
				case client.Send <- message:
					log.Printf("SendToAgents: sent to agent userID=%d", client.UserID)
				default:
				}
				// Only send to the first agent to prevent duplicate messages
				return
			}
		}
	}
}

// BroadcastToViewers sends message only to the latest H5 viewer (not Desktop Agents)
// This prevents duplicate messages when multiple H5 pages are open
// Uses sessionName if provided, otherwise falls back to deviceID
func (h *Hub) BroadcastToViewers(deviceID string, sessionName string, message []byte) {
	h.RecordTerminalOutput(deviceID, sessionName, message)

	// Use sessionName as key if available
	key := deviceID
	if sessionName != "" {
		key = sessionName
	}

	h.mu.Lock()
	// Save last output for new viewers
	h.lastOutput[key] = make([]byte, len(message))
	copy(h.lastOutput[key], message)
	h.mu.Unlock()

	h.mu.RLock()
	defer h.mu.RUnlock()
	if clients, ok := h.clients[key]; ok {
		// Find all H5 viewers and send to each
		var viewers []*Client
		for client := range clients {
			if !client.IsAgent {
				viewers = append(viewers, client)
			}
		}
		log.Printf("BroadcastToViewers: key=%s, viewerCount=%d, msgLen=%d", key, len(viewers), len(message))
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

func (h *Hub) RecordTerminalOutput(deviceID string, sessionName string, message []byte) {
	summary := extractLatestTerminalLine(message)
	if summary == "" {
		return
	}

	key := taskKey(deviceID, sessionName)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.lastEventLine[key] == summary {
		return
	}

	h.lastEventLine[key] = summary
	event := service.TaskEvent{
		Summary:   summary,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Kind:      classifyTaskEvent(summary),
	}

	events := append([]service.TaskEvent{event}, h.recentEvents[key]...)
	if len(events) > 10 {
		events = events[:10]
	}
	h.recentEvents[key] = events
}

func (h *Hub) GetRecentEvents(taskID string) []service.TaskEvent {
	h.mu.RLock()
	defer h.mu.RUnlock()

	events := h.recentEvents[taskID]
	if len(events) == 0 {
		return nil
	}

	result := make([]service.TaskEvent, len(events))
	copy(result, events)
	return result
}

// SendLastOutput sends the last terminal output to a new viewer
func (h *Hub) SendLastOutput(client *Client) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	// Use sessionName as key if available, otherwise use deviceID
	key := client.DeviceID
	if client.SessionName != "" {
		key = client.SessionName
	}
	if output, ok := h.lastOutput[key]; ok && len(output) > 0 {
		select {
		case client.Send <- output:
			log.Printf("SendLastOutput: sent to userID=%d, len=%d, key=%s", client.UserID, len(output), key)
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

func taskKey(deviceID string, sessionName string) string {
	if sessionName == "" {
		return deviceID
	}
	return deviceID + ":" + sessionName
}

func extractLatestTerminalLine(message []byte) string {
	var envelope struct {
		Type    string `json:"type"`
		Payload struct {
			Content string `json:"content"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(message, &envelope); err != nil {
		return ""
	}
	if envelope.Type != "terminal_output" {
		return ""
	}

	lines := strings.Split(envelope.Payload.Content, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(ansiSequencePattern.ReplaceAllString(lines[i], ""))
		if line != "" {
			return line
		}
	}
	return ""
}

func classifyTaskEvent(summary string) service.TaskEventKind {
	lower := strings.ToLower(summary)

	switch {
	case strings.Contains(lower, "task completed"), strings.Contains(lower, "completed successfully"), strings.Contains(lower, "done"), strings.Contains(lower, "finished successfully"):
		return service.TaskEventKindCompleted
	case strings.Contains(lower, "waiting for"), strings.Contains(lower, "confirm"), strings.Contains(lower, "press enter"), strings.Contains(lower, "select an option"):
		return service.TaskEventKindNeedsInput
	case strings.Contains(lower, "error"), strings.Contains(lower, "failed"), strings.Contains(lower, "panic"), strings.Contains(lower, "permission denied"):
		return service.TaskEventKindError
	case strings.Contains(lower, "tests passed"), strings.Contains(lower, "all green"), strings.Contains(lower, "test passed"), strings.Contains(lower, "test failed"), strings.Contains(lower, "failing tests"):
		return service.TaskEventKindTestResult
	case strings.Contains(lower, "updating "), strings.Contains(lower, "creating "), strings.Contains(lower, "applying "), strings.Contains(lower, "running "), strings.Contains(lower, "checking "), strings.Contains(lower, "installing "):
		return service.TaskEventKindToolStep
	default:
		return service.TaskEventKindInfo
	}
}
