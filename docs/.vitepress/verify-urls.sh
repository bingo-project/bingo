#!/bin/bash
# URL 验证脚本
# 用途：验证 Clean URLs 和重定向规则是否正常工作

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试结果统计
PASSED=0
FAILED=0
TOTAL=0

# 打印函数
print_test() {
    echo -e "${BLUE}测试${NC} $1"
}

print_pass() {
    echo -e "${GREEN}✓ PASS${NC} $1"
    ((PASSED++))
    ((TOTAL++))
}

print_fail() {
    echo -e "${RED}✗ FAIL${NC} $1"
    ((FAILED++))
    ((TOTAL++))
}

print_info() {
    echo -e "${YELLOW}ℹ INFO${NC} $1"
}

# 测试 URL 函数
test_url() {
    local url=$1
    local expected_code=$2
    local description=$3

    print_test "$description"
    print_info "URL: $url"

    response=$(curl -s -o /dev/null -w "%{http_code}|%{redirect_url}" "$url")
    http_code=$(echo "$response" | cut -d'|' -f1)
    redirect_url=$(echo "$response" | cut -d'|' -f2)

    if [ "$http_code" = "$expected_code" ]; then
        if [ -n "$redirect_url" ]; then
            print_pass "HTTP $http_code (重定向到: $redirect_url)"
        else
            print_pass "HTTP $http_code"
        fi
    else
        print_fail "期望 HTTP $expected_code，实际 HTTP $http_code"
        if [ -n "$redirect_url" ]; then
            print_info "重定向到: $redirect_url"
        fi
    fi
    echo ""
}

echo ""
echo "======================================"
echo "🔍 Bingo 文档站 URL 验证"
echo "======================================"
echo ""

# 1. 测试首页
echo "【1】测试首页"
echo "---"
test_url "https://bingoctl.dev/" "200" "首页（中文）"
test_url "https://bingoctl.dev/en/" "200" "首页（英文）"

# 2. 测试 Clean URLs（应该返回 200）
echo "【2】测试 Clean URLs（新格式）"
echo "---"
test_url "https://bingoctl.dev/guide/what-is-bingo" "200" "什么是 Bingo（clean URL）"
test_url "https://bingoctl.dev/guide/getting-started" "200" "快速开始（clean URL）"
test_url "https://bingoctl.dev/essentials/architecture" "200" "整体架构（clean URL）"
test_url "https://bingoctl.dev/en/guide/what-is-bingo" "200" "What is Bingo（clean URL）"

# 3. 测试 .html 重定向（应该 301 到 clean URL）
echo "【3】测试 .html 重定向"
echo "---"
test_url "https://bingoctl.dev/guide/what-is-bingo.html" "301" ".html 重定向到 clean URL"
test_url "https://bingoctl.dev/guide/getting-started.html" "301" ".html 重定向到 clean URL"
test_url "https://bingoctl.dev/en/guide/what-is-bingo.html" "301" ".html 重定向到 clean URL（英文）"

# 4. 测试 .html/ 重定向（应该 301 到 clean URL）
echo "【4】测试 .html/ 重定向（修复 Google 索引问题）"
echo "---"
test_url "https://bingoctl.dev/guide/what-is-bingo.html/" "301" ".html/ 重定向到 clean URL"
test_url "https://bingoctl.dev/en/guide/what-is-bingo.html/" "301" ".html/ 重定向到 clean URL（英文）"

# 5. 测试 sitemap 和 robots.txt
echo "【5】测试 SEO 文件"
echo "---"
test_url "https://bingoctl.dev/sitemap.xml" "200" "Sitemap"
test_url "https://bingoctl.dev/robots.txt" "200" "Robots.txt"

# 6. 测试 404
echo "【6】测试 404 处理"
echo "---"
test_url "https://bingoctl.dev/not-exist-page" "404" "不存在的页面应返回 404"

# 打印测试报告
echo "======================================"
echo "📊 测试报告"
echo "======================================"
echo "总计: $TOTAL 个测试"
echo -e "${GREEN}通过: $PASSED${NC}"
echo -e "${RED}失败: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    echo ""
    echo "✅ Clean URLs 配置正确"
    echo "✅ 重定向规则工作正常"
    echo "✅ Google 索引问题已修复"
    echo ""
    exit 0
else
    echo -e "${RED}✗ 有 $FAILED 个测试失败${NC}"
    echo ""
    echo "❌ 请检查服务器配置"
    echo ""
    exit 1
fi
