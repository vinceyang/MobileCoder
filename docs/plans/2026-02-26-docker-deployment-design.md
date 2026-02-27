# Docker 部署方案

## 1. 概述

本文档描述将 Mobile Coder 项目（cloud 服务端 + chat 前端）部署到云服务器的完整方案。

### 1.1 部署目标

- **部署位置**: 云服务器（阿里云/腾讯云）
- **部署模式**: 单服务器 + Docker 容器化
- **网络访问**: 使用 IP + 端口访问

### 1.2 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        云服务器                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │                    Nginx (Docker)                    │   │
│  │   ┌───────────────────┐    ┌───────────────────┐   │   │
│  │   │  localhost:8080    │    │  localhost:3001  │   │   │
│  │   │  (Cloud API)       │    │  (Chat H5)        │   │   │
│  │   └───────────────────┘    └───────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│         │                        │                        │
│    外部端口 8080            外部端口 3001                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────┐     WebSocket      ┌─────────────────┐
│  Desktop Agent  │ ◄────────────────► │   Cloud Server  │
│  (用户本地)      │                    │   (Docker)      │
└─────────────────┘                    └─────────────────┘
                                                │
                                                │ WebSocket
                                                │
                                        ┌──────▼──────┐
                                        │  Chat H5    │
                                        │ (浏览器)    │
                                        └─────────────┘
```

## 2. 技术方案

### 2.1 组件说明

| 组件 | 镜像 | 端口 | 说明 |
|------|------|------|------|
| cloud | 本地构建 | 8080 | Go WebSocket 服务端 |
| chat | 本地构建 | 3000 | Next.js H5 前端 |
| nginx | nginx:latest | 80, 8080, 3001 | 反向代理 |

### 2.2 网络配置

- **服务器端口 8080** → Cloud API (WebSocket + HTTP)
- **服务器端口 3001** → Chat H5 前端
- **Nginx** 作为反向代理，统一管理入口

### 2.3 数据存储

- 使用 Supabase 云服务，无需本地数据库
- 无需持久化存储

## 3. 部署流程

### 3.1 服务器初始化

```bash
# 1. 安装 Docker
curl -fsSL https://get.docker.com | sh

# 2. 启动 Docker
systemctl start docker
systemctl enable docker

# 3. 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 3.2 本地构建

```bash
# 1. 克隆项目
git clone <repository-url>
cd agentapi

# 2. 构建 Cloud 服务端镜像
cd cloud
docker build -t agentapi/cloud:latest .

# 3. 构建 Chat 前端镜像
cd ../chat
docker build -t agentapi/chat:latest .

# 4. 导出镜像（可选，传输到服务器）
docker save -o cloud.tar agentapi/cloud:latest
docker save -o chat.tar agentapi/chat:latest
```

### 3.3 服务器部署

```bash
# 1. 上传镜像到服务器
scp cloud.tar user@server:/tmp/
scp chat.tar user@server:/tmp/
scp deploy/docker-compose.yml user@server:/tmp/

# 2. 服务器加载镜像
docker load -i /tmp/cloud.tar
docker load -i /tmp/chat.tar

# 3. 创建目录
ssh user@server
mkdir -p /opt/agentapi

# 4. 上传配置文件
# (使用 scp 或其他方式将 docker-compose.yml 上传到服务器)

# 5. 启动服务
cd /opt/agentapi
docker-compose up -d
```

### 3.4 验证部署

```bash
# 检查容器状态
docker-compose ps

# 检查日志
docker-compose logs -f

# 验证服务
curl http://localhost:8080/health
curl http://localhost:3001
```

## 4. 配置文件

### 4.1 docker-compose.yml

```yaml
version: '3.8'

services:
  cloud:
    image: agentapi/cloud:latest
    container_name: agentapi-cloud
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - SUPABASE_PROJECT_URL=${SUPABASE_PROJECT_URL:-}
      - SUPABASE_API_KEY=${SUPABASE_API_KEY:-}
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  chat:
    image: agentapi/chat:latest
    container_name: agentapi-chat
    ports:
      - "3001:3000"
    environment:
      - PORT=3000
    restart: unless-stopped

  nginx:
    image: nginx:latest
    container_name: agentapi-nginx
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - cloud
      - chat
    restart: unless-stopped
```

### 4.2 nginx.conf

```nginx
events {
    worker_connections 1024;
}

http {
    # Cloud API (WebSocket + HTTP)
    upstream cloud_backend {
        server cloud:8080;
    }

    # Chat H5
    upstream chat_backend {
        server chat:3000;
    }

    server {
        listen 80;
        server_name _;

        # Cloud API
        location / {
            proxy_pass http://cloud_backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Chat H5
        location /chat/ {
            proxy_pass http://chat_backend/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_http_version 1.1;
        }
    }
}
```

### 4.3 Cloud Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080

CMD ["./server"]
```

### 4.4 Chat Dockerfile

```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app
COPY package.json package-lock.json* ./
RUN npm ci --only=production

COPY . .
ENV GITHUB_PAGES=true
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000

ENV PORT=3000
CMD ["node", "server.js"]
```

## 5. 发布流程

### 5.1 更新部署

```bash
# 1. 本地重新构建镜像
docker build -t agentapi/cloud:latest ./cloud
docker build -t agentapi/chat:latest ./chat

# 2. 导出镜像
docker save -o cloud.tar agentapi/cloud:latest
docker save -o chat.tar agentapi/chat:latest

# 3. 上传到服务器
scp cloud.tar chat.tar user@server:/tmp/

# 4. 服务器加载并重启
ssh user@server
docker load -i /tmp/cloud.tar
docker load -i /tmp/chat.tar
cd /opt/agentapi
docker-compose up -d --build

# 5. 验证
docker-compose ps
docker-compose logs -f
```

### 5.2 回滚

```bash
# 查看历史镜像（如果有打 tag）
docker images agentapi/cloud

# 回滚到指定版本
docker tag agentapi/cloud:1.2.0 agentapi/cloud:latest
docker tag agentapi/chat:1.2.0 agentapi/chat:latest

# 重启
cd /opt/agentapi
docker-compose up -d
```

## 6. 运维

### 6.1 常用命令

```bash
# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f
docker-compose logs -f cloud
docker-compose logs -f chat

# 重启服务
docker-compose restart
docker-compose restart cloud

# 停止服务
docker-compose down

# 更新并重启
docker-compose up -d --build
```

### 6.2 日志管理

```bash
# 查看最近 100 行日志
docker-compose logs --tail=100

# 限制日志文件大小（编辑 daemon.json）
sudo tee /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

sudo systemctl restart docker
```

### 6.3 健康检查

```bash
# Cloud API 健康检查
curl http://localhost:8080/health

# Chat H5 检查
curl -I http://localhost:3001
```

## 7. 安全建议

### 7.1 防火墙配置

```bash
# 只开放必要端口
sudo firewall-cmd --permanent --add-port=8080/tcp  # Cloud API
sudo firewall-cmd --permanent --add-port=3001/tcp  # Chat H5
sudo firewall-cmd --reload
```

### 7.2 非 root 用户运行 Docker

```bash
# 创建 docker 用户组（通常安装 Docker 时已创建）
sudo usermod -aG docker $USER

# 使用非 root 用户管理容器
docker-compose down
```

## 8. 故障排查

### 8.1 常见问题

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| 容器启动失败 | 端口被占用 | 检查端口占用 `netstat -tlnp` |
| WebSocket 连接失败 | Nginx 未配置 WebSocket 代理 | 检查 nginx.conf 中的 Upgrade 配置 |
| 前端无法连接 API | API 地址配置错误 | 检查前端环境变量 NEXT_PUBLIC_API_URL |
| 内存不足 | 容器内存占用过高 | 增加服务器内存或限制容器内存 |

### 8.2 调试命令

```bash
# 进入容器
docker exec -it agentapi-cloud sh
docker exec -it agentapi-chat sh

# 查看资源使用
docker stats

# 查看网络连接
docker network ls
docker network inspect agentapi_default
```

## 9. 目录结构

部署后服务器目录结构：

```
/opt/agentapi/
├── docker-compose.yml
├── nginx.conf
├── cloud/
└── chat/
    └── (可选，前端资源)
```

## 10. 注意事项

1. **端口冲突**: 确保 80、8080、3001 端口未被占用
2. **Supabase 配置**: 生产环境建议配置 Supabase 以实现设备状态持久化
3. **备份**: 定期备份配置文件
4. **监控**: 建议配置基本的监控告警
