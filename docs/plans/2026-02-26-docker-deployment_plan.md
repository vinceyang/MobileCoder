# Docker 部署实施计划

> **For Claude:** 使用 superpowers:subagent-driven-development 逐任务实施

**目标:** 创建完整的 Docker 部署配置，包括 Dockerfile、docker-compose.yml、Nginx 配置和部署脚本

**架构:** 使用 Docker Compose 在单服务器上同时部署 cloud 服务端、chat 前端和 Nginx 反向代理

**技术栈:** Docker, Docker Compose, Nginx, Go, Next.js

---

## 实施任务

### Task 1: 创建部署目录结构

**Files:**
- Create: `deploy/`

**Step 1: 创建 deploy 目录**

```bash
mkdir -p /Users/yangxq/Code/agentapi/deploy
ls -la /Users/yangxq/Code/agentapi/deploy
```

**预期输出:** 目录创建成功

---

### Task 2: 创建 Cloud 服务端 Dockerfile

**Files:**
- Create: `deploy/Dockerfile.cloud`

**Step 1: 写入 Dockerfile.cloud**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY cloud/go.mod cloud/go.sum ./
RUN go mod download

COPY cloud/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
```

---

### Task 3: 创建 Chat 前端 Dockerfile

**Files:**
- Create: `deploy/Dockerfile.chat`

**Step 1: 写入 Dockerfile.chat**

```dockerfile
FROM node:20-alpine AS builder

WORKDIR /app
COPY chat/package.json chat/package-lock.json* ./
RUN npm ci

COPY chat/ .
ENV GITHUB_PAGES=true
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
ENV PORT=3000
CMD ["node", "server.js"]
```

---

### Task 4: 配置 Chat 前端 standalone 模式

**Files:**
- Modify: `chat/next.config.ts`

**Step 1: 检查现有 next.config.ts**

```bash
cat /Users/yangxq/Code/agentapi/chat/next.config.ts
```

**Step 2: 添加 standalone 输出配置**

如果文件不存在或为空，创建包含 standalone 配置的版本：

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  images: {
    unoptimized: true,
  },
};

export default nextConfig;
```

---

### Task 5: 创建 Nginx 配置文件

**Files:**
- Create: `deploy/nginx.conf`

**Step 1: 写入 nginx.conf**

```nginx
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                     '$status $body_bytes_sent "$http_referer" '
                     '"$http_user_agent" "$http_x_forwarded_for"';

    access_log /var/log/nginx/access.log main;
    error_log /var/log/nginx/error.log;

    sendfile on;
    keepalive_timeout 65;

    # Cloud WebSocket + HTTP
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

        # Cloud API (包括 WebSocket)
        location / {
            proxy_pass http://cloud_backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # WebSocket 超时设置
            proxy_read_timeout 86400;
            proxy_send_timeout 86400;
        }

        # Health check
        location /health {
            proxy_pass http://cloud_backend/health;
            proxy_set_header Host $host;
        }

        # Chat H5 前端
        location /chat/ {
            proxy_pass http://chat_backend/;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_http_version 1.1;

            # 禁用缓存确保实时更新
            proxy_cache off;
        }
    }
}
```

---

### Task 6: 创建 docker-compose.yml

**Files:**
- Create: `deploy/docker-compose.yml`

**Step 1: 写入 docker-compose.yml**

```yaml
version: '3.8'

services:
  cloud:
    build:
      context: ..
      dockerfile: deploy/Dockerfile.cloud
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
    networks:
      - agentapi-network

  chat:
    build:
      context: ..
      dockerfile: deploy/Dockerfile.chat
    image: agentapi/chat:latest
    container_name: agentapi-chat
    ports:
      - "3001:3000"
    environment:
      - PORT=3000
      - NODE_ENV=production
    restart: unless-stopped
    networks:
      - agentapi-network

  nginx:
    image: nginx:latest
    container_name: agentapi-nginx
    ports:
      - "80:80"
      - "8080:8080"
      - "3001:3001"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - cloud
      - chat
    restart: unless-stopped
    networks:
      - agentapi-network

networks:
  agentapi-network:
    driver: bridge
```

---

### Task 7: 创建部署脚本

**Files:**
- Create: `deploy/build.sh`
- Create: `deploy/deploy.sh`

**Step 1: 写入 build.sh**

```bash
#!/bin/bash
set -e

echo "=== Building Docker images ==="

# 构建 Cloud 镜像
echo "Building cloud image..."
docker build -t agentapi/cloud:latest -f deploy/Dockerfile.cloud .

# 构建 Chat 镜像
echo "Building chat image..."
docker build -t agentapi/chat:latest -f deploy/Dockerfile.chat .

echo "=== Build complete ==="
docker images | grep agentapi
```

**Step 2: 写入 deploy.sh**

```bash
#!/bin/bash
set -e

SERVER=${1:-}
if [ -z "$SERVER" ]; then
    echo "Usage: ./deploy.sh <user@server>"
    echo "Example: ./deploy.sh root@123.45.67.89"
    exit 1
fi

echo "=== Deploying to $SERVER ==="

# 导出镜像
echo "Exporting images..."
docker save -o /tmp/agentapi-cloud.tar agentapi/cloud:latest
docker save -o /tmp/agentapi-chat.tar agentapi/chat:latest

# 上传到服务器
echo "Uploading to server..."
scp /tmp/agentapi-cloud.tar /tmp/agentapi-chat.tar ${SERVER}:/tmp/

# 上传配置文件
echo "Uploading config files..."
scp deploy/docker-compose.yml ${SERVER}:/opt/agentapi/
scp deploy/nginx.conf ${SERVER}:/opt/agentapi/

# 在服务器上加载镜像并启动
echo "Loading images and starting services..."
ssh $SERVER << 'EOF'
    cd /opt/agentapi

    # 加载镜像
    docker load -i /tmp/agentapi-cloud.tar
    docker load -i /tmp/agentapi-chat.tar

    # 启动服务
    docker-compose up -d

    # 清理临时文件
    rm -f /tmp/agentapi-cloud.tar /tmp/agentapi-chat.tar

    echo "=== Deployment complete ==="
    docker-compose ps
EOF

echo "=== Done ==="
```

**Step 3: 添加执行权限**

```bash
chmod +x /Users/yangxq/Code/agentapi/deploy/build.sh
chmod +x /Users/yangxq/Code/agentapi/deploy/deploy.sh
```

---

### Task 8: 创建服务器初始化脚本

**Files:**
- Create: `deploy/init-server.sh`

**Step 1: 写入 init-server.sh**

```bash
#!/bin/bash
set -e

echo "=== Initializing server for Docker deployment ==="

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# 安装 Docker
echo "Installing Docker..."
curl -fsSL https://get.docker.com | sh

# 启动 Docker
echo "Starting Docker..."
systemctl start docker
systemctl enable docker

# 安装 Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 创建部署目录
echo "Creating deployment directory..."
mkdir -p /opt/agentapi

# 配置防火墙（可选）
echo "Configuring firewall..."
# ufw allow 8080/tcp  # Cloud API
# ufw allow 3001/tcp   # Chat H5

echo "=== Initialization complete ==="
echo "Next steps:"
echo "1. Upload deployment files to /opt/agentapi/"
echo "2. Run: cd /opt/agentapi && docker-compose up -d"
```

**Step 2: 添加执行权限**

```bash
chmod +x /Users/yangxq/Code/agentapi/deploy/init-server.sh
```

---

### Task 9: 创建部署说明文档

**Files:**
- Create: `deploy/README.md`

**Step 1: 写入部署说明**

```markdown
# 部署说明

## 前置要求

- 云服务器（阿里云/腾讯云）
- root 或 sudo 权限

## 快速开始

### 1. 初始化服务器

```bash
# 上传初始化脚本
scp deploy/init-server.sh user@server:/tmp/

# 在服务器上执行
ssh user@server
chmod +x /tmp/init-server.sh
sudo /tmp/init-server.sh
```

### 2. 构建镜像

```bash
cd agentapi
./deploy/build.sh
```

### 3. 部署到服务器

```bash
./deploy/deploy.sh user@server
```

## 服务地址

- Cloud API: http://<server-ip>:8080
- Chat H5: http://<server-ip>:3001

## 常用命令

```bash
# 查看状态
docker-compose -f deploy/docker-compose.yml ps

# 查看日志
docker-compose -f deploy/docker-compose.yml logs -f

# 重启服务
docker-compose -f deploy/docker-compose.yml restart

# 停止服务
docker-compose -f deploy/docker-compose.yml down

# 更新部署
./deploy/build.sh
./deploy/deploy.sh user@server
```

## 防火墙配置

```bash
# 开放端口
sudo ufw allow 8080/tcp
sudo ufw allow 3001/tcp
sudo ufw reload
```
```

---

### Task 10: 更新项目 .gitignore

**Files:**
- Modify: `.gitignore`

**Step 1: 检查现有 .gitignore**

```bash
cat /Users/yangxq/Code/agentapi/.gitignore
```

**Step 2: 添加 deploy 目录（如果需要保留）**

如果希望将部署配置纳入版本控制，添加：
```
# 部署产物
*.tar
*.tar.gz
```

如果不想将 deploy 目录纳入版本控制（推荐），则无需修改。

---

### Task 11: 验证配置

**Step 1: 本地构建测试**

```bash
cd /Users/yangxq/Code/agentapi
docker build -t agentapi/cloud:latest -f deploy/Dockerfile.cloud .
```

**Step 2: 检查构建是否成功**

```bash
docker images | grep agentapi
```

预期输出应显示 agentapi/cloud 和 agentapi/chat 镜像。

---

## 执行选项

**Plan complete and saved to `docs/plans/2026-02-26-docker-deployment-plan.md`. 两个执行选项:**

1. **Subagent-Driven (本会话)** - 我为每个任务派遣一个新的子代理，任务间进行代码审查，快速迭代
2. **Parallel Session (新会话)** - 在新会话中使用 executing-plans，批量执行并设置检查点

**选择哪种方式?**
