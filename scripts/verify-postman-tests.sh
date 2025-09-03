#!/bin/bash

# 本地 Postman 测试验证脚本
# 用于在本地环境验证 Postman 测试集合

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

echo -e "${BLUE}🧪 本地 Postman 测试验证${NC}"
echo "============================================"

# 检查依赖
print_step "检查依赖项..."

if ! command -v newman &> /dev/null; then
    print_error "Newman 未安装"
    echo "请运行: npm install -g newman"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    print_error "jq 未安装"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "请运行: brew install jq"
    else
        echo "请安装 jq"
    fi
    exit 1
fi

print_success "依赖项检查通过"

# 检查测试文件
TEST_DIR="./test_json"
print_step "检查测试文件..."

if [ ! -d "$TEST_DIR" ]; then
    print_error "测试目录不存在: $TEST_DIR"
    exit 1
fi

test_files=(
    "base.postman_collection.json"
    "SmartFox-Students-Tests.postman_collection.json"
    "SmartFox-Teachers-Tests-Full-Fixed.postman_collection.json"
    "smartfox-test.postman_environment.json"
)

for file in "${test_files[@]}"; do
    if [ -f "$TEST_DIR/$file" ]; then
        print_success "找到测试文件: $file"
    else
        print_error "缺少测试文件: $file"
        exit 1
    fi
done

# 设置测试服务器 (可选)
BASE_URL=${BASE_URL:-"http://127.0.0.1:8080"}
print_step "使用测试服务器: $BASE_URL"

# 检查服务器是否可达 (可选)
if curl -s --connect-timeout 5 "$BASE_URL/health" >/dev/null 2>&1; then
    print_success "测试服务器可达"
else
    print_warning "测试服务器不可达 ($BASE_URL)"
    print_warning "如果您想运行实际测试，请确保服务器正在运行"
    echo ""
    echo "要启动本地测试环境，请运行:"
    echo "cd mirco_service_fox"
    echo ".github/scripts/kind-deploy-and-test.sh"
    echo ""
fi

# 验证 Newman 集合
print_step "验证 Postman 集合格式..."

for collection in "base.postman_collection.json" "SmartFox-Students-Tests.postman_collection.json" "SmartFox-Teachers-Tests-Full-Fixed.postman_collection.json"; do
    echo ""
    echo "验证: $collection"
    
    # 检查 JSON 格式
    if jq empty "$TEST_DIR/$collection" 2>/dev/null; then
        print_success "$collection JSON 格式有效"
        
        # 检查是否为有效的 Postman 集合
        if jq -e '.info.schema' "$TEST_DIR/$collection" >/dev/null 2>&1; then
            schema=$(jq -r '.info.schema' "$TEST_DIR/$collection")
            print_success "$collection 是有效的 Postman 集合 (Schema: $schema)"
        else
            print_warning "$collection 可能不是标准的 Postman 集合格式"
        fi
        
        # 统计请求数量
        request_count=$(jq '[.. | objects | select(has("request"))] | length' "$TEST_DIR/$collection" 2>/dev/null || echo "0")
        echo "  📊 包含 $request_count 个请求"
        
    else
        print_error "$collection JSON 格式无效"
    fi
done

# 验证环境文件
echo ""
echo "验证环境文件: smartfox-test.postman_environment.json"
if jq empty "$TEST_DIR/smartfox-test.postman_environment.json" 2>/dev/null; then
    print_success "环境文件 JSON 格式有效"
    
    # 显示环境变量
    echo "  📋 环境变量:"
    jq -r '.values[]? | "    - \(.key): \(.value)"' "$TEST_DIR/smartfox-test.postman_environment.json"
else
    print_error "环境文件 JSON 格式无效"
fi

echo ""
echo "============================================"
print_success "测试验证完成！"

echo ""
echo "📋 下一步:"
echo "1. 启动测试环境: cd mirco_service_fox && .github/scripts/kind-deploy-and-test.sh"
echo "2. 运行 Postman 测试: cd mirco_service_fox && .github/scripts/run-newman.sh"
echo "3. 或者直接运行 CI 工作流来执行完整测试"

echo ""
echo "💡 提示:"
echo "- 确保所有服务都已启动并健康运行"
echo "- 检查端口转发是否正常工作"
echo "- 如果测试失败，检查服务日志和网络连接"

echo ""
echo "🔧 手动运行单个测试集合:"
echo "newman run test_json/base.postman_collection.json -e test_json/smartfox-test.postman_environment.json"
echo "newman run test_json/SmartFox-Students-Tests.postman_collection.json -e test_json/smartfox-test.postman_environment.json"
echo "newman run test_json/SmartFox-Teachers-Tests-Full-Fixed.postman_collection.json -e test_json/smartfox-test.postman_environment.json"
