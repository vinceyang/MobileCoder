# AI 编程工具移动化 - 实现计划 (复用版)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现移动端通过 H5 页面远程查看和控制桌面端 AI 编程工具的终端输出

**Architecture:** 中心化服务模式 - 云端管理用户和设备绑定，Desktop Agent 通过 WebSocket 连接到云端，移动端 H5 通过 WebSocket 接收终端输出并发送指令

**Tech Stack:** Go, WebSocket (gorilla/websocket), Next.js, Supabase, **复用现有 agentapi 的 PTY 终端模块**

---

## 复用的代码资产

| 现有模块 | 复用方式 |
|---------|---------|
| `lib/termexec/` | Desktop Agent 的 PTY 终端启动和管理 |
| `lib/screentracker/` | 终端屏幕快照、稳定性检测、消息解析 |
| `lib/msgfmt/` | 消息格式化、Agent 类型适配 |
| `chat/` | H5 前端基础（Next.js 项目） |
| `cmd/server/` | HTTP 服务器架构参考 |

### Supabase 配置

使用 Supabase 作为云端数据库（PostgreSQL）：

**环境变量：**
```bash
DB_HOST=your-project.supabase.co
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=agentapi
```

**表结构：**
- `users` - 用户表
- `devices` - 设备表（user_id 外键关联 users）

---

## 阶段一：云端服务（新增）

### Task 1: 创建云端服务项目

**Files:**
- Create: `cloud/go.mod`
- Create: `cloud/cmd/server/main.go`
- Create: `cloud/internal/config/config.go`
- Create: `cloud/internal/db/db.go`

**Step 1: 创建 cloud 目录结构**

```bash
mkdir -p cloud/cmd/server cloud/internal/{config,db,handler,service,middleware,ws}
cd cloud
go mod init github.com/coder/agentapi/cloud
go add github.com/gorilla/websocket
go add github.com/golang-jwt/jwt/v5
go add github.com/lib/pq
go add github.com/coder/agentapi
```

**Step 2: 创建 config.go - 支持 Supabase 配置**

```go
// cloud/internal/config/config.go
package config

import (
	"os"
)

type Config struct {
	Port      string
	JWTSecret string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPassword string
	DBName    string
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "agentapi-secret-key"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "agentapi"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

**Step 3: 创建 db.go - 连接 Supabase PostgreSQL**

```go
// cloud/internal/db/db.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadDBConfig() *Config {
	return &Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
	}
}

func InitDB(cfg *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// 创建表（如不存在）
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS devices (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL REFERENCES users(id),
		device_id VARCHAR(255) UNIQUE NOT NULL,
		device_name VARCHAR(255),
		bind_code VARCHAR(255),
		bind_code_exp TIMESTAMP,
		status VARCHAR(50) DEFAULT 'offline',
		last_active_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Supabase database")
	return db, nil
}
```

**Step 4: 提交**

```bash
git add cloud/
git commit -m "feat(cloud): create cloud service project structure"
```

---

### Task 2: 实现认证服务（复用较少，主要新增）

**Files:**
- Create: `cloud/internal/service/auth_service.go`
- Create: `cloud/internal/handler/auth_handler.go`

**Step 1: 创建 auth_service.go - PostgreSQL 版本**

```go
// cloud/internal/service/auth_service.go
package service

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrInvalidCreds = errors.New("invalid credentials")
)

type User struct {
	ID       int64
	Username string
	Password string
	Email    string
}

type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthService struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{db: db, jwtSecret: jwtSecret}
}

func (s *AuthService) Register(username, password, email string) (*User, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	var id int64
	err := s.db.QueryRow(
		"INSERT INTO users (username, password, email) VALUES ($1, $2, $3) RETURNING id",
		username, string(hashedPassword), email,
	).Scan(&id)

	if err != nil {
		return nil, ErrUserExists
	}

	return &User{ID: id, Username: username, Email: email}, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	var user User
	err := s.db.QueryRow(
		"SELECT id, username, password, email FROM users WHERE username = $1",
		username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Email)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return "", ErrInvalidCreds
	}

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, _ := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	claims, _ := token.Claims.(*Claims)
	return claims, nil
}
```

**Step 2: 创建 auth_handler.go**

```go
// cloud/internal/handler/auth_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"github.com/coder/agentapi/cloud/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	json.NewDecoder(r.Body).Decode(&req)

	user, err := h.authService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
```

**Step 3: 提交**

```bash
git add cloud/
git commit -m "feat(cloud): implement authentication service"
```

---

### Task 3: 设备绑定服务

**Files:**
- Create: `cloud/internal/service/device_service.go`
- Create: `cloud/internal/handler/device_handler.go`

**Step 1: 创建 device_service.go - PostgreSQL 版本**

```go
// cloud/internal/service/device_service.go
package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"
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
	db *sql.DB
}

func NewDeviceService(db *sql.DB) *DeviceService {
	return &DeviceService{db: db}
}

func generateCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func (s *DeviceService) CreateBindCode(userID int64, deviceName string) (*Device, error) {
	deviceID := generateCode(16)
	bindCode := generateCode(6)

	device := &Device{
		UserID:      userID,
		DeviceID:    deviceID,
		DeviceName:  deviceName,
		BindCode:    bindCode,
		BindCodeExp: time.Now().Add(10 * time.Minute),
		Status:      "offline",
	}

	_, err := s.db.Exec(
		`INSERT INTO devices (user_id, device_id, device_name, bind_code, bind_code_exp, status) VALUES ($1, $2, $3, $4, $5, $6)`,
		device.UserID, device.DeviceID, device.DeviceName, device.BindCode, device.BindCodeExp, device.Status,
	)
	return device, err
}

func (s *DeviceService) BindDevice(userID int64, bindCode string) (*Device, error) {
	var device Device
	err := s.db.QueryRow(
		`SELECT id, user_id, device_id, device_name, bind_code, bind_code_exp, status FROM devices WHERE bind_code = $1`,
		bindCode,
	).Scan(&device.ID, &device.UserID, &device.DeviceID, &device.DeviceName, &device.BindCode, &device.BindCodeExp, &device.Status)

	if err != nil {
		return nil, ErrDeviceNotFound
	}

	if time.Now().After(device.BindCodeExp) {
		return nil, ErrBindCodeExpired
	}

	// Clear bind code after successful binding
	s.db.Exec("UPDATE devices SET bind_code = NULL, bind_code_exp = NULL, status = 'online' WHERE id = $1", device.ID)
	device.Status = "online"

	return &device, nil
}

func (s *DeviceService) GetUserDevices(userID int64) ([]Device, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, device_id, device_name, status FROM devices WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		rows.Scan(&d.ID, &d.UserID, &d.DeviceID, &d.DeviceName, &d.Status)
		devices = append(devices, d)
	}
	return devices, nil
}
```

**Step 2: 提交**

```bash
git add cloud/
git commit -m "feat(cloud): implement device binding service"
```

---

### Task 4: WebSocket 服务（核心）

**Files:**
- Create: `cloud/internal/ws/hub.go` - WebSocket hub 管理
- Create: `cloud/internal/handler/ws_handler.go` - WebSocket 处理器

**Step 1: 创建 hub.go**

```go
// cloud/internal/ws/hub.go
package ws

import (
	"encoding/json"
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
	Send     chan []byte
}

type Hub struct {
	clients    map[string]*Client
	mu         sync.RWMutex
	register   chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.DeviceID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.DeviceID]; ok {
				delete(h.clients, client.DeviceID)
				close(c.Send)
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
	if client, ok := h.clients[deviceID]; ok {
		select {
		case client.Send <- message:
			return true
		default:
			return false
		}
	}
	return false
}

func (h *Hub) BroadcastToUser(userID int64, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}
```

**Step 2: 创建 ws_handler.go**

```go
// cloud/internal/handler/ws_handler.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/coder/agentapi/cloud/internal/ws"
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
	if deviceID == "" {
		http.Error(w, "device_id required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WS upgrade error: %v", err)
		return
	}

	client := &ws.Client{
		Conn:     conn,
		DeviceID: deviceID,
		Send:     make(chan []byte, 256),
	}

	h.hub.Register(client)

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

		// Forward to all user clients (mobile)
		var msg ws.Message
		if json.Unmarshal(message, &msg) == nil {
			// Store device_id in message for routing
			h.hub.BroadcastToUser(client.UserID, message)
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
```

**Step 3: 提交**

```bash
git add cloud/
git commit -m "feat(cloud): implement WebSocket hub service"
```

---

### Task 5: 云端主服务器

**Files:**
- Modify: `cloud/cmd/server/main.go`

**Step 1: 创建 main.go - 支持 Supabase 配置**

```go
// cloud/cmd/server/main.go
package main

import (
	"log"
	"net/http"

	"github.com/coder/agentapi/cloud/internal/config"
	"github.com/coder/agentapi/cloud/internal/db"
	"github.com/coder/agentapi/cloud/internal/handler"
	"github.com/coder/agentapi/cloud/internal/service"
	"github.com/coder/agentapi/cloud/internal/ws"
)

func main() {
	cfg := config.Load()

	// 连接 Supabase PostgreSQL
	dbCfg := &db.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}

	database, err := db.InitDB(dbCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Initialize services
	authService := service.NewAuthService(database, cfg.JWTSecret)
	deviceService := service.NewDeviceService(database)
	hub := ws.NewHub()

	// Start WebSocket hub
	go hub.Run()

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	deviceHandler := handler.NewDeviceHandler(authService, deviceService)
	wsHandler := handler.NewWSHubHandler(hub)

	// Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/register", authHandler.Register)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/device/create-bind-code", deviceHandler.CreateBindCode)
	mux.HandleFunc("/api/device/bind", deviceHandler.BindDevice)
	mux.HandleFunc("/api/device/list", deviceHandler.ListDevices)
	mux.HandleFunc("/ws", wsHandler.HandleConnection)

	log.Printf("Cloud server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}
```

**环境变量配置：**
```bash
# Supabase 配置
export DB_HOST="your-project.supabase.co"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="your-db-password"
export DB_NAME="agentapi"

# 服务配置
export PORT="8080"
export JWT_SECRET="your-jwt-secret"
```

**Step 2: 提交**

```bash
git add cloud/
git commit -m "feat(cloud): implement main server with routes"
```

---

## 阶段二：Desktop Agent（大量复用）

### Task 6: Desktop Agent - 复用 lib/termexec 和 lib/screentracker

**Files:**
- Create: `agent/cmd/client/main.go`
- Create: `agent/internal/client/ws_client.go`

**Step 1: 创建 agent 模块，复用 agentapi**

```bash
mkdir -p agent/cmd/client agent/internal/client
cd agent
go mod init github.com/coder/agentapi/agent
go add github.com/coder/agentapi
go add github.com/gorilla/websocket
```

**Step 2: 创建 ws_client.go**

```go
// agent/internal/client/ws_client.go
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
		"type": msgType,
		"payload": payload,
	}
	data, _ := json.Marshal(msg)
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}
```

**Step 3: 创建 main.go - 核心复用 termexec 和 screentracker**

```go
// agent/cmd/client/main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/coder/agentapi/agent/internal/client"
	"github.com/coder/agentapi/lib/msgfmt"
	"github.com/coder/agentapi/lib/screentracker"
	"github.com/coder/agentapi/lib/termexec"
)

func main() {
	serverURL := flag.String("server", "ws://localhost:8080", "Cloud server URL")
	deviceID := flag.String("device-id", "", "Device ID")
	agentType := flag.String("agent", "claude", "Agent type (claude, aider, etc.)")
	flag.Parse()

	if *deviceID == "" {
		log.Fatal("device-id is required")
	}

	ctx := context.Background()

	// 复用 agentapi 的 termexec 启动 AI 工具
	agentIO, err := termexec.StartProcess(ctx, termexec.StartProcessConfig{
		Program:        *agentType,
		TerminalWidth:  80,
		TerminalHeight: 24,
	})
	if err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}

	// 复用 agentapi 的 screentracker 处理终端输出
	conversation, err := screentracker.NewPTYConversation(ctx, screentracker.PTYConversationConfig{
		AgentIO:   agentIO,
		AgentType: msgfmt.AgentType(*agentType),
	})
	if err != nil {
		log.Fatalf("Failed to create conversation: %v", err)
	}

	// 连接到云端
	wsURL := "ws://" + *serverURL + "/ws"
	ws, err := client.NewWSClient(wsURL, *deviceID)
	if err != nil {
		log.Fatalf("Failed to connect to cloud: %v", err)
	}
	defer ws.Close()

	// 复用 screentracker 的消息回调，将终端输出发送到云端
	conversation.OnMessage(func(msg screentracker.ConversationMessage) {
		ws.Send("terminal_output", map[string]string{
			"content": msg.Content,
		})
	})

	// 接收云端指令并写入终端
	ws.OnMessage(func(data []byte) {
		var msg map[string]interface{}
		json.Unmarshal(data, &msg)
		if msg["type"] == "terminal_input" {
			input := msg["payload"].(map[string]interface{})["content"].(string)
			agentIO.Write([]byte(input))
		}
	})

	log.Println("Desktop Agent started with", *agentType)
	select {}
}
```

**Step 4: 提交**

```bash
git add agent/
git commit -m "feat(agent): create Desktop Agent using existing termexec and screentracker"
```

---

## 阶段三：移动端 H5（复用 chat/）

### Task 7: 移动端 H5 - 复用 chat/ 项目

**Files:**
- Modify: `chat/app/page.tsx` - 添加设备绑定和终端显示
- Create: `chat/app/components/BindPage.tsx`
- Create: `chat/app/components/Terminal.tsx`

**Step 1: 创建 BindPage.tsx**

```tsx
// chat/app/components/BindPage.tsx
'use client';

import { useState } from 'react';

export default function BindPage({ onBind }: { onBind: (deviceId: string) => void }) {
  const [bindCode, setBindCode] = useState('');
  const [loading, setLoading] = useState(false);

  const handleBind = async () => {
    setLoading(true);
    try {
      const token = localStorage.getItem('token');
      const res = await fetch('http://localhost:8080/api/device/bind', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({ bind_code: bindCode }),
      });
      const data = await res.json();
      if (data.device_id) {
        onBind(data.device_id);
      }
    } catch (err) {
      console.error(err);
    }
    setLoading(false);
  };

  return (
    <div className="min-h-screen bg-gray-900 flex flex-col items-center justify-center p-4">
      <h1 className="text-2xl font-bold text-white mb-8">绑定设备</h1>
      <input
        type="text"
        value={bindCode}
        onChange={(e) => setBindCode(e.target.value)}
        placeholder="输入绑定码"
        className="w-full max-w-md px-4 py-3 bg-gray-800 border border-gray-700 rounded-lg mb-4 text-white"
      />
      <button
        onClick={handleBind}
        disabled={loading}
        className="px-8 py-3 bg-blue-600 text-white rounded-lg disabled:opacity-50"
      >
        {loading ? '绑定中...' : '绑定'}
      </button>
    </div>
  );
}
```

**Step 2: 创建 Terminal.tsx**

```tsx
// chat/app/components/Terminal.tsx
'use client';

import { useEffect, useRef, useState } from 'react';

interface TerminalProps {
  deviceId: string;
}

export default function Terminal({ deviceId }: TerminalProps) {
  const [output, setOutput] = useState<string[]>([]);
  const [input, setInput] = useState('');
  const wsRef = useRef<WebSocket | null>(null);
  const outputRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const token = localStorage.getItem('token');
    const ws = new WebSocket(`ws://localhost:8080/ws?device_id=${deviceId}&token=${token}`);

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === 'terminal_output') {
        setOutput(prev => [...prev, msg.payload.content]);
      }
    };
    wsRef.current = ws;

    return () => ws.close();
  }, [deviceId]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (outputRef.current) {
      outputRef.current.scrollTop = outputRef.current.scrollHeight;
    }
  }, [output]);

  const handleSend = () => {
    if (!input.trim() || !wsRef.current) return;
    wsRef.current.send(JSON.stringify({
      type: 'terminal_input',
      payload: { content: input + '\n' }
    }));
    setInput('');
  };

  return (
    <div className="h-screen bg-gray-900 flex flex-col">
      <div ref={outputRef} className="flex-1 overflow-auto p-4 font-mono text-sm text-green-400 whitespace-pre-wrap">
        {output.map((line, i) => (
          <div key={i}>{line}</div>
        ))}
      </div>
      <div className="flex p-4 border-t border-gray-800">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          className="flex-1 px-4 py-2 bg-gray-800 border border-gray-700 rounded-l-lg text-white"
          placeholder="输入指令..."
        />
        <button
          onClick={handleSend}
          className="px-6 py-2 bg-blue-600 text-white rounded-r-lg"
        >
          发送
        </button>
      </div>
    </div>
  );
}
```

**Step 3: 修改 chat/app/page.tsx**

```tsx
// chat/app/page.tsx
'use client';

import { useState, useEffect } from 'react';
import BindPage from './components/BindPage';
import Terminal from './components/Terminal';

export default function Home() {
  const [deviceId, setDeviceId] = useState<string | null>(null);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    const savedToken = localStorage.getItem('token');
    const savedDeviceId = localStorage.getItem('device_id');
    setToken(savedToken);
    if (savedDeviceId) setDeviceId(savedDeviceId);
  }, []);

  const handleBind = (id: string) => {
    localStorage.setItem('device_id', id);
    setDeviceId(id);
  };

  if (!token) {
    return (
      <div className="min-h-screen bg-gray-900 flex items-center justify-center">
        <p className="text-white">请先登录</p>
      </div>
    );
  }

  if (!deviceId) {
    return <BindPage onBind={handleBind} />;
  }

  return <Terminal deviceId={deviceId} />;
}
```

**Step 4: 提交**

```bash
git add chat/
git commit -m "feat(chat): add mobile terminal interface with device binding"
```

---

## 实现顺序

1. Task 1: 云端项目结构 + 数据库
2. Task 2: 认证服务
3. Task 3: 设备绑定服务
4. Task 4: WebSocket 服务
5. Task 5: 云端主服务器
6. Task 6: Desktop Agent（**复用 termexec + screentracker**）
7. Task 7: 移动端 H5（**复用 chat/ 项目**）

---

**代码复用总结**

| 任务 | 复用模块 |
|------|---------|
| Desktop Agent | `lib/termexec`, `lib/screentracker`, `lib/msgfmt` |
| 移动端 H5 | `chat/` 整个 Next.js 项目 |
| 云端服务 | 新建（无现有可复用） |

---

**Plan complete and saved to `docs/plans/2026-02-25-ai-coding-mobile-plan.md`.**

Two execution options:

1. **Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

2. **Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
