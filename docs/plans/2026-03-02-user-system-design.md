# 用户体系设计方案

> 日期: 2026-03-02

## 方案概述

引入用户账号体系，实现用户、设备、Session 三个层级的管理：
- 用户：邮箱 + 密码登录，最多绑定 5 台设备
- 设备：Agent 所在的物理电脑（device_id 标识），一台设备只需绑定一次
- Session：设备上的 tmux 会话（每个项目一个，可同时运行多个）

## 核心概念

| 概念 | 说明 |
|------|------|
| 用户 | 邮箱 + 密码登录，最多绑定 5 台物理设备 |
| 设备 | Agent 所在的物理电脑，首次启动需要用户绑定，后续自动重连 |
| Session | 设备上的 tmux 会话（每个项目一个），用户可选择连接 |

## 数据模型

### Users 表
```sql
id          -- 主键
email       -- 邮箱（唯一）
password    -- 密码（加密存储）
created_at  -- 创建时间
```

### Devices 表
```sql
id              -- 主键
user_id         -- 关联用户（外键）
device_id       -- Agent 生成的唯一标识
device_name     -- 设备名称（如 "MacBook Pro"）
status          -- 在线状态 (online/offline)
last_active_at -- 最后活跃时间
created_at      -- 创建时间
```

### Sessions 表
```sql
id              -- 主键
device_id       -- 关联设备（外键）
session_name    -- tmux 会话名（如 "claude-mobilecoder"）
project_path    -- 项目路径
status          -- 状态 (active/inactive)
created_at      -- 创建时间
```

## API 设计

### 认证 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录（返回 token） |

### 设备 API
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/devices | 获取当前用户的所有设备 |
| POST | /api/devices | 添加设备（通过绑定码） |
| DELETE | /api/devices/{id} | 解绑设备 |
| GET | /api/devices/{id} | 获取设备详情 |
| GET | /api/devices/{id}/sessions | 获取设备的所有 Session |

### Session API
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/sessions/{id} | 获取 Session 详情 |
| POST | /api/sessions/{id}/connect | 连接到指定 Session |

### WebSocket
| 路径 | 说明 |
|------|------|
| /ws?session_id={id}&token={token} | 连接到指定 Session |

## 交互流程

### 1. 用户注册/登录
```
用户 -> H5 注册/登录页面 -> POST /api/auth/register/login -> 返回 token
```

### 2. 首次启动 Agent（需要绑定）
```
Agent 启动 -> POST /api/device/register -> 返回 device_id + bind_code
Agent 显示绑定码 -> 用户登录 H5 -> 输入绑定码 -> POST /api/device/bind
-> 服务器绑定 user_id 到 device -> 绑定成功
```

### 3. 后续启动 Agent（自动重连）
```
Agent 启动 -> 检查本地 device_id -> POST /api/device/check
-> 有效 -> 直接连接 WebSocket -> 无需用户操作
```

### 4. 用户使用 H5
```
用户登录 H5 -> GET /api/devices -> 显示所有设备
点击设备 -> GET /api/devices/{id}/sessions -> 显示所有 Session
选择 Session -> WebSocket 连接 -> 操作终端
```

## 绑定方式

### 方式一：绑定码（当前方式）
- Agent 启动后显示 6 位绑定码
- 用户在 H5 登录后输入绑定码
- 服务器将设备绑定到当前用户

### 方式二：绑定链接（推荐）
- Agent 启动后显示绑定链接：`https://xxx.com/bind?device=xxxxx&token=yyyyy`
- token 包含用户身份信息
- 用户在 H5 已登录状态下点击链接，直接绑定

## 设计原则

1. **设备一次绑定**：物理设备只需绑定一次，后续启动自动重连
2. **用户多设备**：一个用户可管理多台设备（最多 5 台）
3. **多 Session**：一台设备可运行多个 Claude Session（按项目区分）
4. **渐进式实现**：优先实现核心功能，后续迭代优化

## 实现优先级

1. **Phase 1**：用户注册/登录
2. **Phase 2**：设备绑定到用户
3. **Phase 3**：H5 查看设备列表和 Session
4. **Phase 4**：连接 Session
