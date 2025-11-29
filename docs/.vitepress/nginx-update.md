# 手动更新 Nginx 配置

Nginx 配置不会自动部署（权限限制），需要手动更新。

## 📋 何时需要更新

仅当 `docs/.vitepress/nginx.conf` 文件有变更时才需要手动更新。

**检查是否需要更新：**
```bash
git log -1 --name-only | grep nginx.conf
```

如果有输出，说明 nginx.conf 有变更，需要手动更新。

---

## 🚀 手动更新步骤

### 方式一：使用 SCP（推荐）

```bash
# 1. 上传配置到服务器
scp docs/.vitepress/nginx.conf root@你的服务器IP:/tmp/

# 2. SSH 登录服务器
ssh root@你的服务器IP

# 3. 备份旧配置
cp /etc/nginx/sites-available/bingoctl.dev /etc/nginx/sites-available/bingoctl.dev.backup-$(date +%Y%m%d)

# 4. 应用新配置
cp /tmp/nginx.conf /etc/nginx/sites-available/bingoctl.dev

# 5. 测试配置
nginx -t

# 6. 如果测试通过，重载 Nginx
systemctl reload nginx

# 7. 清理临时文件
rm /tmp/nginx.conf

# 8. 退出
exit
```

### 方式二：直接编辑

```bash
# 1. SSH 登录服务器
ssh root@你的服务器IP

# 2. 备份配置
cp /etc/nginx/sites-available/bingoctl.dev /etc/nginx/sites-available/bingoctl.dev.backup-$(date +%Y%m%d)

# 3. 编辑配置
vim /etc/nginx/sites-available/bingoctl.dev

# 4. 测试配置
nginx -t

# 5. 重载 Nginx
systemctl reload nginx

# 6. 退出
exit
```

---

## ✅ 验证更新

更新后，验证重定向规则是否生效：

```bash
# 测试 clean URL
curl -I https://bingoctl.dev/guide/what-is-bingo
# 应该返回 HTTP/2 200

# 测试 .html 重定向
curl -I https://bingoctl.dev/guide/what-is-bingo.html
# 应该返回 HTTP/2 301

# 测试 .html/ 重定向
curl -I https://bingoctl.dev/guide/what-is-bingo.html/
# 应该返回 HTTP/2 301
```

---

## 🔄 回滚配置

如果更新后有问题，快速回滚：

```bash
# SSH 登录服务器
ssh root@你的服务器IP

# 查看备份
ls -lh /etc/nginx/sites-available/bingoctl.dev.backup-*

# 恢复备份（选择最新的日期）
cp /etc/nginx/sites-available/bingoctl.dev.backup-20251129 /etc/nginx/sites-available/bingoctl.dev

# 重载 Nginx
systemctl reload nginx
```

---

## 📝 注意事项

1. **权限要求：** 需要 root 或有 sudo 权限的用户
2. **备份重要：** 每次更新前务必备份
3. **测试先行：** 更新后必须运行 `nginx -t` 测试配置
4. **更新频率：** Nginx 配置很少变动，通常只在：
   - 添加新的重定向规则
   - 修改缓存策略
   - 调整 SSL 配置
   - 添加新的 location 规则

---

## 🔗 相关文件

- 配置文件：`docs/.vitepress/nginx.conf`
- 服务器路径：`/etc/nginx/sites-available/bingoctl.dev`
- 备份目录：`/etc/nginx/sites-available/` (*.backup-* 文件)
