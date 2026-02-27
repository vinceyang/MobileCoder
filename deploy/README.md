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
