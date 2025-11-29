#!/bin/zsh
# Bingo 文档站部署脚本
# 用途：构建文档并部署到生产服务器

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
SERVER_USER="${DEPLOY_USER:-root}"
SERVER_HOST="${DEPLOY_HOST:-your-server-ip}"
SERVER_PATH="/var/www/bingo/docs"
NGINX_CONFIG_PATH="/etc/nginx/sites-available/bingoctl.dev"
BACKUP_DIR="/var/www/backups/bingo-docs"

# 打印带颜色的消息
print_step() {
    echo -e "${BLUE}==>${NC} ${1}"
}

print_success() {
    echo -e "${GREEN}✓${NC} ${1}"
}

print_error() {
    echo -e "${RED}✗${NC} ${1}"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} ${1}"
}

# 检查环境变量
check_env() {
    print_step "检查部署环境变量..."

    if [ "$SERVER_HOST" = "your-server-ip" ]; then
        print_error "请设置服务器地址："
        echo "  export DEPLOY_HOST=your-server-ip"
        echo "  export DEPLOY_USER=your-username"
        exit 1
    fi

    print_success "环境变量配置正确"
}

# 构建文档
build_docs() {
    print_step "开始构建文档..."

    # 检查 node_modules
    if [ ! -d "node_modules" ]; then
        print_warning "node_modules 不存在，正在安装依赖..."
        npm install
    fi

    # 构建文档
    npm run docs:build

    if [ $? -eq 0 ]; then
        print_success "文档构建成功"
    else
        print_error "文档构建失败"
        exit 1
    fi
}

# 验证构建产物
verify_build() {
    print_step "验证构建产物..."

    DIST_DIR="docs/.vitepress/dist"

    # 检查关键文件
    if [ ! -f "$DIST_DIR/index.html" ]; then
        print_error "构建产物不完整：缺少 index.html"
        exit 1
    fi

    if [ ! -f "$DIST_DIR/sitemap.xml" ]; then
        print_error "构建产物不完整：缺少 sitemap.xml"
        exit 1
    fi

    # 检查 clean URLs（不应该有 .html 文件在 guide 目录）
    HTML_COUNT=$(find "$DIST_DIR" -name "*.html" | grep -v "404.html\|index.html" | wc -l)
    print_warning "检测到 $HTML_COUNT 个 HTML 文件（cleanUrls 启用后仍会生成 .html 文件用于服务器端）"

    print_success "构建产物验证通过"
}

# 在服务器上创建备份
create_backup() {
    print_step "创建服务器备份..."

    ssh "${SERVER_USER}@${SERVER_HOST}" "
        sudo mkdir -p ${BACKUP_DIR}
        if [ -d ${SERVER_PATH} ]; then
            BACKUP_NAME=backup-\$(date +%Y%m%d-%H%M%S).tar.gz
            sudo tar -czf ${BACKUP_DIR}/\${BACKUP_NAME} -C ${SERVER_PATH} .
            echo '备份已创建: ${BACKUP_DIR}/\${BACKUP_NAME}'
            # 只保留最近 5 个备份
            sudo ls -t ${BACKUP_DIR}/backup-*.tar.gz | tail -n +6 | xargs -r sudo rm
        fi
    "

    print_success "备份创建成功"
}

# 部署文件到服务器
deploy_files() {
    print_step "部署文件到服务器..."

    # 使用 rsync 同步文件
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='node_modules' \
        docs/.vitepress/dist/ \
        "${SERVER_USER}@${SERVER_HOST}:${SERVER_PATH}/"

    if [ $? -eq 0 ]; then
        print_success "文件部署成功"
    else
        print_error "文件部署失败"
        exit 1
    fi
}

# 更新 Nginx 配置
update_nginx() {
    print_step "更新 Nginx 配置..."

    # 上传新的 nginx 配置
    scp docs/.vitepress/nginx.conf "${SERVER_USER}@${SERVER_HOST}:/tmp/bingoctl.dev.conf"

    # 在服务器上更新配置
    ssh "${SERVER_USER}@${SERVER_HOST}" "
        # 备份旧配置
        if [ -f ${NGINX_CONFIG_PATH} ]; then
            sudo cp ${NGINX_CONFIG_PATH} ${NGINX_CONFIG_PATH}.backup
        fi

        # 复制新配置
        sudo mv /tmp/bingoctl.dev.conf ${NGINX_CONFIG_PATH}

        # 测试配置
        sudo nginx -t
        if [ \$? -eq 0 ]; then
            # 重载 Nginx
            sudo systemctl reload nginx
            echo 'Nginx 配置已更新并重载'
        else
            echo 'Nginx 配置测试失败，回滚到旧配置'
            sudo mv ${NGINX_CONFIG_PATH}.backup ${NGINX_CONFIG_PATH}
            exit 1
        fi
    "

    if [ $? -eq 0 ]; then
        print_success "Nginx 配置更新成功"
    else
        print_error "Nginx 配置更新失败"
        exit 1
    fi
}

# 验证部署
verify_deployment() {
    print_step "验证部署结果..."

    # 测试首页
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoctl.dev/)
    if [ "$HTTP_CODE" = "200" ]; then
        print_success "首页访问正常 (HTTP $HTTP_CODE)"
    else
        print_error "首页访问异常 (HTTP $HTTP_CODE)"
    fi

    # 测试 clean URL
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoctl.dev/guide/what-is-bingo)
    if [ "$HTTP_CODE" = "200" ]; then
        print_success "Clean URL 访问正常 (HTTP $HTTP_CODE)"
    else
        print_warning "Clean URL 访问异常 (HTTP $HTTP_CODE)"
    fi

    # 测试重定向
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoctl.dev/guide/what-is-bingo.html)
    if [ "$HTTP_CODE" = "301" ]; then
        print_success "重定向规则正常 (.html -> clean URL)"
    else
        print_warning "重定向规则可能有问题 (HTTP $HTTP_CODE，期望 301)"
    fi

    # 测试 sitemap
    HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" https://bingoctl.dev/sitemap.xml)
    if [ "$HTTP_CODE" = "200" ]; then
        print_success "Sitemap 访问正常 (HTTP $HTTP_CODE)"
    else
        print_error "Sitemap 访问异常 (HTTP $HTTP_CODE)"
    fi
}

# 生成部署报告
generate_report() {
    print_step "生成部署报告..."

    echo ""
    echo "======================================"
    echo "📊 部署完成报告"
    echo "======================================"
    echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "服务器: ${SERVER_HOST}"
    echo "路径: ${SERVER_PATH}"
    echo ""
    echo "🔗 验证链接："
    echo "  - 首页: https://bingoctl.dev/"
    echo "  - Clean URL: https://bingoctl.dev/guide/what-is-bingo"
    echo "  - Sitemap: https://bingoctl.dev/sitemap.xml"
    echo ""
    echo "📝 下一步操作："
    echo "  1. 访问以上链接验证部署"
    echo "  2. 在 Google Search Console 重新提交 sitemap"
    echo "  3. 等待 1-2 天观察 Google 索引更新"
    echo "======================================"
}

# 主流程
main() {
    echo ""
    echo "======================================"
    echo "🚀 Bingo 文档部署脚本"
    echo "======================================"
    echo ""

    # 检查是否在正确的目录
    if [ ! -f "docs/.vitepress/config.mts" ]; then
        print_error "请在项目根目录运行此脚本"
        exit 1
    fi

    # 执行部署流程
    check_env
    build_docs
    verify_build
    create_backup
    deploy_files
    # update_nginx
    verify_deployment
    generate_report

    echo ""
    print_success "部署完成！"
    echo ""
}

# 处理中断信号
trap 'print_error "部署被中断"; exit 1' INT TERM

# 运行主流程
main
