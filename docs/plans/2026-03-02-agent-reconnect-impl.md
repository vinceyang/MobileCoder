# Agent 重连与 Session 恢复实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 实现 Agent 自动重连和 Claude Session 恢复功能

**架构：** Agent 端使用本地文件存储 device_id，按项目目录隔离；WebSocket 断开后自动重连；使用 `claude -c` 继续上次 session

**技术栈：** Go, Gorilla WebSocket, tmux

---

## 任务清单

### 任务 1: Cloud 端添加 device check 接口

**Files:**
- Modify: `cloud/internal/handler/device_handler.go`
- Modify: `cloud/internal/service/device_service.go`

**Step 1: 添加 DeviceCheckRequest 结构体和 CheckDevice 方法**

在 `cloud/internal/handler/device_handler.go` 添加:

```go
type DeviceCheckRequest struct {
    DeviceID string `json:"device_id"`
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
```

**Step 2: 在 service 层添加 GetDeviceByDeviceID 方法**

在 `cloud/internal/service/device_service.go` 添加:

```go
// GetDeviceByDeviceID gets a device by device_id
func (s *DeviceService) GetDeviceByDeviceID(deviceID string) (*Device, error) {
    device, err := s.db.GetDeviceByDeviceID(deviceID)
    if err != nil {
        return nil, ErrDeviceNotFound
    }
    return &Device{
        ID:         device.ID,
        UserID:     device.UserID,
        DeviceID:   device.DeviceID,
        DeviceName: device.DeviceName,
        Status:     device.Status,
    }, nil
}
```

**Step 3: 在 db 层添加 GetDeviceByDeviceID 方法**

在 `cloud/internal/db/supabase.go` 添加:

```go
func (s *SupabaseDB) GetDeviceByDeviceID(deviceID string) (*Device, error) {
    resp, err := s.do("GET", "/devices?device_id=eq."+deviceID, nil)
    if err != nil {
        return nil, err
    }

    var devices []Device
    json.Unmarshal(resp, &devices)
    if len(devices) == 0 {
        return nil, fmt.Errorf("device not found")
    }
    return &devices[0], nil
}
```

**Step 4: 注册路由**

在 `cloud/cmd/server/main.go` 添加路由:

```go
mux.HandleFunc("/api/device/check", deviceHandler.CheckDevice)
```

**Step 5: 测试编译**

```bash
cd /Users/yangxq/Code/MobileCoder/cloud && go build -o bin/cloud ./cmd/server
```

Expected: 编译成功

---

### 任务 2: Agent 端添加 device_id 持久化

**Files:**
- Modify: `agent/cmd/client/main.go`

**Step 1: 添加 device_id 存储相关函数**

在 main.go 添加:

```go
import (
    "os"
    "path/filepath"
)

// getDeviceIDPath returns the path to store device_id for the current project
func getDeviceIDPath() string {
    // Get current working directory
    cwd, _ := os.Getwd()
    dirName := filepath.Base(cwd)
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, ".mobile-coder", dirName, "device-id")
}

// loadOrCreateDeviceID loads existing device_id or creates a new one
func loadOrCreateDeviceID(serverURL string) (string, error) {
    deviceIDPath := getDeviceIDPath()

    // Try to load existing device_id
    if data, err := os.ReadFile(deviceIDPath); err == nil {
        deviceID := strings.TrimSpace(string(data))
        if deviceID != "" {
            // Check if device_id is still valid
            resp, err := http.Post("http://"+serverURL+"/api/device/check", "application/json",
                strings.NewReader(`{"device_id":"`+deviceID+`"}`))
            if err == nil {
                defer resp.Body.Close()
                var result map[string]interface{}
                json.NewDecoder(resp.Body).Decode(&result)
                if valid, ok := result["valid"].(bool); ok && valid {
                    return deviceID, nil
                }
            }
        }
    }

    // Generate new device_id
    deviceID := generateCode(16)

    // Create directory if not exists
    os.MkdirAll(filepath.Dir(deviceIDPath), 0755)

    // Save to file
    os.WriteFile(deviceIDPath, []byte(deviceID), 0644)

    // Register with cloud
    resp, err := http.Post("http://"+serverURL+"/api/device/register", "application/json",
        strings.NewReader(`{"bind_code":"`+bindCode+`","device_name":"Desktop Agent"}`))
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    return deviceID, nil
}
```

**Step 2: 修改 main 函数使用 device_id 持久化**

替换注册设备部分:

```go
// 替换原来的注册逻辑
deviceID, err := loadOrCreateDeviceID(*serverURL)
if err != nil {
    log.Fatalf("Failed to load/create device ID: %v", err)
}
```

**Step 3: 移除 generateCode(6) 改为使用目录名作为绑定码显示**

由于 device_id 已经持久化，绑定码主要用于首次绑定显示:

```go
// 显示绑定信息（仅首次需要）
fmt.Println("==========================================")
fmt.Println("  请在 H5 页面输入以下绑定码:")
fmt.Println("==========================================")
// 从 device_id 前6位作为便捷识别码
fmt.Printf("  设备码: %s\n", deviceID[:6])
fmt.Println("==========================================")
fmt.Println("  首次绑定后，后续启动将自动重连")
fmt.Println("==========================================")
```

**Step 4: 编译测试**

```bash
cd /Users/yangxq/Code/MobileCoder/agent && go build -o bin/agent ./cmd/client
```

Expected: 编译成功

---

### 任务 3: Agent 端添加 WebSocket 自动重连

**Files:**
- Modify: `agent/internal/client/ws_client.go`
- Modify: `agent/cmd/client/main.go`

**Step 1: 修改 WSClient 添加重连机制**

在 `agent/internal/client/ws_client.go` 添加:

```go
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
    serverURL  string
    mu         sync.Mutex
    onMessage  func(msg []byte)
    reconnect  bool
}

func NewWSClient(serverURL, deviceID string) (*WSClient, error) {
    ws := &WSClient{
        serverURL: serverURL,
        deviceID:  deviceID,
        reconnect: true,
    }
    if err := ws.connect(); err != nil {
        return nil, err
    }
    return ws, nil
}

func (c *WSClient) connect() error {
    conn, _, err := websocket.DefaultDialer.Dial(c.serverURL+"?device_id="+c.deviceID, nil)
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
```

**Step 2: 修改 main.go 中的 WS 使用**

```go
// WebSocket 连接（支持自动重连）
ws, err := client.NewWSClient("ws://"+*serverURL+"/ws", deviceID)
if err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

**Step 3: 编译测试**

```bash
cd /Users/yangxq/Code/MobileCoder/agent && go build -o bin/agent ./cmd/client
```

Expected: 编译成功

---

### 任务 4: Agent 端使用 `claude -c` 继续 session

**Files:**
- Modify: `agent/cmd/client/main.go`

**Step 1: 修改启动 Claude 的逻辑**

替换原来的启动命令:

```go
// 检查是否需要恢复 tmux session
sessionName := fmt.Sprintf("claude-%s", deviceID[:6])

// 检查 tmux session 是否已存在
cmd := exec.Command("tmux", "has-session", "-t", sessionName)
if err := cmd.Run(); err != nil {
    // session 不存在，创建并启动 Claude
    exec.Command("tmux", "new-session", "-d", "-s", sessionName).Run()
    exec.Command("tmux", "set-option", "-t", sessionName, "default-terminal", "screen-256color").Run()

    // 首次启动使用 claude（不带 -c）
    exec.Command("tmux", "send-keys", "-t", sessionName, "claude").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "--dangerously-skip-permissions").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
} else {
    // session 已存在，使用 claude -c 继续
    exec.Command("tmux", "send-keys", "-t", sessionName, "claude").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "-c").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "--dangerously-skip-permissions").Run()
    exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
}
```

**Step 2: 编译测试**

```bash
cd /Users/yangxq/Code/MobileCoder/agent && go build -o bin/agent ./cmd/client
```

Expected: 编译成功

---

### 任务 5: 整体测试

**Step 1: 启动 cloud 服务**

```bash
./cloud/bin/cloud
```

**Step 2: 启动 agent（首次）**

```bash
./agent/bin/agent -server localhost:8080
```

Expected:
- 首次启动显示设备码
- H5 绑定成功后启动 Claude

**Step 3: 模拟断开（Ctrl+C 停止 agent）**

**Step 4: 重新启动 agent**

```bash
./agent/bin/agent -server localhost:8080
```

Expected:
- 不再显示绑定码提示
- 自动重连到之前的设备
- Claude session 被恢复（使用 -c）

**Step 5: 提交代码**

```bash
git add -A && git commit -m "feat: add agent reconnect and session resume"
```

---

## 执行方式

**Plan complete and saved to `docs/plans/2026-03-02-agent-reconnect-design.md`.**

Two execution options:

1. **Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

2. **Parallel Session (separate)** - Open new session with executing_plans, batch execution with checkpoints

Which approach?
