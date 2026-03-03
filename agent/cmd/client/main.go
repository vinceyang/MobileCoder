package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mobile-coder/agent/internal/client"
)

func generateCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:length]
}

// getDeviceName returns the computer's hostname
func getDeviceName() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		return "Desktop Agent"
	}
	return hostname
}

// getDeviceIDPath returns the path to store device_id (one per computer, not per directory)
func getDeviceIDPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".MobileCoder", "device-id")
}

// getBindCodePath returns the path to store bind_code
func getBindCodePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".MobileCoder", "bind-code")
}

// loadOrCreateDeviceID loads existing device_id or creates a new one
func loadOrCreateDeviceID(serverURL string) (string, string, error) {
	deviceIDPath := getDeviceIDPath()
	bindCodePath := getBindCodePath()

	// Try to load existing device_id and bind_code
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
					// Load bind_code
					bindCode := ""
					if data, err := os.ReadFile(bindCodePath); err == nil {
						bindCode = strings.TrimSpace(string(data))
					}
					return deviceID, bindCode, nil
				}
			}
		}
	}

	// Generate bindCode for initial registration
	bindCode := generateCode(6)

	// Get the computer's hostname
	deviceName := getDeviceName()

	// Register with cloud - cloud will generate deviceID
	resp, err := http.Post("http://"+serverURL+"/api/device/register", "application/json",
		strings.NewReader(fmt.Sprintf(`{"bind_code":"%s","device_name":"%s"}`, bindCode, deviceName)))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	// Use the deviceID returned from cloud
	deviceID, ok := result["device_id"].(string)
	if !ok || deviceID == "" {
		return "", "", fmt.Errorf("failed to get device_id from cloud")
	}

	// Create directory if not exists
	os.MkdirAll(filepath.Dir(deviceIDPath), 0755)

	// Save device_id and bind_code to file
	os.WriteFile(deviceIDPath, []byte(deviceID), 0644)
	os.WriteFile(bindCodePath, []byte(bindCode), 0644)

	return deviceID, bindCode, nil
}

// keyToTmux 将按键映射到 tmux 格式（使用转义序列）
func keyToTmux(key string, modifiers []interface{}) string {
	// 使用转义序列的按键映射
	// 使用 raw string 避免转义问题
	keyMap := map[string]string{
		"Enter":     string(rune(13)),  // \r
		"Tab":       string(rune(9)),   // \t
		"Escape":    string(rune(27)),   // \e
		"Backspace": string(rune(127)),  // DEL
		"Delete":    "\x1b[3~",
		"Up":        "\x1b[A",
		"Down":      "\x1b[B",
		"Right":     "\x1b[C",
		"Left":     "\x1b[D",
		"Home":     "\x1b[H",
		"End":      "\x1b[F",
		"PageUp":   "\x1b[5~",
		"PageDown": "\x1b[6~",
		"F1":       "\x1bOP",
		"F2":       "\x1bOQ",
		"F3":       "\x1bOR",
		"F4":       "\x1bOS",
		"F5":       "\x1b[15~",
		"F6":       "\x1b[17~",
		"F7":       "\x1b[18~",
		"F8":       "\x1b[19~",
		"F9":       "\x1b[20~",
		"F10":      "\x1b[21~",
		"F11":      "\x1b[23~",
		"F12":      "\x1b[24~",
	}

	tmuxKey, exists := keyMap[key]
	if !exists {
		tmuxKey = key
	}

	// 处理修饰键
	hasCtrl := false
	hasShift := false

	for _, m := range modifiers {
		if mod, ok := m.(string); ok {
			if mod == "ctrl" {
				hasCtrl = true
			} else if mod == "shift" {
				hasShift = true
			}
		}
	}

	// 对于 Ctrl+ 组合，使用 C- 格式
	if hasCtrl && len(key) == 1 {
		// 转为 Ctrl+字母 (a-z 对应 \x01-\x1a)
		lowerKey := strings.ToLower(key)
		tmuxKey = string(rune(lowerKey[0] - 'a' + 1))
	} else if hasShift && len(key) == 1 {
		// Shift+ 字母
		tmuxKey = strings.ToUpper(key)
	}

	return tmuxKey
}

func main() {
	serverURL := flag.String("server", "localhost:8080", "Cloud server URL")
	flag.Parse()

	// 使用 device_id 持久化
	deviceID, bindCode, err := loadOrCreateDeviceID(*serverURL)
	if err != nil {
		log.Fatalf("Failed to load/create device ID: %v", err)
	}

	// 如果有绑定码，说明需要绑定
	if bindCode != "" {
		// 显示绑定信息
		fmt.Println("==========================================")
		fmt.Println("  请在 H5 页面输入以下绑定码:")
		fmt.Println("==========================================")
		fmt.Printf("  绑定码: %s\n", bindCode)
		fmt.Println("==========================================")
		fmt.Println("\n请在 H5 页面输入绑定码后按回车继续...")
		var input string
		fmt.Scanln(&input)
	} else {
		// 已绑定，直接重连
		fmt.Println("设备已绑定，自动重连中...")
	}

	// WebSocket 连接
	ws, err := client.NewWSClient("ws://"+*serverURL+"/ws", deviceID)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 更新设备名称（如果与当前主机名不同）
	deviceName := getDeviceName()
	go func() {
		updateData := fmt.Sprintf(`{"device_id":"%s","device_name":"%s"}`, deviceID, deviceName)
		resp, err := http.Post("http://"+*serverURL+"/api/device/update", "application/json", strings.NewReader(updateData))
		if err == nil {
			defer resp.Body.Close()
		}
	}()

	// 创建 tmux 会话
	sessionName := fmt.Sprintf("claude-%s", deviceID[:6])

	// 获取当前工作目录作为项目路径
	cwd, _ := os.Getwd()
	projectPath := cwd

	// 检查 tmux session 是否已存在
	cmd := exec.Command("tmux", "has-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		// 创建新的 tmux session 并在其中运行 claude（移除 CLAUDECODE 环境变量）
		exec.Command("tmux", "new-session", "-d", "-s", sessionName, "env", "-u", "CLAUDECODE", "claude", "--dangerously-skip-permissions").Run()
	} else {
		// session 已存在，发送 Ctrl+C 停止当前，然后发送继续命令
		exec.Command("tmux", "send-keys", "-t", sessionName, "C-c").Run()
		time.Sleep(300 * time.Millisecond)
		exec.Command("tmux", "send-keys", "-t", sessionName, "env", "-u", "CLAUDECODE", "claude", "-c", "--dangerously-skip-permissions", "\r").Run()
	}

	// 向服务器注册 session
	sessionJSON := fmt.Sprintf(`{"device_id":"%s","session_name":"%s","project_path":"%s"}`, deviceID, sessionName, projectPath)
	resp, err := http.Post("http://"+*serverURL+"/api/sessions", "application/json", strings.NewReader(sessionJSON))
	if err == nil {
		defer resp.Body.Close()
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		if sessionID, ok := result["session_id"].(float64); ok {
			log.Printf("Session registered: %d, name: %s", int(sessionID), sessionName)
		}
	}

	// 捕获终端输出并发送到 H5
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		var lastContent string

		for range ticker.C {
			// 捕获 tmux pane 内容（-e 保留转义序列）
			cmd := exec.Command("tmux", "capture-pane", "-t", sessionName, "-p", "-e")
			out, err := cmd.Output()
			if err != nil {
				continue
			}
			output := string(out)

			// 只发送有变化的内容
			if output != lastContent && output != "" {
				lastContent = output
				ws.Send("terminal_output", map[string]interface{}{
					"content": output,
				})
			}
		}
	}()

	// 处理 H5 输入
	go func() {
		ws.OnMessage(func(data []byte) {
			var msg map[string]interface{}
			json.Unmarshal(data, &msg)
			log.Printf("Received WS message: %s", string(data))
			if msg["type"] == "terminal_input" {
				payload := msg["payload"].(map[string]interface{})

				// 检查是否是特殊按键
				action, _ := payload["action"].(string)
				if action == "key" {
					// 处理特殊按键
					key, _ := payload["key"].(string)
					modifiers, _ := payload["modifiers"].([]interface{})

					tmuxKey := keyToTmux(key, modifiers)
					log.Printf("H5 key: %s, modifiers: %v -> tmux: %s", key, modifiers, tmuxKey)

					// 发送按键到 tmux，使用 -l  literal 模式发送转义序列
					exec.Command("tmux", "send-keys", "-t", sessionName, "-l", tmuxKey).Run()

					// 特殊按键不需要额外 Enter
					// 这些按键已经包含执行
					isSpecialKey := key == "Enter" || key == "Tab" || key == "Escape" ||
						key == "Up" || key == "Down" || key == "Left" || key == "Right" ||
						key == "Backspace" || key == "Delete" || key == "Home" || key == "End" ||
						key == "PageUp" || key == "PageDown" ||
						strings.HasPrefix(key, "F")

					if !isSpecialKey && len(key) > 1 {
						// 非特殊的功能键（如 Ctrl+A）发送后需要 Enter 执行
						exec.Command("tmux", "send-keys", "-t", sessionName, "-l", "\r").Run()
					}
				} else {
					// 处理普通文本输入
					content, _ := payload["content"].(string)
					log.Printf("H5 input: %q, session: %s", content, sessionName)

					// 去掉末尾的换行符，单独发送
					content = strings.TrimSuffix(content, "\n")

					// 发送内容
					exec.Command("tmux", "send-keys", "-t", sessionName, "-l", content).Run()
					// 发送回车
					exec.Command("tmux", "send-keys", "-t", sessionName, "-l", "\r").Run()

					log.Printf("Sent to tmux")
				}
			}
		})
	}()

	// 提示
	fmt.Println("\nClaude Code 已在 tmux 会话中启动!")
	fmt.Printf("查看终端: tmux attach -t %s\n", sessionName)
	fmt.Println("退出 tmux: 按 Ctrl+B 然后按 D")
	fmt.Println("\nH5 页面应该能看到终端输出")

	// 保持运行
	select {}
}
