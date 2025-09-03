# Self-Hosted Runner 设置指南

本文档描述如何在本地机器上设置 GitHub Actions Self-Hosted Runner 来运行 SmartFox CI/CD 流水线。

## 前置要求

### 必需软件
在设置 Self-Hosted Runner 之前，请确保您的机器上已安装以下软件：

#### 1. Docker
```bash
# macOS
brew install docker
# 或者下载 Docker Desktop

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install docker.io docker-compose
sudo usermod -aG docker $USER

# 验证安装
docker --version
```

#### 2. Go (版本 1.21+)
```bash
# macOS
brew install go

# Ubuntu/Debian
sudo apt-get install golang-go

# 验证安装
go version
```

#### 3. Node.js (版本 18+)
```bash
# macOS
brew install node

# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# 验证安装
node --version
npm --version
```

#### 4. Python (版本 3.8+)
```bash
# macOS
brew install python3

# Ubuntu/Debian
sudo apt-get install python3 python3-pip

# 验证安装
python3 --version
pip3 --version
```

#### 5. Kubernetes 工具
```bash
# kubectl
# macOS
brew install kubectl

# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl
curl -fsSL https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-archive-keyring.gpg
echo "deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
sudo apt-get update
sudo apt-get install -y kubectl

# kind (Kubernetes in Docker)
# macOS
brew install kind

# Linux
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# 验证安装
kubectl version --client
kind --version
```

#### 6. 工具软件
```bash
# jq (JSON 处理工具)
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# yq (YAML 处理工具)
# macOS
brew install yq

# Ubuntu/Debian
sudo snap install yq
```

## GitHub Self-Hosted Runner 设置

### 1. 在 GitHub 仓库中添加 Runner

1. 进入你的 GitHub 仓库
2. 点击 "Settings" → "Actions" → "Runners"
3. 点击 "New self-hosted runner"
4. 选择你的操作系统（macOS/Linux）
5. 按照页面上的指令下载并配置 runner

### 2. 配置 Runner

```bash
# 下载
mkdir actions-runner && cd actions-runner
curl -o actions-runner-linux-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-linux-x64-2.311.0.tar.gz
tar xzf ./actions-runner-linux-x64-2.311.0.tar.gz

# 配置
./config.sh --url https://github.com/lvsany2-ops/smartfox_cicd --token YOUR_TOKEN

# 启动 runner
./run.sh
```

### 3. 设置为系统服务（推荐）

```bash
# 安装服务
sudo ./svc.sh install

# 启动服务
sudo ./svc.sh start

# 检查状态
sudo ./svc.sh status
```

## 环境变量配置

在 runner 机器上设置以下环境变量（可选）：

```bash
# ~/.bashrc 或 ~/.zshrc
export GOPROXY=https://proxy.golang.org,direct
export GO111MODULE=on
export DOCKER_BUILDKIT=1

# 重新加载配置
source ~/.bashrc  # 或 source ~/.zshrc
```

## 验证设置

运行以下命令验证所有依赖项都已正确安装：

```bash
#!/bin/bash
echo "=== 验证 Self-Hosted Runner 环境 ==="

echo "检查 Docker..."
docker --version || echo "❌ Docker 未安装"

echo "检查 Go..."
go version || echo "❌ Go 未安装"

echo "检查 Node.js..."
node --version || echo "❌ Node.js 未安装"
npm --version || echo "❌ npm 未安装"

echo "检查 Python..."
python3 --version || echo "❌ Python3 未安装"
pip3 --version || echo "❌ pip3 未安装"

echo "检查 kubectl..."
kubectl version --client || echo "❌ kubectl 未安装"

echo "检查 kind..."
kind --version || echo "❌ kind 未安装"

echo "检查 jq..."
jq --version || echo "❌ jq 未安装"

echo "检查 yq..."
yq --version || echo "❌ yq 未安装"

echo "=== 验证完成 ==="
```

## 故障排除

### 常见问题

1. **Docker 权限问题**
   ```bash
   # 将用户添加到 docker 组
   sudo usermod -aG docker $USER
   # 重新登录或重启终端
   ```

2. **端口冲突**
   ```bash
   # 检查端口使用情况
   lsof -i :8080-8085
   # 或者
   netstat -tulpn | grep :808
   ```

3. **磁盘空间不足**
   ```bash
   # 清理 Docker 镜像和容器
   docker system prune -af
   
   # 清理 kind 集群
   kind delete cluster --name smartfox-ci
   ```

4. **内存不足**
   - 确保机器有至少 8GB RAM
   - 监控资源使用：`htop` 或 `top`

### 日志查看

```bash
# Runner 服务日志
sudo journalctl -u actions.runner.* -f

# Docker 日志
docker logs <container_name>

# Kubernetes Pod 日志
kubectl logs <pod_name> -n default
```

## 安全考虑

1. **隔离环境**：建议在虚拟机或容器中运行 self-hosted runner
2. **定期更新**：保持 runner 和依赖软件的最新版本
3. **监控**：监控 runner 的资源使用和日志
4. **备份**：定期备份 runner 配置

## 性能优化

1. **使用 SSD**：提高 I/O 性能
2. **增加内存**：至少 8GB，推荐 16GB
3. **网络**：确保稳定的网络连接
4. **并行构建**：利用多核 CPU 进行并行构建

```bash
# 设置 Go 并行构建
export GOMAXPROCS=$(nproc)

# 设置 Docker 并行构建
export DOCKER_BUILDKIT=1
```

## 监控和维护

定期执行以下维护任务：

```bash
# 清理脚本 (cleanup.sh)
#!/bin/bash
echo "清理过期的 Docker 镜像..."
docker image prune -af --filter "until=24h"

echo "清理过期的 Docker 卷..."
docker volume prune -f

echo "清理 kind 集群..."
kind delete cluster --name smartfox-ci || true

echo "清理 Go 模块缓存..."
go clean -modcache

echo "清理完成"
```

建议每天运行一次清理脚本，可以设置 cron job：

```bash
# 添加到 crontab
0 2 * * * /path/to/cleanup.sh >> /var/log/runner-cleanup.log 2>&1
```
