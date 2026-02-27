# 用户登录和设备管理设计方案

> 日期: 2026-02-26

## 方案概述

采用渐进式方案：
- **第一阶段**：去掉用户登录，扫码即用，设备粒度绑定
- **第二阶段**：可选账号体系（未来）

## 当前实现（将被移除/简化）

### 1. 后端 API

#### 移除的 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/auth/register | 用户注册 |
| POST | /api/auth/login | 用户登录 |
| POST | /api/device/bind | 用户绑定设备（需 token） |

#### 保留/修改的 API
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /api/device/register | Desktop Agent 注册（生成绑定码） |
| GET | /api/device/list | 列出用户设备（移除 token 验证） |
| POST | /api/device/bind | H5 绑定设备（无需 token） |

### 2. 数据库

#### Users 表
- 保留但不使用（未来账号体系备用）
- 可选择删除或归档

#### Devices 表
- 移除 user_id 字段
- bind_code 永久有效（移除过期逻辑）
- 设备独立存在，不依赖账号

### 3. 服务层

#### 移除的 Service
- `auth_service.go` - 用户认证服务
- 相关的 JWT token 生成/验证

#### 保留的 Service
- `device_service.go` - 设备管理服务（简化）

### 4. 前端

#### 移除的页面/组件
- LoginPage.tsx - 登录页面
- RegisterPage - 注册页面（如有）

#### 简化的页面/组件
- BindPage.tsx - 移除 token 验证，直接绑定
- Terminal.tsx - 无需登录即可访问

## 简化后的流程

### Desktop Agent 启动
```
1. Desktop Agent 启动
2. 调用 POST /api/device/register
3. 获取 device_id 和 bind_code
4. 显示绑定码供用户扫码
```

### H5 用户绑定
```
1. 用户访问 H5 页面（无需登录）
2. 输入绑定码
3. 调用 POST /api/device/bind（无需 token）
4. 直接进入终端页面
5. 绑定永久有效
```

## 未来可选：账号体系

当需要多设备同步时，可选增加：

### 功能
- 云账号注册/登录（可选）
- 设备数据同步
- 多设备管理

### 保留的数据库设计
- Users 表结构保持不变
- Devices 表可选择增加 user_id（可选关联）

## 实现步骤

1. 移除 auth_handler.go 相关路由
2. 简化 device_handler.go，移除 token 验证
3. 简化 device_service.go，移除用户关联逻辑
4. 移除/归档 auth_service.go
5. 简化前端，移除登录页面
6. 保留 Users 表但不使用
