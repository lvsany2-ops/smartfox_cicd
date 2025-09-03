#!/bin/bash
set -e

echo "=== Self-Hosted Runner 环境验证脚本 ==="
echo "此脚本检查运行 SmartFox CI/CD 流水线所需的所有依赖项"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

check_command() {
    local cmd=$1
    local name=$2
    local install_hint=$3
    
    if command -v "$cmd" &> /dev/null; then
        echo -e "${GREEN}✅ $name 已安装${NC}"
        if [[ "$cmd" == "docker" ]]; then
            echo "   版本: $(docker --version)"
        elif [[ "$cmd" == "go" ]]; then
            echo "   版本: $(go version)"
        elif [[ "$cmd" == "node" ]]; then
            echo "   版本: $(node --version)"
        elif [[ "$cmd" == "python3" ]]; then
            echo "   版本: $(python3 --version)"
        elif [[ "$cmd" == "kubectl" ]]; then
            echo "   版本: $(kubectl version --client --short 2>/dev/null || echo "客户端版本检查失败")"
        elif [[ "$cmd" == "kind" ]]; then
            echo "   版本: $(kind --version)"
        elif [[ "$cmd" == "jq" ]]; then
            echo "   版本: $(jq --version)"
        elif [[ "$cmd" == "yq" ]]; then
            echo "   版本: $(yq --version 2>/dev/null || echo "yq 版本检查失败")"
        fi
        return 0
    else
        echo -e "${RED}❌ $name 未安装${NC}"
        if [[ -n "$install_hint" ]]; then
            echo -e "${YELLOW}   安装提示: $install_hint${NC}"
        fi
        return 1
    fi
}

echo "检查必需的软件依赖..."
echo ""

# 检查 Docker
check_command "docker" "Docker" "macOS: brew install docker, Linux: sudo apt-get install docker.io"

# 检查 Docker Compose
if command -v docker-compose &> /dev/null; then
    echo -e "${GREEN}✅ Docker Compose 已安装${NC}"
    echo "   版本: $(docker-compose --version)"
elif docker compose version &> /dev/null; then
    echo -e "${GREEN}✅ Docker Compose (v2) 已安装${NC}"
    echo "   版本: $(docker compose version)"
else
    echo -e "${YELLOW}⚠️  Docker Compose 未安装 (可选)${NC}"
fi

# 检查 Go
check_command "go" "Go" "macOS: brew install go, Linux: sudo apt-get install golang-go"

# 检查 Node.js
check_command "node" "Node.js" "macOS: brew install node, Linux: curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -"

# 检查 npm
check_command "npm" "npm" "通常与 Node.js 一起安装"

# 检查 Python3
check_command "python3" "Python3" "macOS: brew install python3, Linux: sudo apt-get install python3"

# 检查 pip3
check_command "pip3" "pip3" "macOS: brew install python3, Linux: sudo apt-get install python3-pip"

# 检查 kubectl
check_command "kubectl" "kubectl" "macOS: brew install kubectl, Linux: 参考 Kubernetes 官方文档"

# 检查 kind
check_command "kind" "kind" "macOS: brew install kind, Linux: 下载二进制文件"

# 检查 jq
check_command "jq" "jq" "macOS: brew install jq, Linux: sudo apt-get install jq"

# 检查 yq
check_command "yq" "yq" "macOS: brew install yq, Linux: sudo snap install yq"

echo ""
echo "检查 Docker 服务状态..."
if docker info &> /dev/null; then
    echo -e "${GREEN}✅ Docker 服务正在运行${NC}"
else
    echo -e "${RED}❌ Docker 服务未运行${NC}"
    echo -e "${YELLOW}   请启动 Docker 服务${NC}"
fi

echo ""
echo "检查 Docker 权限..."
if docker ps &> /dev/null; then
    echo -e "${GREEN}✅ Docker 权限正常${NC}"
else
    echo -e "${RED}❌ Docker 权限不足${NC}"
    echo -e "${YELLOW}   运行: sudo usermod -aG docker \$USER && newgrp docker${NC}"
fi

echo ""
echo "检查端口可用性..."
ports=(8080 8081 8082 8083 8084 8085)
for port in "${ports[@]}"; do
    if lsof -Pi :$port -sTCP:LISTEN -t &> /dev/null; then
        echo -e "${YELLOW}⚠️  端口 $port 被占用${NC}"
    else
        echo -e "${GREEN}✅ 端口 $port 可用${NC}"
    fi
done

echo ""
echo "检查磁盘空间..."
available_space=$(df / | awk 'NR==2 {print $4}')
if [[ $available_space -gt 10485760 ]]; then  # 10GB in KB
    echo -e "${GREEN}✅ 磁盘空间充足 ($(df -h / | awk 'NR==2 {print $4}') 可用)${NC}"
else
    echo -e "${YELLOW}⚠️  磁盘空间不足 ($(df -h / | awk 'NR==2 {print $4}') 可用)${NC}"
    echo -e "${YELLOW}   建议至少有 10GB 可用空间${NC}"
fi

echo ""
echo "检查内存..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    total_mem=$(sysctl -n hw.memsize)
    total_mem_gb=$((total_mem / 1024 / 1024 / 1024))
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    total_mem_gb=$(free -g | awk 'NR==2{print $2}')
fi

if [[ $total_mem_gb -ge 8 ]]; then
    echo -e "${GREEN}✅ 内存充足 (${total_mem_gb}GB)${NC}"
else
    echo -e "${YELLOW}⚠️  内存可能不足 (${total_mem_gb}GB)${NC}"
    echo -e "${YELLOW}   建议至少 8GB 内存${NC}"
fi

echo ""
echo "=== 测试 Docker 构建 ==="
echo "创建临时 Dockerfile 进行测试..."
cat > /tmp/test.Dockerfile << 'EOF'
FROM alpine:latest
RUN echo "Docker build test successful"
EOF

if docker build -f /tmp/test.Dockerfile -t runner-test:latest /tmp &> /dev/null; then
    echo -e "${GREEN}✅ Docker 构建测试成功${NC}"
    docker rmi runner-test:latest &> /dev/null
else
    echo -e "${RED}❌ Docker 构建测试失败${NC}"
fi

rm -f /tmp/test.Dockerfile

echo ""
echo "=== 测试 Go 构建 ==="
mkdir -p /tmp/go-test
cd /tmp/go-test
cat > main.go << 'EOF'
package main
import "fmt"
func main() { fmt.Println("Go build test successful") }
EOF

if go mod init test &> /dev/null && go build . &> /dev/null; then
    echo -e "${GREEN}✅ Go 构建测试成功${NC}"
else
    echo -e "${RED}❌ Go 构建测试失败${NC}"
fi

cd - &> /dev/null
rm -rf /tmp/go-test

echo ""
echo "=== 测试 kind 集群 ==="
if kind create cluster --name runner-test &> /dev/null; then
    echo -e "${GREEN}✅ kind 集群创建测试成功${NC}"
    kind delete cluster --name runner-test &> /dev/null
else
    echo -e "${RED}❌ kind 集群创建测试失败${NC}"
fi

echo ""
echo "=== 验证完成 ==="
echo "如果所有检查都通过，你的环境已准备好运行 Self-Hosted Runner"
echo "如果有任何失败项，请参考 SELF_HOSTED_RUNNER_SETUP.md 文档进行安装"
