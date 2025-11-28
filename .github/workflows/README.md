# GitHub Actions 配置说明

## 文档自动部署

`deploy-docs.yml` 工作流会在以下情况自动部署文档到服务器：

- 推送到 `main` 或 `develop` 分支
- `docs/**` 目录下的文件发生变化
- 也可以手动触发部署

## 配置 GitHub Secrets

在部署之前，需要在 GitHub 仓库中配置以下 Secrets：

### 1. 进入仓库设置

访问：`https://github.com/YOUR_USERNAME/bingo/settings/secrets/actions`

### 2. 添加以下 Secrets

#### SSH_PRIVATE_KEY

SSH 私钥，用于连接服务器。

**生成步骤**：

```bash
# 在本地生成 SSH 密钥对（如果还没有）
ssh-keygen -t ed25519 -C "github-actions@bingoctl.dev" -f ~/.ssh/github_actions

# 将公钥添加到服务器
ssh-copy-id -i ~/.ssh/github_actions.pub user@your-server

# 或手动添加到服务器的 ~/.ssh/authorized_keys
cat ~/.ssh/github_actions.pub

# 复制私钥内容到 GitHub Secrets
cat ~/.ssh/github_actions
```

**注意**：复制整个私钥内容，包括 `-----BEGIN OPENSSH PRIVATE KEY-----` 和 `-----END OPENSSH PRIVATE KEY-----`。

#### REMOTE_HOST

服务器 IP 地址或域名。

**示例**：
- `192.168.1.100`
- `bingoctl.dev`

#### REMOTE_USER

SSH 登录用户名。

**示例**：
- `root`
- `deploy`
- `ubuntu`

## 服务器配置

### 1. 创建部署目录

```bash
# 在服务器上创建目录
sudo mkdir -p /var/www/bingo/docs

# 设置正确的权限
sudo chown -R $USER:$USER /var/www/bingo
```

### 2. 配置 SSH 访问

确保 GitHub Actions 可以通过 SSH 访问服务器：

```bash
# 编辑 SSH 配置（可选）
sudo vim /etc/ssh/sshd_config

# 确保允许公钥认证
PubkeyAuthentication yes

# 重启 SSH 服务
sudo systemctl restart sshd
```

### 3. 测试 SSH 连接

```bash
# 使用生成的密钥测试连接
ssh -i ~/.ssh/github_actions user@your-server
```

## 部署流程

### 自动部署

当你推送代码到 main 或 develop 分支时，如果 docs 目录有变化，GitHub Actions 会自动：

1. ✅ 检出代码
2. ✅ 安装 Node.js 和依赖
3. ✅ 构建 VitePress 文档
4. ✅ 通过 SSH 部署到服务器
5. ✅ 发送部署通知

### 手动部署

你也可以手动触发部署：

1. 访问 GitHub 仓库的 Actions 页面
2. 选择 "Deploy Documentation" 工作流
3. 点击 "Run workflow" 按钮
4. 选择分支并确认

## 监控部署

### 查看部署状态

访问：`https://github.com/YOUR_USERNAME/bingo/actions`

### 查看部署日志

1. 点击具体的工作流运行
2. 查看每个步骤的详细日志

### 常见问题

#### 1. 部署失败：Permission denied

**原因**：SSH 密钥权限不正确或公钥未添加到服务器。

**解决**：
```bash
# 检查服务器上的 authorized_keys
cat ~/.ssh/authorized_keys

# 确保权限正确
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

#### 2. 部署失败：目录不存在

**原因**：服务器上的目标目录不存在。

**解决**：
```bash
# 在服务器上创建目录
sudo mkdir -p /var/www/bingo/docs
sudo chown -R $USER:$USER /var/www/bingo
```

#### 3. 构建失败：依赖安装错误

**原因**：package.json 或 package-lock.json 问题。

**解决**：
```bash
# 本地重新生成 lock 文件
rm package-lock.json
npm install
git add package-lock.json
git commit -m "chore: update package-lock.json"
```

## 部署后验证

部署完成后，访问以下 URL 验证：

- 🌐 首页：https://bingoctl.dev
- 🇨🇳 中文文档：https://bingoctl.dev/zh/
- 🇬🇧 英文文档：https://bingoctl.dev/en/

## 安全建议

1. **使用专用部署用户**：不要使用 root 用户部署
2. **限制 SSH 访问**：在服务器上只允许特定 IP 访问
3. **定期轮换密钥**：定期更新 SSH 密钥
4. **使用 SSH 密钥密码**：为私钥设置密码保护（需要配置 ssh-agent）
5. **最小权限原则**：部署用户只需要对 /var/www/bingo 有写权限

## 优化建议

### 1. 添加缓存

在 workflow 中添加缓存以加速构建：

```yaml
- name: Cache node modules
  uses: actions/cache@v3
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

### 2. 并行构建

如果有多个部署目标，可以使用矩阵策略并行部署。

### 3. 部署通知

添加 Slack、Discord 或邮件通知，及时了解部署状态。
