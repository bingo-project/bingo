# VitePress 文档部署指南

本指南介绍如何将 Bingo 文档部署到生产服务器。

## 构建文档

在部署前，先构建静态文件：

```bash
# 构建文档
npx vitepress build docs

# 构建产物位于 docs/.vitepress/dist 目录
```

## Nginx 部署

### 1. 上传文件到服务器

将构建产物上传到服务器：

```bash
# 使用 rsync 同步（推荐）
rsync -avz --delete docs/.vitepress/dist/ user@your-server:/var/www/bingo/docs/

# 或使用 scp
scp -r docs/.vitepress/dist/* user@your-server:/var/www/bingo/docs/
```

### 2. 配置 Nginx

复制 Nginx 配置文件到服务器：

```bash
# 复制配置文件
sudo cp docs/.vitepress/nginx.conf /etc/nginx/sites-available/bingo

# 创建软链接启用站点
sudo ln -s /etc/nginx/sites-available/bingo /etc/nginx/sites-enabled/

# 编辑配置文件，修改域名
sudo vim /etc/nginx/sites-available/bingo
```

### 3. 测试并重启 Nginx

```bash
# 测试配置文件语法
sudo nginx -t

# 重启 Nginx
sudo systemctl reload nginx
# 或
sudo nginx -s reload
```

### 4. 配置 HTTPS（推荐）

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

## Docker 部署（可选）

创建 `Dockerfile`：

```dockerfile
FROM nginx:alpine

# 复制构建产物
COPY docs/.vitepress/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY docs/.vitepress/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
```

构建并运行：

```bash
# 构建镜像
docker build -t bingo-docs .

# 运行容器
docker run -d -p 80:80 --name bingo bingo-docs
```

## 自动化部署

使用 GitHub Actions 自动部署：

创建 `.github/workflows/deploy-docs.yml`：

```yaml
name: Deploy Docs

on:
  push:
    branches:
      - main
    paths:
      - 'docs/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: npm install

      - name: Build docs
        run: npx vitepress build docs

      - name: Deploy to server
        uses: easingthemes/ssh-deploy@main
        env:
          SSH_PRIVATE_KEY: ${{ secrets.SSH_PRIVATE_KEY }}
          REMOTE_HOST: ${{ secrets.REMOTE_HOST }}
          REMOTE_USER: ${{ secrets.REMOTE_USER }}
          SOURCE: "docs/.vitepress/dist/"
          TARGET: "/var/www/bingo/docs/"
```

## 验证部署

访问你的域名检查部署是否成功：

```bash
# 检查网站状态
curl -I https://bingoctl.dev

# 检查多语言路由
curl https://bingoctl.dev/zh/
curl https://bingoctl.dev/en/
```

## 常见问题

### 1. 404 错误

确保 Nginx 配置中的 `try_files` 正确设置：

```nginx
location / {
    try_files $uri $uri/ $uri.html /index.html;
}
```

### 2. 静态资源 404

检查 `root` 路径是否正确指向 `dist` 目录。

### 3. 中文路径显示异常

确保 Nginx 配置文件使用 UTF-8 编码：

```nginx
charset utf-8;
```

## 性能优化建议

1. **启用 HTTP/2**
   ```nginx
   listen 443 ssl http2;
   ```

2. **配置 CDN**
   使用 Cloudflare、阿里云 CDN 等加速访问

3. **优化缓存策略**
   已在配置文件中包含，根据需要调整

4. **压缩优化**
   Gzip 压缩已配置，可考虑启用 Brotli

## 监控和日志

```bash
# 查看访问日志
tail -f /var/log/nginx/bingo-access.log

# 查看错误日志
tail -f /var/log/nginx/bingo-error.log

# 分析访问统计
cat /var/log/nginx/bingo-access.log | awk '{print $7}' | sort | uniq -c | sort -rn | head -20
```
