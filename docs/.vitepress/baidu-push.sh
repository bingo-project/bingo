#!/bin/bash
# 百度主动推送脚本
# 使用方法：
# 1. 在百度搜索资源平台注册并验证站点 https://ziyuan.baidu.com/
# 2. 获取推送接口和 token（在"数据引入" → "链接提交" → "API提交"）
# 3. 替换下面的 YOUR_TOKEN 为实际 token
# 4. 运行脚本：bash docs/.vitepress/baidu-push.sh

SITE="https://bingoctl.dev"
TOKEN="YOUR_TOKEN"  # 替换为你的百度站长平台 token
URLS_FILE="docs/.vitepress/seo-urls.txt"

# 检查 token 是否已配置
if [ "$TOKEN" = "YOUR_TOKEN" ]; then
    echo "❌ 错误：请先配置百度站长平台 token"
    echo "1. 访问 https://ziyuan.baidu.com/ 注册并验证站点"
    echo "2. 在"数据引入" → "链接提交" → "API提交"获取 token"
    echo "3. 编辑此脚本，将 TOKEN 替换为实际值"
    exit 1
fi

# 检查 URL 文件是否存在
if [ ! -f "$URLS_FILE" ]; then
    echo "❌ 错误：找不到 URL 列表文件 $URLS_FILE"
    exit 1
fi

echo "📤 开始向百度推送 URL..."
echo "站点：$SITE"
echo "URL 数量：$(wc -l < "$URLS_FILE")"
echo ""

# 推送 URL
response=$(curl -H 'Content-Type:text/plain' \
    --data-binary @"$URLS_FILE" \
    "http://data.zz.baidu.com/urls?site=${SITE}&token=${TOKEN}")

echo "📊 推送结果："
echo "$response"
echo ""

# 解析结果
if echo "$response" | grep -q "success"; then
    success_count=$(echo "$response" | grep -o '"success":[0-9]*' | cut -d: -f2)
    echo "✅ 成功推送 $success_count 个 URL"
else
    echo "❌ 推送失败，请检查 token 是否正确"
fi

# 提示后续步骤
echo ""
echo "📝 后续步骤："
echo "1. 登录百度搜索资源平台查看收录状态"
echo "2. 等待 1-7 天后检查收录情况（site:bingoctl.dev）"
echo "3. 每次更新文档后重新运行此脚本"
