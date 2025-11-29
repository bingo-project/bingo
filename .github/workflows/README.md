# GitHub Actions 配置说明

## 文档自动部署

`deploy-docs.yml` 工作流会在以下情况自动部署文档到服务器：

- 推送到 `main` 或 `develop` 分支
- `docs/**` 目录下的文件发生变化
- 也可以手动触发部署

### 部署内容

工作流会自动完成以下操作：

1. ✅ 构建 VitePress 文档（启用 Clean URLs）
2. ✅ 验证构建产物
3. ✅ 部署文件到服务器
4. ✅ 验证部署结果（测试 Clean URLs 和重定向）

**注意：** Nginx 配置**不会**自动更新（权限限制），需要手动更新。参见 `docs/.vitepress/nginx-update.md`。

## 配置步骤

### 一、服务器端配置

#### 1. 创建部署用户

```bash
# 在服务器上（以 root 或 sudo 用户身份）
sudo adduser deploy
# 设置一个强密码（用于 sudo 操作）
```

#### 2. 配置部署目录权限

```bash
# 创建部署目录
sudo mkdir -p /var/www/bingo/docs

# 将目录所有权转给 deploy 用户
sudo chown -R deploy:deploy /var/www/bingo

# 设置适当的权限
sudo chmod -R 755 /var/www/bingo
```

#### 3. 配置 sudo 规则（可选）

**注意：** Nginx 配置现在需要**手动更新**，不再自动部署。因此不需要配置 deploy 用户的 sudo 权限。

如果将来需要自动化 Nginx 配置更新，可以参考 `docs/.vitepress/nginx-update.md`。

#### 4. 配置 SSH 目录

```bash
# 切换到 deploy 用户
sudo su - deploy

# 创建 .ssh 目录
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# 创建 authorized_keys 文件
touch ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys

# 退出 deploy 用户
exit
```

#### 5. 安装 rsync

部署工作流使用 rsync 来同步文件：

```bash
# Ubuntu/Debian
sudo apt-get update && sudo apt-get install -y rsync

# CentOS/RHEL
sudo yum install -y rsync

# 验证安装
rsync --version
```

#### 6. 配置 SSH 服务

```bash
# 编辑 SSH 配置（可选）
sudo vim /etc/ssh/sshd_config

# 确保允许公钥认证
PubkeyAuthentication yes

# 重启 SSH 服务
sudo systemctl restart sshd
```

---

### 二、本地配置

#### 1. 生成 SSH 密钥对

```bash
# 在本地机器生成专用于部署的 SSH 密钥对
ssh-keygen -t ed25519 -C "github-actions@bingoctl.dev" -f ~/.ssh/github-actions

# 生成的文件：
# - 私钥：~/.ssh/github-actions
# - 公钥：~/.ssh/github-actions.pub

# 重要：不要设置密码短语（GitHub Actions 需要无密码认证）
```

#### 2. 添加公钥到服务器

方法一：使用 ssh-copy-id（推荐）

```bash
ssh-copy-id -i ~/.ssh/github-actions.pub deploy@your-server-ip
```

方法二：手动添加

```bash
# 查看公钥内容
cat ~/.ssh/github-actions.pub

# 在服务器上（以 deploy 用户身份）
sudo su - deploy
echo "粘贴公钥内容" >> ~/.ssh/authorized_keys
exit
```

#### 3. 测试 SSH 连接

```bash
# 测试连接
ssh -i ~/.ssh/github-actions deploy@your-server-ip

# 测试目录权限
cd /var/www/bingo/docs
touch test.txt
rm test.txt

# 如果配置了 sudo，测试 nginx 重启
sudo systemctl reload nginx

# 退出
exit
```

---

### 三、GitHub 配置

#### 1. 配置 Environment Secrets

访问 GitHub 仓库：**Settings** → **Environments** → **docs**

配置以下 secrets：

| Secret 名称 | 值 | 说明 |
|------------|-----|------|
| `SSH_PRIVATE_KEY` | `~/.ssh/github-actions` 的内容 | 完整的私钥文件内容 |
| `REMOTE_HOST` | 服务器 IP 或域名 | 例如：`123.45.67.89` 或 `bingoctl.dev` |
| `REMOTE_USER` | `deploy` | 部署用户名 |

#### 2. 获取私钥内容

```bash
cat ~/.ssh/github-actions
# 复制完整输出，包括：
# -----BEGIN OPENSSH PRIVATE KEY-----
# ... (密钥内容) ...
# -----END OPENSSH PRIVATE KEY-----
```

**安全提示**：
- 私钥绝对不能提交到 git 仓库
- 私钥只保存在本地和 GitHub Secrets 中
- 定期轮换密钥（建议每 6-12 个月）

---

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

---

## 监控部署

### 查看部署状态

访问：`https://github.com/YOUR_USERNAME/bingo/actions`

### 查看部署日志

1. 点击具体的工作流运行
2. 查看每个步骤的详细日志

---

## 常见问题

### 1. 部署失败：Permission denied

**原因**：SSH 密钥权限不正确或公钥未添加到服务器。

**解决**：
```bash
# 在服务器上检查（以 deploy 用户身份）
cat ~/.ssh/authorized_keys

# 确保权限正确
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

### 2. 部署失败：目录不存在

**原因**：服务器上的目标目录不存在或权限不正确。

**解决**：
```bash
# 在服务器上
sudo mkdir -p /var/www/bingo/docs
sudo chown -R deploy:deploy /var/www/bingo
```

### 3. 构建失败：依赖安装错误

**原因**：package.json 或 package-lock.json 问题。

**解决**：
```bash
# 本地重新生成 lock 文件
rm package-lock.json
npm install
git add package-lock.json
git commit -m "chore: update package-lock.json"
```

### 4. SSH 连接超时

**原因**：服务器防火墙或安全组阻止了 SSH 连接。

**解决**：
```bash
# 检查服务器防火墙
sudo ufw status
sudo ufw allow 22/tcp

# 或检查云服务器的安全组设置，确保允许 22 端口
```

---

## 部署后验证

部署完成后，访问以下 URL 验证：

- 🌐 首页：https://bingoctl.dev
- 🇨🇳 中文文档：https://bingoctl.dev/zh/
- 🇬🇧 英文文档：https://bingoctl.dev/en/

---

## 安全建议

1. **限制权限**：deploy 用户仅拥有部署目录的写权限
2. **限制 SSH 访问**：在服务器上只允许特定 IP 访问（可选）
3. **定期轮换密钥**：建议每 6-12 个月更新 SSH 密钥
4. **监控部署日志**：定期检查部署日志，发现异常及时处理
5. **最小权限原则**：sudo 规则仅允许必要的命令

---

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

### 4. 增量部署

考虑使用 rsync 的增量传输功能，只同步变更的文件，加快部署速度。
