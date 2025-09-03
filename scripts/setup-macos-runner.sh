#!/bin/bash

# macOS Self-hosted Runner 快速设置脚本
# 自动安装所有必需的依赖项

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_step() {
    echo -e "${BLUE}🔄 $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo -e "${BLUE}🚀 macOS Self-hosted Runner 快速设置${NC}"
echo "=================================================="
echo ""

# 检查是否为 macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    print_error "此脚本仅适用于 macOS 系统"
    exit 1
fi

# 检查并安装 Homebrew
print_step "检查 Homebrew..."
if ! command -v brew >/dev/null 2>&1; then
    print_step "安装 Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    
    # 设置 PATH (对于 Apple Silicon Mac)
    if [[ $(uname -m) == "arm64" ]]; then
        echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
        eval "$(/opt/homebrew/bin/brew shellenv)"
    fi
    
    print_success "Homebrew 安装完成"
else
    print_success "Homebrew 已安装"
fi

# 更新 Homebrew
print_step "更新 Homebrew..."
brew update

# 安装必需的工具
print_step "安装开发工具..."

# 定义要安装的包
packages=(
    "git"           # 版本控制
    "go"            # Go 语言
    "node"          # Node.js
    "python3"       # Python 3
    "docker"        # Docker CLI
    "kubectl"       # Kubernetes CLI
    "kind"          # Kubernetes in Docker
    "jq"            # JSON 处理工具
    "yq"            # YAML 处理工具
    "curl"          # HTTP 客户端
    "wget"          # 下载工具
)

# 可选包
optional_packages=(
    "newman"        # Postman CLI (通过 npm 安装)
    "helm"          # Kubernetes 包管理器
    "watch"         # 监控命令
    "tree"          # 目录树显示
)

echo "安装必需的包..."
for package in "${packages[@]}"; do
    if brew list "$package" >/dev/null 2>&1; then
        print_success "$package 已安装"
    else
        print_step "安装 $package..."
        brew install "$package"
        print_success "$package 安装完成"
    fi
done

echo ""
echo "安装可选的包..."
for package in "${optional_packages[@]}"; do
    if [[ "$package" == "newman" ]]; then
        # Newman 通过 npm 安装
        if command -v newman >/dev/null 2>&1; then
            print_success "newman 已安装"
        else
            print_step "安装 newman (通过 npm)..."
            npm install -g newman
            print_success "newman 安装完成"
        fi
    else
        if brew list "$package" >/dev/null 2>&1; then
            print_success "$package 已安装"
        else
            print_step "安装 $package..."
            brew install "$package" || print_warning "$package 安装失败，但这是可选的"
        fi
    fi
done

# 安装 Docker Desktop
print_step "检查 Docker Desktop..."
if [ -d "/Applications/Docker.app" ]; then
    print_success "Docker Desktop 已安装"
else
    print_step "安装 Docker Desktop..."
    brew install --cask docker
    print_success "Docker Desktop 安装完成"
    print_warning "请手动启动 Docker Desktop 应用程序"
fi

# 配置环境变量
print_step "配置环境变量..."

# 检查 shell 类型
SHELL_RC=""
if [[ $SHELL == *"zsh"* ]]; then
    SHELL_RC="$HOME/.zshrc"
elif [[ $SHELL == *"bash"* ]]; then
    SHELL_RC="$HOME/.bash_profile"
fi

if [ -n "$SHELL_RC" ]; then
    print_step "配置 $SHELL_RC..."
    
    # 备份现有配置
    if [ -f "$SHELL_RC" ]; then
        cp "$SHELL_RC" "$SHELL_RC.backup.$(date +%Y%m%d_%H%M%S)"
    fi
    
    # 添加环境变量配置
    cat >> "$SHELL_RC" << 'EOF'

# GitHub Actions Self-hosted Runner 环境配置
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Node.js 配置
export NODE_OPTIONS="--max-old-space-size=4096"

# Docker 配置
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Go 模块代理 (提高下载速度)
export GOPROXY=https://proxy.golang.org,direct

EOF
    
    print_success "环境变量配置完成"
    print_warning "请运行 'source $SHELL_RC' 或重新启动终端来加载新的环境变量"
fi

# 创建工作目录
print_step "创建工作目录..."
mkdir -p "$HOME/go/bin"
mkdir -p "$HOME/go/src"
mkdir -p "$HOME/go/pkg"
print_success "Go 工作目录创建完成"

# 验证安装
echo ""
print_step "验证安装..."

# 检查关键命令
commands_to_check=(
    "git --version"
    "go version"
    "node --version"
    "npm --version"
    "python3 --version"
    "pip3 --version"
    "docker --version"
    "kubectl version --client"
    "kind --version"
    "jq --version"
)

all_good=true
for cmd in "${commands_to_check[@]}"; do
    if eval "$cmd" >/dev/null 2>&1; then
        print_success "$cmd"
    else
        print_error "$cmd 失败"
        all_good=false
    fi
done

echo ""
echo "=================================================="

if $all_good; then
    print_success "所有工具安装成功！"
    echo ""
    echo "🎉 恭喜！您的 macOS self-hosted runner 环境已准备就绪。"
    echo ""
    echo "📋 下一步:"
    echo "1. 重启终端或运行: source $SHELL_RC"
    echo "2. 启动 Docker Desktop 应用程序"
    echo "3. 运行验证脚本: ./scripts/verify-runner-env-macos.sh"
    echo "4. 按照 docs/self-hosted-runner-setup-macos.md 设置 GitHub Actions Runner"
    echo ""
    echo "💡 提示:"
    echo "- 确保 Docker Desktop 已启动并运行"
    echo "- 为 Docker Desktop 分配至少 8GB 内存"
    echo "- 检查防火墙设置，确保必要的网络连接"
else
    print_error "安装过程中出现了一些问题"
    echo ""
    echo "🛠️  故障排除:"
    echo "1. 检查网络连接"
    echo "2. 重新运行此脚本"
    echo "3. 手动安装失败的工具"
    echo "4. 查看 docs/self-hosted-runner-setup-macos.md 获取帮助"
fi

echo ""
echo "📚 有用的命令:"
echo "- 更新所有工具: brew update && brew upgrade"
echo "- 检查 Docker: docker info"
echo "- 验证环境: ./scripts/verify-runner-env-macos.sh"
echo "- 清理缓存: docker system prune -a"

echo ""
print_step "设置完成！"
