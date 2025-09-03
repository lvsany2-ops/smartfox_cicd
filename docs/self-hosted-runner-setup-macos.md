# macOS Self-Hosted Runner 设置指南

本项目配置为使用 self-hosted runner 运行 GitHub Actions，这样可以更好地控制构建环境和资源。

## macOS Self-Hosted Runner 设置

### 前置要求

在设置 GitHub Actions self-hosted runner 之前，确保您的 macOS 系统已安装以下软件：

#### 1. 必需软件

```bash
# 1. 安装 Homebrew（如果尚未安装）
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# 2. 安装必需的开发工具
brew install go
brew install node
brew install python3
brew install docker
brew install kubectl
brew install kind
brew install jq
brew install yq

# 3. 启动 Docker Desktop
open /Applications/Docker.app
# 等待 Docker 完全启动后再继续

# 4. 验证安装
go version          # 应显示 Go 版本
node --version      # 应显示 Node.js 版本
python3 --version   # 应显示 Python 版本
docker --version    # 应显示 Docker 版本
kubectl version --client  # 应显示 kubectl 版本
kind --version      # 应显示 KinD 版本
```

#### 2. 可选软件（推荐）

```bash
# 安装额外的有用工具
brew install newman     # Postman CLI
brew install helm       # Kubernetes 包管理器
brew install watch      # 监控命令
```

### 设置 GitHub Actions Runner

#### 1. 在 GitHub 仓库中添加 Self-Hosted Runner

1. 转到您的 GitHub 仓库
2. 点击 **Settings** > **Actions** > **Runners**
3. 点击 **New self-hosted runner**
4. 选择 **macOS** 作为操作系统
5. 选择 **x64** 作为架构（如果是 Apple Silicon，选择 **ARM64**）

#### 2. 下载并配置 Runner

按照 GitHub 提供的说明进行操作：

```bash
# 创建一个目录来存放 runner
mkdir ~/actions-runner && cd ~/actions-runner

# 下载最新的 runner 包（URL 来自 GitHub 页面）
curl -o actions-runner-osx-x64-2.311.0.tar.gz -L https://github.com/actions/runner/releases/download/v2.311.0/actions-runner-osx-x64-2.311.0.tar.gz

# 可选：验证哈希值
echo "5f0bb23f5ed6b8b2dd24ac2c96b0b3ff9a1896c0b56a9c1db9cf60e9f79ff96f  actions-runner-osx-x64-2.311.0.tar.gz" | shasum -a 256 -c

# 解压
tar xzf ./actions-runner-osx-x64-2.311.0.tar.gz
```

#### 3. 配置 Runner

```bash
# 配置 runner（使用 GitHub 提供的令牌）
./config.sh --url https://github.com/YOUR_USERNAME/YOUR_REPO --token YOUR_TOKEN

# 安装并启动 runner 服务
sudo ./svc.sh install
sudo ./svc.sh start
```

#### 4. 设置环境变量（可选）

创建 `.env` 文件来设置环境变量：

```bash
# 在 actions-runner 目录中创建 .env 文件
cat > .env << 'EOF'
# Go 环境
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Node.js 环境
export NODE_OPTIONS="--max-old-space-size=4096"

# Python 环境
export PYTHONPATH=/usr/local/lib/python3.11/site-packages

# Docker 环境
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
EOF
```

### 系统配置优化

#### 1. 增加文件描述符限制

```bash
# 编辑 launchd 配置
sudo nano /Library/LaunchDaemons/limit.maxfiles.plist

# 添加以下内容：
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>Label</key>
    <string>limit.maxfiles</string>
    <key>ProgramArguments</key>
    <array>
      <string>launchctl</string>
      <string>limit</string>
      <string>maxfiles</string>
      <string>65536</string>
      <string>65536</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>ServiceIPC</key>
    <false/>
  </dict>
</plist>

# 加载配置
sudo launchctl load -w /Library/LaunchDaemons/limit.maxfiles.plist
```

#### 2. 配置 Docker 资源

在 Docker Desktop 中：
- 内存：至少 8GB（推荐 16GB）
- CPU：至少 4 核
- 磁盘空间：至少 100GB 可用空间

### 故障排除

#### 常见问题

1. **Docker 未启动**
   ```bash
   # 检查 Docker 状态
   docker info
   
   # 如果失败，启动 Docker Desktop
   open /Applications/Docker.app
   ```

2. **端口冲突**
   ```bash
   # 检查端口使用情况
   lsof -i :8080
   lsof -i :3000
   
   # 终止占用端口的进程
   sudo kill -9 <PID>
   ```

3. **权限问题**
   ```bash
   # 确保 runner 用户有适当权限
   sudo chown -R $(whoami) ~/actions-runner
   ```

4. **内存不足**
   ```bash
   # 监控内存使用情况
   top -o MEM
   
   # 清理 Docker 缓存
   docker system prune -a -f
   ```

#### 日志查看

```bash
# Runner 服务日志
tail -f ~/actions-runner/_diag/Runner_*.log

# 作业日志
tail -f ~/actions-runner/_diag/Worker_*.log
```

### 维护和更新

#### 定期维护

```bash
# 清理 Docker 资源
docker system prune -a -f
docker volume prune -f

# 清理 Go 模块缓存
go clean -modcache

# 清理 npm 缓存
npm cache clean --force

# 清理 KinD 集群
kind delete cluster --name smartfox-ci

# 更新 Homebrew 包
brew update && brew upgrade
```

#### 更新 Runner

```bash
# 停止服务
sudo ./svc.sh stop

# 下载新版本并重新配置
# （按照 GitHub 提供的说明）

# 重新启动服务
sudo ./svc.sh start
```

### 安全考虑

1. **网络安全**
   - 确保防火墙正确配置
   - 仅允许必要的出站连接

2. **文件权限**
   - 限制 runner 用户的权限
   - 定期审查文件权限

3. **密钥管理**
   - 使用 GitHub Secrets 存储敏感信息
   - 定期轮换访问令牌

## 验证设置

运行验证脚本来确保环境正确配置：

```bash
chmod +x scripts/verify-runner-env.sh
./scripts/verify-runner-env.sh
```

这个脚本会检查所有必需的依赖项并报告任何问题。

## 性能优化建议

### 1. 资源分配

- **内存**: 为 Docker Desktop 分配至少 8GB RAM
- **CPU**: 使用所有可用核心
- **存储**: 使用 SSD 并保持足够的可用空间

### 2. 缓存优化

```bash
# 设置 Go 模块代理（提高下载速度）
export GOPROXY=https://proxy.golang.org,direct

# 设置 npm 镜像源
npm config set registry https://registry.npmjs.org/

# Docker 构建缓存
export DOCKER_BUILDKIT=1
```

### 3. 并发设置

```bash
# Go 构建并发数
export GOMAXPROCS=$(sysctl -n hw.ncpu)

# Docker 构建并发数
export DOCKER_CLI_EXPERIMENTAL=enabled
```

## Apple Silicon (M1/M2) 特殊说明

如果您使用的是 Apple Silicon Mac：

1. **选择正确的架构**：在设置 runner 时选择 **ARM64**
2. **使用兼容的镜像**：确保 Docker 镜像支持 ARM64 架构
3. **Rosetta 2**：某些工具可能需要 Rosetta 2 兼容层

```bash
# 安装 Rosetta 2（如果需要）
sudo softwareupdate --install-rosetta

# 检查架构
uname -m  # 应显示 arm64
```

## 监控和警报

设置监控来跟踪 runner 的健康状态：

```bash
# 创建监控脚本
cat > ~/monitor-runner.sh << 'EOF'
#!/bin/bash
# 监控 runner 状态
while true; do
    echo "=== $(date) ==="
    echo "Memory usage:"
    top -l 1 | grep "PhysMem"
    echo "Disk usage:"
    df -h /
    echo "Docker status:"
    docker info > /dev/null 2>&1 && echo "Docker: OK" || echo "Docker: ERROR"
    echo "Runner status:"
    ps aux | grep "Runner.Listener" | grep -v grep > /dev/null && echo "Runner: Running" || echo "Runner: Stopped"
    echo ""
    sleep 300  # 每5分钟检查一次
done
EOF

chmod +x ~/monitor-runner.sh
```

通过这个配置，您的 macOS self-hosted runner 应该能够成功运行所有的 CI/CD 流水线。
