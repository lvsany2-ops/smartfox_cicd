#!/bin/bash

# Self-hosted runner 环境验证脚本 (macOS 优化版)
# 检查所有必需的依赖项是否正确安装

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    if [ $2 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "🔍 验证 Self-hosted Runner 环境 (macOS)"
echo "================================================"

# 检查操作系统
echo ""
echo "📍 系统信息:"
echo "操作系统: $(uname -s)"
echo "架构: $(uname -m)"
echo "版本: $(sw_vers -productVersion)"

# 检查基础工具
echo ""
echo "🔧 基础工具:"

# Git
if command -v git >/dev/null 2>&1; then
    print_status "Git: $(git --version)" 0
else
    print_status "Git: 未安装" 1
fi

# Curl
if command -v curl >/dev/null 2>&1; then
    print_status "Curl: $(curl --version | head -n1)" 0
else
    print_status "Curl: 未安装" 1
fi

# Homebrew (macOS)
if command -v brew >/dev/null 2>&1; then
    print_status "Homebrew: $(brew --version | head -n1)" 0
else
    print_warning "Homebrew: 建议安装以便管理依赖"
fi

# 检查开发工具
echo ""
echo "💻 开发工具:"

# Go
if command -v go >/dev/null 2>&1; then
    go_version=$(go version)
    print_status "Go: $go_version" 0
    
    # 检查 GOPATH
    if [ -n "$GOPATH" ]; then
        print_status "GOPATH: $GOPATH" 0
    else
        print_warning "GOPATH: 未设置 (使用默认值)"
    fi
else
    print_status "Go: 未安装" 1
fi

# Node.js
if command -v node >/dev/null 2>&1; then
    node_version=$(node --version)
    npm_version=$(npm --version)
    print_status "Node.js: $node_version" 0
    print_status "npm: $npm_version" 0
else
    print_status "Node.js: 未安装" 1
fi

# Python
if command -v python3 >/dev/null 2>&1; then
    python_version=$(python3 --version)
    print_status "Python3: $python_version" 0
    
    if command -v pip3 >/dev/null 2>&1; then
        pip_version=$(pip3 --version)
        print_status "pip3: $pip_version" 0
    else
        print_status "pip3: 未安装" 1
    fi
else
    print_status "Python3: 未安装" 1
fi

# 检查容器化工具
echo ""
echo "🐳 容器化工具:"

# Docker
if command -v docker >/dev/null 2>&1; then
    docker_version=$(docker --version)
    print_status "Docker: $docker_version" 0
    
    # 检查 Docker 是否运行
    if docker info >/dev/null 2>&1; then
        print_status "Docker 守护进程: 运行中" 0
    else
        print_status "Docker 守护进程: 未运行" 1
        print_warning "请启动 Docker Desktop"
    fi
    
    # 检查 Docker Desktop (macOS)
    if [ -d "/Applications/Docker.app" ]; then
        print_status "Docker Desktop: 已安装" 0
    else
        print_warning "Docker Desktop: 未找到，请安装"
    fi
else
    print_status "Docker: 未安装" 1
fi

# Docker Compose
if command -v docker-compose >/dev/null 2>&1; then
    compose_version=$(docker-compose --version)
    print_status "Docker Compose: $compose_version" 0
elif docker compose version >/dev/null 2>&1; then
    compose_version=$(docker compose version)
    print_status "Docker Compose (plugin): $compose_version" 0
else
    print_status "Docker Compose: 未安装" 1
fi

# 检查 Kubernetes 工具
echo ""
echo "⚙️ Kubernetes 工具:"

# kubectl
if command -v kubectl >/dev/null 2>&1; then
    kubectl_version=$(kubectl version --client --short 2>/dev/null || kubectl version --client)
    print_status "kubectl: $kubectl_version" 0
else
    print_status "kubectl: 未安装" 1
fi

# KinD
if command -v kind >/dev/null 2>&1; then
    kind_version=$(kind --version)
    print_status "KinD: $kind_version" 0
else
    print_status "KinD: 未安装" 1
fi

# Helm (可选)
if command -v helm >/dev/null 2>&1; then
    helm_version=$(helm version --short)
    print_status "Helm: $helm_version" 0
else
    print_warning "Helm: 未安装 (可选)"
fi

# 检查其他工具
echo ""
echo "🛠️ 其他工具:"

# jq
if command -v jq >/dev/null 2>&1; then
    jq_version=$(jq --version)
    print_status "jq: $jq_version" 0
else
    print_status "jq: 未安装" 1
fi

# yq
if command -v yq >/dev/null 2>&1; then
    yq_version=$(yq --version)
    print_status "yq: $yq_version" 0
else
    print_warning "yq: 未安装 (工作流会自动安装)"
fi

# Newman (可选)
if command -v newman >/dev/null 2>&1; then
    newman_version=$(newman --version)
    print_status "Newman: $newman_version" 0
else
    print_warning "Newman: 未安装 (可选，用于 Postman 测试)"
fi

# 检查系统资源
echo ""
echo "💾 系统资源:"

# 内存
total_memory=$(sysctl -n hw.memsize)
total_memory_gb=$((total_memory / 1024 / 1024 / 1024))
if [ $total_memory_gb -ge 8 ]; then
    print_status "总内存: ${total_memory_gb}GB" 0
else
    print_status "总内存: ${total_memory_gb}GB (建议至少 8GB)" 1
fi

# CPU 核心数
cpu_cores=$(sysctl -n hw.ncpu)
if [ $cpu_cores -ge 4 ]; then
    print_status "CPU 核心: $cpu_cores" 0
else
    print_status "CPU 核心: $cpu_cores (建议至少 4 核)" 1
fi

# 磁盘空间
available_space=$(df -h / | awk 'NR==2 {print $4}')
print_status "可用磁盘空间: $available_space" 0

# 检查网络连接
echo ""
echo "🌐 网络连接:"

# GitHub 连接
if curl -s --connect-timeout 5 https://github.com >/dev/null; then
    print_status "GitHub 连接: 正常" 0
else
    print_status "GitHub 连接: 失败" 1
fi

# Docker Hub 连接
if curl -s --connect-timeout 5 https://hub.docker.com >/dev/null; then
    print_status "Docker Hub 连接: 正常" 0
else
    print_status "Docker Hub 连接: 失败" 1
fi

# Go 代理连接
if curl -s --connect-timeout 5 https://proxy.golang.org >/dev/null; then
    print_status "Go 代理连接: 正常" 0
else
    print_status "Go 代理连接: 失败" 1
fi

# 检查端口可用性
echo ""
echo "🔌 端口检查:"

check_port() {
    local port=$1
    if lsof -i :$port >/dev/null 2>&1; then
        print_status "端口 $port: 被占用" 1
        lsof -i :$port | head -n 2
    else
        print_status "端口 $port: 可用" 0
    fi
}

# 常用端口
for port in 8080 8081 8082 8083 8084 8085 3000 3306; do
    check_port $port
done

# 环境变量检查
echo ""
echo "🔧 环境变量:"

# GOPATH
if [ -n "$GOPATH" ]; then
    print_status "GOPATH: $GOPATH" 0
else
    print_warning "GOPATH: 未设置"
fi

# PATH 中的 Go
if echo $PATH | grep -q "go/bin"; then
    print_status "Go bin in PATH: 是" 0
else
    print_warning "Go bin in PATH: 否"
fi

# Docker 环境变量
if [ -n "$DOCKER_BUILDKIT" ]; then
    print_status "DOCKER_BUILDKIT: $DOCKER_BUILDKIT" 0
else
    print_warning "DOCKER_BUILDKIT: 未设置 (建议设置为 1)"
fi

# 生成报告
echo ""
echo "📊 总结报告:"
echo "================================================"

# 计算错误数
errors=0

# 重新检查关键依赖
critical_deps=("git" "docker" "go" "node" "python3" "kubectl" "kind" "jq")
for dep in "${critical_deps[@]}"; do
    if ! command -v $dep >/dev/null 2>&1; then
        errors=$((errors + 1))
    fi
done

# Docker 运行检查
if ! docker info >/dev/null 2>&1; then
    errors=$((errors + 1))
fi

if [ $errors -eq 0 ]; then
    echo -e "${GREEN}✅ 环境验证通过！所有必需的依赖都已正确安装。${NC}"
    echo ""
    echo "🚀 下一步:"
    echo "1. 按照 docs/self-hosted-runner-setup-macos.md 设置 GitHub Actions Runner"
    echo "2. 配置必要的 GitHub Secrets"
    echo "3. 运行测试工作流"
else
    echo -e "${RED}❌ 发现 $errors 个问题。请安装缺失的依赖项。${NC}"
    echo ""
    echo "📚 安装指南:"
    echo "查看 docs/self-hosted-runner-setup-macos.md 获取详细安装说明"
    echo ""
    echo "快速修复 (macOS):"
    echo "brew install go node python3 docker kubectl kind jq"
    exit 1
fi

echo ""
echo "📚 更多信息:"
echo "- 设置指南: docs/self-hosted-runner-setup-macos.md"
echo "- 故障排除: 查看设置指南中的故障排除部分"
echo "- 性能优化: 确保 Docker Desktop 分配足够资源"
