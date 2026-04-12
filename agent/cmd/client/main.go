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

// AI coding tool types
type AIClient string

const (
	AIClientClaude AIClient = "claude"
	AIClientCodex  AIClient = "codex"
	AIClientCursor AIClient = "cursor"
)

// AI tool configuration
type ToolConfig struct {
	Name        string   // Command name
	CheckCmd    string   // Command to check if installed
	CheckArgs   []string // Args for version check
	StartArgs   []string // Args to start the tool
	InstallHint string   // Hint for installation
}

var toolConfigs = map[AIClient]ToolConfig{
	AIClientClaude: {
		Name:        "claude",
		CheckCmd:    "claude",
		CheckArgs:   []string{"--version"},
		StartArgs:   []string{"--dangerously-skip-permissions"},
		InstallHint: "npm install -g @anthropic-ai/claude-code",
	},
	AIClientCodex: {
		Name:        "codex",
		CheckCmd:    "codex",
		CheckArgs:   []string{"--version"},
		StartArgs:   []string{"--c"},
		InstallHint: "npm install -g @openai/codex or see https://docs.codex.dev",
	},
	AIClientCursor: {
		Name:        "cursor",
		CheckCmd:    "agent",
		CheckArgs:   []string{"--version"},
		StartArgs:   []string{"--c"},
		InstallHint: "Download from https://cursor.sh",
	},
}

// checkTool checks if the AI tool is installed and available
func checkTool(tool AIClient) error {
	config, ok := toolConfigs[tool]
	if !ok {
		return fmt.Errorf("unknown AI tool: %s", tool)
	}

	// Check if command exists
	cmd := exec.Command("which", config.CheckCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s not found. Install with: %s", config.Name, config.InstallHint)
	}

	// Try to get version
	cmd = exec.Command(config.CheckCmd, config.CheckArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: %s --version failed: %v", config.Name, err)
		// Still allow running if command exists
	} else {
		log.Printf("%s version: %s", config.Name, strings.TrimSpace(string(output)))
	}

	return nil
}

// getToolCommand returns the command and args to start the AI tool
func getToolCommand(tool AIClient, projectPath string) (string, []string) {
	switch tool {
	case AIClientClaude:
		// Claude Code: need to remove CLAUDECODE env var and add --dangerously-skip-permissions
		return "env", []string{"-u", "CLAUDECODE", "claude", "--dangerously-skip-permissions"}
	case AIClientCodex:
		// Codex: run without args (interactive mode)
		return "codex", []string{}
	case AIClientCursor:
		// Cursor (agent): run without args (interactive mode)
		return "agent", []string{}
	default:
		return string(tool), []string{"--c"}
	}
}

// checkDependencies checks if all required dependencies are installed
func checkDependencies() error {
	// Check tmux
	cmd := exec.Command("which", "tmux")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tmux not found. Install with: brew install tmux (macOS) or apt install tmux (Linux)")
	}

	// Check tmux version
	cmd = exec.Command("tmux", "-V")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Warning: tmux -V failed: %v", err)
	} else {
		log.Printf("tmux version: %s", strings.TrimSpace(string(output)))
	}

	return nil
}

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

func getAgentTokenPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".MobileCoder", "agent-token")
}

type deviceCheckResponse struct {
	Valid      bool   `json:"valid"`
	Bound      bool   `json:"bound"`
	Status     string `json:"status"`
	AgentToken string `json:"agent_token"`
}

func checkDevice(serverURL, deviceID, bindCode string) (deviceCheckResponse, error) {
	payload := map[string]string{"device_id": deviceID}
	if bindCode != "" {
		payload["bind_code"] = bindCode
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return deviceCheckResponse{}, err
	}
	resp, err := http.Post("http://"+serverURL+"/api/device/check", "application/json",
		strings.NewReader(string(body)))
	if err != nil {
		return deviceCheckResponse{}, err
	}
	defer resp.Body.Close()

	var result deviceCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return deviceCheckResponse{}, err
	}
	return result, nil
}

func clearBindCode() error {
	err := os.Remove(getBindCodePath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func loadAgentToken() string {
	data, err := os.ReadFile(getAgentTokenPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveAgentToken(token string) error {
	if err := os.MkdirAll(filepath.Dir(getAgentTokenPath()), 0755); err != nil {
		return err
	}
	return os.WriteFile(getAgentTokenPath(), []byte(token), 0644)
}

func waitForDeviceBinding(serverURL, deviceID, bindCode string, timeout, interval time.Duration) error {
	if bindCode == "" {
		return nil
	}

	deadline := time.Now().Add(timeout)
	for {
		result, err := checkDevice(serverURL, deviceID, bindCode)
		if err == nil && result.Valid && result.Bound && result.AgentToken != "" {
			if err := saveAgentToken(result.AgentToken); err != nil {
				return err
			}
			return clearBindCode()
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("binding timed out for code %s", bindCode)
		}
		time.Sleep(interval)
	}
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
			result, err := checkDevice(serverURL, deviceID, "")
			if err == nil && result.Valid {
				if loadAgentToken() != "" {
					if err := clearBindCode(); err != nil {
						return "", "", err
					}
					return deviceID, "", nil
				}

				bindCode := ""
				if data, err := os.ReadFile(bindCodePath); err == nil {
					bindCode = strings.TrimSpace(string(data))
				}
				if bindCode == "" {
					return "", "", fmt.Errorf("device %s requires rebind to obtain agent token", deviceID)
				}
				return deviceID, bindCode, nil
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

func terminalInputToTmuxCommands(sessionName string, payload map[string]interface{}) [][]string {
	action, _ := payload["action"].(string)
	if action == "key" {
		key, _ := payload["key"].(string)
		modifiers, _ := payload["modifiers"].([]interface{})
		return [][]string{tmuxKeyCommand(sessionName, key, modifiers)}
	}

	content, _ := payload["content"].(string)
	content = strings.TrimRight(content, "\r\n")

	commands := make([][]string, 0, 2)
	if content != "" {
		commands = append(commands, []string{"send-keys", "-t", sessionName, "-l", content})
	}
	commands = append(commands, []string{"send-keys", "-t", sessionName, "C-m"})
	return commands
}

func tmuxKeyCommand(sessionName string, key string, modifiers []interface{}) []string {
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

	if hasCtrl && len(key) == 1 {
		return []string{"send-keys", "-t", sessionName, "C-" + strings.ToLower(key)}
	}

	if hasShift && key == "Tab" {
		return []string{"send-keys", "-t", sessionName, "BTab"}
	}

	keyMap := map[string]string{
		"Enter":     "C-m",
		"Tab":       "Tab",
		"Escape":    "Escape",
		"Backspace": "BSpace",
		"Delete":    "Delete",
		"Up":        "Up",
		"Down":      "Down",
		"Right":     "Right",
		"Left":      "Left",
		"Home":      "Home",
		"End":       "End",
		"PageUp":    "PageUp",
		"PageDown":  "PageDown",
		"F1":        "F1",
		"F2":        "F2",
		"F3":        "F3",
		"F4":        "F4",
		"F5":        "F5",
		"F6":        "F6",
		"F7":        "F7",
		"F8":        "F8",
		"F9":        "F9",
		"F10":       "F10",
		"F11":       "F11",
		"F12":       "F12",
	}
	if tmuxKey, ok := keyMap[key]; ok {
		return []string{"send-keys", "-t", sessionName, tmuxKey}
	}

	if hasShift && len(key) == 1 {
		key = strings.ToUpper(key)
	}
	return []string{"send-keys", "-t", sessionName, "-l", key}
}

func isLiteralTmuxInput(args []string) bool {
	for _, arg := range args {
		if arg == "-l" {
			return true
		}
	}
	return false
}

func main() {
	serverURL := flag.String("server", "localhost:8080", "Cloud server URL")
	aiTool := flag.String("ai", "claude", "AI coding tool: claude, codex, cursor")
	flag.Parse()

	// Check dependencies first
	fmt.Println("==========================================")
	fmt.Println("  MobileCoder Desktop Agent")
	fmt.Println("==========================================")
	fmt.Println()

	// Check tmux dependency (required)
	fmt.Print("Checking tmux... ")
	if err := checkDependencies(); err != nil {
		fmt.Printf("FAILED\n%v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")

	// Check all known AI tools (warnings only, not errors)
	fmt.Println("Checking available AI tools...")
	for t, config := range toolConfigs {
		cmd := exec.Command("which", config.CheckCmd)
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s: not found (hint: %s)\n", config.Name, config.InstallHint)
		} else {
			fmt.Printf("  %s: found\n", config.Name)
		}
		_ = t // suppress unused variable warning
	}
	fmt.Println()

	// Parse and check the specified AI tool (must exist)
	tool := AIClient(*aiTool)
	if _, ok := toolConfigs[tool]; !ok {
		fmt.Printf("Error: unknown AI tool '%s'. Available options: claude, codex, cursor\n", *aiTool)
		os.Exit(1)
	}

	// Check if the specified AI tool is installed
	fmt.Printf("Checking %s (specified)... ", tool)
	if err := checkTool(tool); err != nil {
		fmt.Printf("FAILED\n%v\n", err)
		os.Exit(1)
	}
	fmt.Println("OK")
	fmt.Println()

	// Check device registration
	fmt.Println("Connecting to server...")
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
		fmt.Println()
		fmt.Println("等待 H5 页面完成绑定...")
		if err := waitForDeviceBinding(*serverURL, deviceID, bindCode, 10*time.Minute, 2*time.Second); err != nil {
			log.Fatalf("Device binding failed: %v", err)
		}
		fmt.Println("设备绑定成功，继续启动...")
	} else {
		// 已绑定，直接重连
		fmt.Println("设备已绑定，自动重连中...")
	}

	// 获取当前工作目录作为项目路径
	cwd, _ := os.Getwd()
	projectPath := cwd

	// 创建 tmux 会话名，包含目录名以区分不同项目
	dirName := filepath.Base(cwd)
	if dirName == "" || dirName == "/" {
		dirName = "root"
	}
	// 清理目录名，移除非法字符
	dirName = strings.ReplaceAll(dirName, "/", "-")
	dirName = strings.ReplaceAll(dirName, " ", "_")
	sessionName := fmt.Sprintf("%s-%s-%s", tool, deviceID[:6], dirName)
	log.Printf("Agent starting with tool=%s, sessionName=%s, deviceID=%s, dirName=%s", tool, sessionName, deviceID, dirName)

	// WebSocket 连接
	log.Printf("Connecting to WebSocket with sessionName=%s", sessionName)
	ws, err := client.NewWSClient("ws://"+*serverURL+"/ws", deviceID, sessionName)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// 更新设备名称（如果与当前主机名不同）
	deviceName := getDeviceName()
	go func() {
		updateData := fmt.Sprintf(`{"device_id":"%s","device_name":"%s"}`, deviceID, deviceName)
		req, err := http.NewRequest("POST", "http://"+*serverURL+"/api/device/update", strings.NewReader(updateData))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")
		if token := loadAgentToken(); token != "" {
			req.Header.Set("Authorization", token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
		}
	}()

	// Get command for the selected AI tool
	cmdName, cmdArgs := getToolCommand(tool, projectPath)

	// 设置较大的历史记录缓冲，避免长输出被截断
	historyLimit := 5000

	// 检查 tmux session 是否已存在
	cmd := exec.Command("tmux", "-u", "has-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		// 创建新的 tmux session 并在其中运行 AI 工具
		if tool == AIClientClaude {
			// Claude Code: 使用 env -u CLAUDECODE 移除环境变量
			fullArgs := []string{"-u", "new-session", "-d", "-s", sessionName, "env", "-u", "CLAUDECODE", cmdName}
			fullArgs = append(fullArgs, cmdArgs...)
			exec.Command("tmux", fullArgs...).Run()
		} else {
			// Codex/Cursor: 直接运行
			fullArgs := []string{"-u", "new-session", "-d", "-s", sessionName, cmdName}
			fullArgs = append(fullArgs, cmdArgs...)
			exec.Command("tmux", fullArgs...).Run()
		}
		// 设置历史记录大小
		exec.Command("tmux", "-u", "set-option", "-t", sessionName, "history-limit", fmt.Sprintf("%d", historyLimit)).Run()
	} else {
		// session 已存在时只接管，不重启。对 Codex/Claude 发送 Ctrl+C 可能会让
		// 唯一 pane 退出并销毁 tmux session，导致 H5 看起来在线但无法接收输入。
		exec.Command("tmux", "-u", "set-option", "-t", sessionName, "history-limit", fmt.Sprintf("%d", historyLimit)).Run()
	}

	// 向服务器注册 session
	sessionJSON := fmt.Sprintf(`{"device_id":"%s","session_name":"%s","project_path":"%s"}`, deviceID, sessionName, projectPath)
	log.Printf("Registering session: %s", sessionJSON)
	req, err := http.NewRequest("POST", "http://"+*serverURL+"/api/sessions", strings.NewReader(sessionJSON))
	if err != nil {
		log.Printf("Session registration request build failed: %v", err)
	} else {
		req.Header.Set("Content-Type", "application/json")
		if token := loadAgentToken(); token != "" {
			req.Header.Set("Authorization", token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			defer resp.Body.Close()
			var result map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&result)
			if sessionID, ok := result["session_id"].(float64); ok {
				log.Printf("Session registered: %d, name: %s", int(sessionID), sessionName)
			} else if resp.StatusCode >= 400 {
				log.Printf("Session registration failed: status=%d body=%v", resp.StatusCode, result)
			}
		} else {
			log.Printf("Session registration failed: %v", err)
		}
	}

	// 捕获终端输出并发送到 H5
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		var lastContent string

		for range ticker.C {
			// 捕获 tmux 历史记录（完整历史，不只是可见区域）
			// -S -5000 从最后 5000 行开始捕获
			cmd := exec.Command("tmux", "-u", "capture-pane", "-t", sessionName, "-p", "-e", "-S", "-5000")
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
				for _, args := range terminalInputToTmuxCommands(sessionName, payload) {
					log.Printf("tmux %s", strings.Join(args, " "))
					if err := exec.Command("tmux", args...).Run(); err != nil {
						log.Printf("tmux send failed: %v", err)
					}
					if isLiteralTmuxInput(args) {
						time.Sleep(150 * time.Millisecond)
					}
				}
			}
		})
	}()

	// 提示
	fmt.Printf("\n%s 已在 tmux 会话中启动!\n", strings.Title(string(tool)))
	fmt.Printf("查看终端: tmux attach -t %s\n", sessionName)
	fmt.Println("退出 tmux: 按 Ctrl+B 然后按 D")
	fmt.Println("\nH5 页面应该能看到终端输出")

	// 保持运行
	select {}
}
