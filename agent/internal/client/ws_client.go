package client

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn     *websocket.Conn
	deviceID string
}

func NewWSClient(serverURL, deviceID string) (*WSClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL+"?device_id="+deviceID, nil)
	if err != nil {
		return nil, err
	}
	return &WSClient{conn: conn, deviceID: deviceID}, nil
}

func (c *WSClient) OnMessage(handler func(msg []byte)) {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("WS read error: %v", err)
			return
		}
		handler(msg)
	}
}

func (c *WSClient) Send(msgType string, payload interface{}) error {
	msg := map[string]interface{}{
		"type":    msgType,
		"payload": payload,
	}
	data, _ := json.Marshal(msg)
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) SendRaw(data []byte) error {
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}
