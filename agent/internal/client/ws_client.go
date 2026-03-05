package client

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn       *websocket.Conn
	deviceID   string
	sessionName string
	serverURL  string
	mu         sync.Mutex
	onMessage  func(msg []byte)
	reconnect  bool
}

func NewWSClient(serverURL, deviceID, sessionName string) (*WSClient, error) {
	ws := &WSClient{
		serverURL:   serverURL,
		deviceID:    deviceID,
		sessionName: sessionName,
		reconnect:   true,
	}
	if err := ws.connect(); err != nil {
		return nil, err
	}
	return ws, nil
}

func (c *WSClient) connect() error {
	url := c.serverURL + "?device_id=" + c.deviceID
	if c.sessionName != "" {
		url += "&session_name=" + c.sessionName
	}
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *WSClient) OnMessage(handler func(msg []byte)) {
	c.onMessage = handler
	go c.readPump()
}

func (c *WSClient) readPump() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if c.reconnect {
				c.reconnectLoop()
			}
			return
		}
		if c.onMessage != nil {
			c.onMessage(msg)
		}
	}
}

func (c *WSClient) reconnectLoop() {
	backoff := time.Second
	maxBackoff := 30 * time.Second

	for {
		log.Printf("Attempting to reconnect...")
		if err := c.connect(); err != nil {
			log.Printf("Reconnect failed: %v, retrying in %v", err, backoff)
			time.Sleep(backoff)
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		log.Printf("Reconnected successfully")
		// 重连成功后恢复读取
		go c.readPump()
		return
	}
}

func (c *WSClient) Send(msgType string, payload interface{}) error {
	msg := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}
	data, _ := json.Marshal(msg)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return websocket.ErrCloseSent
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) SendRaw(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return websocket.ErrCloseSent
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
