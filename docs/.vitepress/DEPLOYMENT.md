# 部署指南

本文档说明如何部署 Bingo 文档站，包括 Clean URLs 配置和自动化部署。

## 快速部署

### 方式一：使用自动化脚本（推荐）

```bash
# 1. 设置服务器信息
export DEPLOY_HOST=your-server-ip
export DEPLOY_USER=your-username

# 2. 运行部署脚本
bash docs/.vitepress/deploy.sh
```

脚本会自动完成：
- ✅ 构建文档
- ✅ 验证构建产物
- ✅ 创建服务器备份
- ✅ 部署文件
- ✅ 更新 Nginx 配置
- ✅ 验证部署结果

### 方式二：手动部署

```bash
# 1. 构建文档
npm run docs:build

# 2. 上传到服务器
rsync -avz --delete docs/.vitepress/dist/ user@server:/var/www/bingo/docs/

# 3. 更新 Nginx 配置
scp docs/.vitepress/nginx.conf user@server:/tmp/
ssh user@server "sudo mv /tmp/nginx.conf /etc/nginx/sites-available/bingoctl.dev"

# 4. 重载 Nginx
ssh user@server "sudo nginx -t && sudo systemctl reload nginx"
```

## 配置 HTTPS（推荐）

使用 Let's Encrypt 获取免费 SSL 证书：

```bash
# 安装 certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书并自动配置 Nginx（包含 www 子域名）
sudo certbot --nginx -d bingoctl.dev -d www.bingoctl.dev

# 或使用 webroot 方式（推荐，适用于自定义配置）
sudo certbot certonly --webroot \
  -w /var/www/letsencrypt \
  -d bingoctl.dev \
  -d www.bingoctl.dev

# 设置自动续期
sudo certbot renew --dry-run

# 查看已申请的证书
sudo certbot certificates
```

**注意：** 确保 DNS 已正确配置 www 记录，否则证书申请会失败。

## 部署后验证

运行验证脚本：

```bash
bash docs/.vitepress/verify-urls.sh
```

## 相关文档

- `SEO.md` - SEO 优化完整指南
- `google-submit.md` - Google Search Console 提交指南
- `baidu-push.sh` - 百度主动推送脚本
