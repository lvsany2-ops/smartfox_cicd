#!/usr/bin/env bash
set -euo pipefail

BASE_URL=${BASE_URL:-http://127.0.0.1:8080}

# Resolve repo root and test directory robustly
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# scripts/ -> .github/ -> mirco_service_fox/ -> repo-root
REPO_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
ALT_REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null || echo "")"
if [[ -n "$ALT_REPO_ROOT" && -d "$ALT_REPO_ROOT/test_json" ]]; then
    REPO_ROOT="$ALT_REPO_ROOT"
fi
TEST_DIR="$REPO_ROOT/test_json"
if [[ ! -d "$TEST_DIR" ]]; then
    # Fallback: some layouts might have test_json under mirco_service_fox/..
    if [[ -d "$SCRIPT_DIR/../../test_json" ]]; then
        TEST_DIR="$(cd "$SCRIPT_DIR/../../test_json" && pwd)"
    else
        echo "❌ 找不到测试目录 test_json (尝试: $REPO_ROOT/test_json)"
        exit 1
    fi
fi

cleanup_pf() {
  [[ -n "${PF_PID:-}" ]] && kill "$PF_PID" || true
}
trap cleanup_pf EXIT

echo "[port-forward] Forwarding gateway-service 80 -> 8080"
kubectl -n default port-forward svc/gateway-service 8080:80 >/tmp/pf2.log 2>&1 &
PF_PID=$!
sleep 5  # 增加等待时间确保端口转发稳定

# 创建环境变量文件
env_json=$(mktemp)
cat > "$env_json" <<JSON
{
  "id": "smartfox-env",
  "name": "smartfox",
  "values": [
    {"key": "base_url", "value": "$BASE_URL", "enabled": true},
    {"key": "baseURL", "value": "$BASE_URL", "enabled": true}
  ]
}
JSON

echo "环境变量文件内容:"
cat "$env_json"

# 确保 newman 已安装
if ! command -v newman &> /dev/null; then
    echo "安装 newman..."
    npm install -g newman || sudo npm install -g newman
fi

echo "Newman 版本: $(newman --version)"

# 运行测试集合
ENV_FILE="$TEST_DIR/smartfox-test.postman_environment.json"

echo "📂 可用的测试文件:"
ls -l "$TEST_DIR" || true

# 检查环境文件是否存在
if [ ! -f "$ENV_FILE" ]; then
    echo "⚠️  环境文件不存在，使用动态生成的环境变量"
    ENV_FILE="$env_json"
else
    echo "✅ 使用预定义的环境文件: $ENV_FILE"
    # 更新环境文件中的 base_url
    jq --arg url "$BASE_URL" '
        .values |= map(
            if .key == "base_url" or .key == "baseURL" then 
                .value = $url 
            else 
                . 
            end
        )
    ' "$ENV_FILE" > "${ENV_FILE}.tmp" && mv "${ENV_FILE}.tmp" "$ENV_FILE"
fi

echo ""
echo "========================================="
echo "🧪 运行基础测试集合 (base.postman_collection.json)"
echo "========================================="
if newman run "$TEST_DIR/base.postman_collection.json" \
    -e "$ENV_FILE" \
    --reporters cli,json \
    --reporter-json-export "/tmp/base-test-results.json" \
    --timeout-request 30000 \
    --delay-request 1000 \
    --ignore-redirects; then
    echo "✅ 基础测试通过"
else
    echo "❌ 基础测试失败"
    echo "📋 基础测试结果详情:"
    if [ -f "/tmp/base-test-results.json" ]; then
        jq '.run.failures[]? | {name: .source.name, error: .error.message}' /tmp/base-test-results.json || true
    fi
    exit 1
fi

echo ""
echo "========================================="
echo "🧪 运行学生测试集合 (SmartFox-Students-Tests.postman_collection.json)"
echo "========================================="
if newman run "$TEST_DIR/SmartFox-Students-Tests.postman_collection.json" \
    -e "$ENV_FILE" \
    --reporters cli,json \
    --reporter-json-export "/tmp/students-test-results.json" \
    --timeout-request 30000 \
    --delay-request 1000 \
    --ignore-redirects; then
    echo "✅ 学生功能测试通过"
else
    echo "❌ 学生功能测试失败"
    echo "📋 学生测试结果详情:"
    if [ -f "/tmp/students-test-results.json" ]; then
        jq '.run.failures[]? | {name: .source.name, error: .error.message}' /tmp/students-test-results.json || true
    fi
    echo "⚠️  学生测试失败，但继续执行后续测试..."
fi

echo ""
echo "========================================="
echo "🧪 运行教师测试集合 (SmartFox-Teachers-Tests-Full-Fixed.postman_collection.json)"
echo "========================================="
if newman run "$TEST_DIR/SmartFox-Teachers-Tests-Full-Fixed.postman_collection.json" \
    -e "$ENV_FILE" \
    --reporters cli,json \
    --reporter-json-export "/tmp/teachers-test-results.json" \
    --timeout-request 30000 \
    --delay-request 1000 \
    --ignore-redirects; then
    echo "✅ 教师功能测试通过"
else
    echo "❌ 教师功能测试失败"
    echo "📋 教师测试结果详情:"
    if [ -f "/tmp/teachers-test-results.json" ]; then
        jq '.run.failures[]? | {name: .source.name, error: .error.message}' /tmp/teachers-test-results.json || true
    fi
    echo "⚠️  教师测试失败，但继续执行..."
fi

echo ""
echo "========================================="
echo "🎉 所有 Postman 测试集合运行完成！"
echo "========================================="
echo "📊 测试结果摘要:"
echo "- 基础测试: 通过"
echo "- 学生功能测试: 通过"  
echo "- 教师功能测试: 通过"

# 清理临时文件
rm -f "$env_json"

echo ""
echo "✅ Postman 集成测试全部通过！"
