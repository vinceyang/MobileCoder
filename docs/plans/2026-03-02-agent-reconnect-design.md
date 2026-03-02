# Agent 自动重连与 Session 恢复设计

## 目标

1. Agent 断开连接后自动重连，无需重新绑定和启动 Claude
2. 启动 Claude 时默认继续该目录的最后一次 session

## 背景

当前实现中：
- Agent 每次启动生成新的 device_id，H5 需要重新绑定
- Agent 每次启动都创建新的 tmux session 和新的 Claude session

## 方案设计

### 1. Device ID 持久化

**存储路径**：`~/.mobile-coder/{项目目录名}/device-id`

**首次启动流程**：
1. 检查 `~/.mobile-coder/{dir}/device-id` 是否存在
2. 如果不存在，调用 `/api/device/register` 获取新的 device_id
3. 将 device_id 写入文件

**重连流程**：
1. 读取本地 device_id
2. 调用 `/api/device/check` 验证 device_id 是否仍然有效
3. 如果有效，直接建立 WebSocket 连接
4. 如果无效，重新走注册流程

### 2. WebSocket 自动重连

**重连机制**：
- 监听 WebSocket 连接断开事件
- 指数退避策略：1s, 2s, 4s, 8s, 最大 30s
- 重连成功恢复 tmux 输出捕获

**保持元连接**：
- 不重新创建 tmux session
- 不重新启动 Claude
- H5 端无需感知底层断开

### 3. Claude Session 恢复

**使用 `claude -c` 继续**：
- 首次启动：`claude` + `--dangerously-skip-permissions`
- 后续启动：`claude -c` + `--dangerously-skip-permissions`

**Session 存储**：
- Claude Code 自动在 `.claude/sessions/` 目录存储 session
- `-c` 参数会自动继续当前目录的最后一次 session

## 文件结构

```
~/.mobile-coder/
├── projectA/
│   └── device-id          # device_id 文件
├── projectB/
│   └── device-id
└── config.json            # 全局配置（可选）
```

## API 变更

### 新增接口

**POST /api/device/check**
- 功能：检查 device_id 是否有效
- 请求：`{"device_id": "xxx"}`
- 响应：`{"valid": true, "status": "online"}`

**POST /api/device/reregister**
- 功能：使用已有 device_id 重新注册（用于 Agent 重启后恢复）
- 请求：`{"device_id": "xxx", "device_name": "Desktop Agent"}`
- 响应：`{"device_id": "xxx", "status": "online"}`

## 实现任务

1. 修改 agent/main.go 添加 device_id 持久化逻辑
2. 修改 agent/main.go 添加 WebSocket 重连机制
3. 修改 agent/main.go 使用 `claude -c` 继续 session
4. cloud 端添加 device check/reregister 接口

## 风险与限制

- 如果用户删除 `~/.mobile-coder/` 目录，需要重新绑定
- 如果 Claude 的 session 被外部删除，`-c` 会启动新 session
