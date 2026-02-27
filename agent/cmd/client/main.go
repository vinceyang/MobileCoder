package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/mobile-coder/agent/internal/client"
)

func generateCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:length]
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

	// 本地生成绑定码
	bindCode := generateCode(6)
	fmt.Println("==========================================")
	fmt.Println("  请在 H5 页面输入以下绑定码:")
	fmt.Println("==========================================")
	fmt.Printf("  绑定码: %s\n", bindCode)
	fmt.Println("==========================================")
	fmt.Println("  绑定成功后程序将启动 Claude Code")
	fmt.Println("  同时在终端和 H5 页面都可以与 Claude 交互")
	fmt.Println("==========================================")

	// 调用 API 注册设备
	resp, err := http.Post("http://"+*serverURL+"/api/device/register", "application/json",
		strings.NewReader(`{"bind_code":"`+bindCode+`","device_name":"Desktop Agent"}`))
	if err != nil {
		log.Fatalf("Failed to register: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	deviceID, _ := result["device_id"].(string)

	// 等待用户绑定
	fmt.Println("\n请在 H5 页面输入绑定码后按回车继续...")
	var input string
	fmt.Scanln(&input)

	// WebSocket 连接
	ws, err := client.NewWSClient("ws://"+*serverURL+"/ws", deviceID)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 创建 tmux 会话
	sessionName := fmt.Sprintf("claude-%s", bindCode[:6])

	// 杀掉可能存在的旧会话
	exec.Command("tmux", "kill-session", "-t", sessionName).Run()

	// 创建新的 tmux 会话并在其中运行 claude
	exec.Command("tmux", "new-session", "-d", "-s", sessionName).Run()

	// 设置 256 色支持
	exec.Command("tmux", "set-option", "-t", sessionName, "default-terminal", "screen-256color").Run()

	// 发送命令到 tmux
	exec.Command("tmux", "send-keys", "-t", sessionName, "claude").Run()
	exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()
	exec.Command("tmux", "send-keys", "-t", sessionName, "--dangerously-skip-permissions").Run()
	exec.Command("tmux", "send-keys", "-t", sessionName, "C-j").Run()

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
