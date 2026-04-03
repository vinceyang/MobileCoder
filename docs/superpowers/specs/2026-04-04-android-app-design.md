# MobileCoder Android App 设计文档

**日期**: 2026-04-04

## 概述

为 MobileCoder 项目开发 Android 客户端，让用户可以在手机上管理设备、查看终端输出、发送指令。

## 技术方案

### 架构选择

| 项目 | 选择 | 说明 |
|------|------|------|
| 平台 | Android | iOS 后续兼容 |
| 框架 | Capacitor | 将 H5 打包成原生 App |
| 项目位置 | `mobile-app/` | 独立于现有 H5 项目 |
| H5 更新 | 热更新 | 自动同步最新功能 |
| WebView | 远程加载 | 始终加载服务器 H5 |

### 技术栈

- **前端框架**: React 18 + TypeScript
- **UI 组件**: Tailwind CSS + shadcn/ui
- **原生封装**: Capacitor 7
- **状态管理**: React Context
- **推送通知**: Firebase Cloud Messaging (FCM)
- **本地存储**: Capacitor Preferences

## 项目结构

```
mobile-app/
├── src/
│   ├── App.tsx              # 主应用入口
│   ├── main.tsx            # React 渲染入口
│   ├── components/         # 公共组件
│   ├── pages/              # 页面组件
│   ├── services/           # API 服务
│   ├── stores/             # 状态管理
│   └── theme/              # 主题配置
├── android/                # Android 原生项目 (自动生成)
├── capacitor.config.ts     # Capacitor 配置
├── package.json
└── tsconfig.json
```

## 功能模块

### 1. 登录注册

**功能**:
- 邮箱 + 密码登录
- 用户注册
- Token 持久化存储

**API 端点** (复用现有):
- `POST /api/auth/register` - 注册
- `POST /api/auth/login` - 登录

**流程**:
1. 用户输入邮箱、密码
2. 调用登录 API，获取 token
3. Token 存储到 Capacitor Preferences
4. 自动登录（Token 有效期内）

### 2. 设备管理

**功能**:
- 设备列表展示
- 设备详情查看
- Session 列表
- 设备绑定

**API 端点** (复用现有):
- `GET /api/devices` - 获取用户设备列表
- `GET /api/devices/sessions` - 获取设备 Sessions
- `POST /api/device/bind` - 绑定设备

**UI 页面**:
- `/devices` - 设备列表
- `/devices/:id` - 设备详情 + Session 列表

### 3. 终端

**功能**:
- WebSocket 实时连接
- 终端输出展示
- 发送指令/按键
- 快捷键面板

**API 端点**:
- `WS /ws?device_id=xxx&session_name=xxx&token=xxx`

**UI 页面**:
- `/terminal` - 终端页面

**核心组件**:
- WebSocket 连接管理
- ANSI 转 HTML 渲染
- 快捷键发送

### 4. 推送通知

**功能**:
- Claude Code 执行完成通知
- Agent 断开连接提醒

**实现**:
- Firebase Cloud Messaging (FCM)
- 设备注册时上传 FCM Token
- 服务端推送通知到指定设备

**服务端改动** (需扩展):
- 添加 FCM Token 存储
- 添加推送通知接口

### 5. 本地存储

**存储内容**:
- `token` - 用户认证 Token
- `user_id` - 用户 ID
- `fcm_token` - FCM 推送 Token
- `settings` - 用户偏好设置

## Capacitor 配置

```typescript
// capacitor.config.ts
import { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.mobilecoder.app',
  appName: 'MobileCoder',
  webDir: 'dist',
  server: {
    // 加载远程服务器 H5
    url: 'http://121.41.69.142:3001',
    cleartext: true,
  },
  plugins: {
    PushNotifications: {
      presentationOptions: ['badge', 'sound', 'alert'],
    },
  },
};

export default config;
```

## 热更新配置

```typescript
// 使用 @capacitor/haptics 或 @capacitor/local-notifications
// 配合服务端版本检测实现热更新
```

## 实施计划

### Phase 1: 基础搭建
1. 创建 Capacitor 项目
2. 配置 Android 平台
3. 实现登录/注册页面
4. 实现设备列表页面
5. 实现设备详情 + Session 页面

### Phase 2: 终端功能
1. 实现 Terminal 页面
2. WebSocket 连接
3. 快捷键面板
4. ANSI 渲染

### Phase 3: 原生功能
1. 集成 FCM 推送
2. 本地存储封装
3. App 图标配置
4. 打包签名

## 依赖包

```json
{
  "dependencies": {
    "@capacitor/core": "^7.0.0",
    "@capacitor/android": "^7.0.0",
    "@capacitor/push-notifications": "^7.0.0",
    "@capacitor/preferences": "^7.0.0",
    "@capacitor/haptics": "^7.0.0"
  }
}
```

## 现有 H5 复用

项目将复用现有 H5 的组件和服务:
- `chat/src/app/login/page.tsx` → 登录页面
- `chat/src/app/devices/page.tsx` → 设备列表
- `chat/src/app/devices/[deviceId]/page.tsx` → 设备详情
- `chat/src/app/terminal/page.tsx` → 终端页面
- `chat/src/components/Terminal.tsx` → 终端组件

**复用策略**:
1. 直接复制现有组件到 `mobile-app/src/`
2. 根据移动端需求调整样式
3. 接入 Capacitor 原生能力

## Logo

- 名称: MobileCoder
- 图标: 沿用现有 `chat/public/icon.svg`
- 蓝色背景 + 白色 "M" 字母

## 后续扩展

- iOS 客户端 (共用 H5 代码)
- 扫码绑定设备
- 深色模式支持
- 多语言支持
