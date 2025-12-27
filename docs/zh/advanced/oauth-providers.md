# OAuth 平台配置指南

Bingo 内置支持 5 个主流 OAuth 平台：Google、Apple、GitHub、Discord 和 Twitter。本文档详细说明每个平台的配置方法。

## 平台要求概览

| 平台 | 费用 | 需要 App/网站 | 审核要求 | 门槛 |
|------|------|--------------|----------|------|
| GitHub | 免费 | 不需要 | 无 | 🟢 最低 |
| Discord | 免费 | 不需要 | 无 | 🟢 最低 |
| Google | 免费 | 生产环境需要 | 需要验证 | 🟡 中等 |
| Twitter/X | 免费起步 | 不需要 | 需要开发者账户 | 🟡 中等 |
| Apple | **$99/年** | 需要 | 需要开发者账户 | 🔴 最高 |

::: tip 快速开始建议
如果你想快速测试 OAuth 功能，建议从 **GitHub** 或 **Discord** 开始，它们无需审核、即刻可用。
:::

## 通用配置

所有 OAuth 平台共享以下配置项：

| 字段 | 说明 | 示例 |
|------|------|------|
| `client_id` | OAuth 应用的 Client ID | `xxx.apps.googleusercontent.com` |
| `client_secret` | OAuth 应用的 Client Secret | `GOCSPX-xxx` |
| `redirect_url` | 授权回调地址 | `https://api.example.com/v1/auth/callback` |
| `scopes` | 请求的权限范围（空格分隔） | `openid email profile` |
| `pkce_enabled` | 是否启用 PKCE 安全机制 | `true` |

## Google

### 平台要求

- **费用**：免费
- **测试阶段**：无需审核，但只能添加最多 100 个测试用户
- **生产环境**：需要通过 [品牌验证](https://support.google.com/cloud/answer/13464321)（2-3 个工作日），需提供：
  - 公开的首页（已验证的域名）
  - 隐私政策链接
  - 服务条款链接
- 敏感/受限 scope 需要额外安全评估

### 创建 OAuth 应用

1. 访问 [Google Cloud Console](https://console.cloud.google.com/)
2. 创建新项目或选择现有项目
3. 导航到 **APIs & Services** → **Credentials**
4. 点击 **Create Credentials** → **OAuth client ID**
5. 选择应用类型（Web application）
6. 配置授权重定向 URI

### 配置参数

```json
{
  "name": "google",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID.apps.googleusercontent.com",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth",
  "token_url": "https://oauth2.googleapis.com/token",
  "user_info_url": "https://www.googleapis.com/oauth2/v3/userinfo",
  "scopes": "openid email profile",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "sub",
    "email": "email",
    "nickname": "name",
    "avatar": "picture"
  }
}
```

### 获取凭据

1. 在 Google Cloud Console 创建 OAuth 2.0 Client ID 后获取：
   - **Client ID**: 类似 `123456789.apps.googleusercontent.com`
   - **Client Secret**: 类似 `GOCSPX-xxxxxxx`

## Apple

### 平台要求

- **费用**：$99/年（[Apple Developer Program](https://developer.apple.com/programs/enroll/)）
- **身份验证**：需要护照或政府签发的身份证件
- **组织要求**：需要 D-U-N-S 编号（企业账户）
- **私钥管理**：私钥只能在创建时下载一次，务必妥善保管

::: warning 注意
Apple 是唯一需要付费的平台。如果只是测试 OAuth 功能，建议先使用其他免费平台。
:::

### 创建 OAuth 应用

1. 访问 [Apple Developer Portal](https://developer.apple.com/)
2. 导航到 **Certificates, Identifiers & Profiles**
3. 创建 **App ID**（启用 Sign In with Apple）
4. 创建 **Services ID**（用作 client_id）
5. 创建 **Key**（用于生成 client_secret）

### 配置参数

```json
{
  "name": "apple",
  "status": "enabled",
  "client_id": "com.example.app.service",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://appleid.apple.com/auth/authorize",
  "token_url": "https://appleid.apple.com/auth/token",
  "scopes": "name email",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "sub",
    "email": "email"
  },
  "info": {
    "team_id": "YOUR_TEAM_ID",
    "key_id": "YOUR_KEY_ID",
    "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
  }
}
```

### 获取凭据

| 凭据 | 说明 | 获取位置 |
|------|------|----------|
| `client_id` | Services ID | Identifiers → Services IDs |
| `team_id` | 10 位团队标识符 | 开发者账户右上角 |
| `key_id` | Sign In with Apple Key ID | Keys 页面 |
| `private_key` | ECDSA P-256 私钥 | 创建 Key 时下载（仅一次） |

::: warning 注意
Apple 私钥只能在创建时下载一次，请妥善保管。如果丢失需要重新创建 Key。
:::

### Apple 特殊机制

Apple 不使用静态的 `client_secret`，而是需要使用私钥动态生成 JWT。Bingo 会自动处理这个过程，你只需在 `info` 字段中提供必要的密钥信息。

## GitHub

### 平台要求

- **费用**：免费
- **审核**：无需审核，即刻可用
- **限制**：每小时 2000 次 token 请求，每用户最多 10 个有效 token

::: tip 推荐
GitHub 是最容易上手的平台，只需要一个 GitHub 账号即可创建 OAuth 应用。
:::

### 创建 OAuth 应用

1. 访问 [GitHub Developer Settings](https://github.com/settings/developers)
2. 点击 **New OAuth App**
3. 填写应用信息和回调 URL
4. 注册后获取 Client ID 和 Client Secret

### 配置参数

```json
{
  "name": "github",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://github.com/login/oauth/authorize",
  "token_url": "https://github.com/login/oauth/access_token",
  "user_info_url": "https://api.github.com/user",
  "scopes": "read:user user:email",
  "pkce_enabled": false,
  "field_mapping": {
    "account_id": "id",
    "username": "login",
    "nickname": "name",
    "email": "email",
    "avatar": "avatar_url",
    "bio": "bio"
  }
}
```

::: tip 提示
GitHub 目前不支持 PKCE，因此 `pkce_enabled` 设为 `false`。
:::

### 获取凭据

1. 在 OAuth App 设置页面获取：
   - **Client ID**: 20 位字符串
   - **Client Secret**: 点击 "Generate a new client secret"

## Discord

### 平台要求

- **费用**：免费
- **审核**：无需审核，即刻可用
- **限制**：部分 scope 需要 Discord 批准（如 `bot`、`guilds.join`）

::: tip 推荐
Discord 和 GitHub 一样简单，只需要一个 Discord 账号即可。
:::

### 创建 OAuth 应用

1. 访问 [Discord Developer Portal](https://discord.com/developers/applications)
2. 点击 **New Application**
3. 导航到 **OAuth2** 页面
4. 添加重定向 URL
5. 复制 Client ID 和 Client Secret

### 配置参数

```json
{
  "name": "discord",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://discord.com/api/oauth2/authorize",
  "token_url": "https://discord.com/api/oauth2/token",
  "user_info_url": "https://discord.com/api/users/@me",
  "scopes": "identify email",
  "pkce_enabled": true,
  "field_mapping": {
    "account_id": "id",
    "username": "username",
    "nickname": "global_name",
    "email": "email",
    "avatar": "avatar"
  }
}
```

### 获取凭据

1. 在 Application 的 OAuth2 页面获取：
   - **Client ID**: Application ID
   - **Client Secret**: 点击 "Reset Secret" 生成

::: warning 注意
Discord 头像字段返回的是头像 hash，完整 URL 格式为：
`https://cdn.discordapp.com/avatars/{user_id}/{avatar}.png`
:::

## Twitter

### 平台要求

- **费用**：免费版可用，高级功能需付费
- **开发者账户**：需要在 [Developer Portal](https://developer.twitter.com/) 申请
- **Access Token 有效期**：默认 2 小时，需要使用 `offline.access` scope 获取 refresh token
- **凭据安全**：Client ID 和 Secret 只显示一次，务必立即保存

### 创建 OAuth 应用

1. 访问 [Twitter Developer Portal](https://developer.twitter.com/en/portal/dashboard)
2. 创建 Project 和 App
3. 在 App 设置中启用 **OAuth 2.0**
4. 配置回调 URL
5. 获取 Client ID 和 Client Secret

### 配置参数

```json
{
  "name": "twitter",
  "status": "enabled",
  "client_id": "YOUR_CLIENT_ID",
  "client_secret": "YOUR_CLIENT_SECRET",
  "redirect_url": "https://api.example.com/v1/auth/callback",
  "auth_url": "https://twitter.com/i/oauth2/authorize",
  "token_url": "https://api.twitter.com/2/oauth2/token",
  "user_info_url": "https://api.twitter.com/2/users/me",
  "scopes": "users.read tweet.read",
  "pkce_enabled": true,
  "extra_headers": {
    "User-Agent": "BingoApp/1.0"
  },
  "field_mapping": {
    "account_id": "data.id",
    "username": "data.username",
    "nickname": "data.name"
  }
}
```

### 获取凭据

1. 在 App 的 Keys and tokens 页面获取：
   - **Client ID**: OAuth 2.0 Client ID
   - **Client Secret**: OAuth 2.0 Client Secret

::: tip Twitter API 特性
Twitter v2 API 返回的数据嵌套在 `data` 对象中，因此 `field_mapping` 使用点号路径（如 `data.id`）来提取字段。
:::

## 字段映射说明

`field_mapping` 用于将不同平台的用户信息字段映射到统一的内部字段：

| 内部字段 | 说明 |
|----------|------|
| `account_id` | 用户在平台的唯一标识符（必需） |
| `username` | 用户名 |
| `email` | 邮箱地址 |
| `nickname` | 昵称/显示名称 |
| `avatar` | 头像 URL |
| `bio` | 个人简介 |

### 嵌套字段

对于嵌套的 JSON 响应，使用点号分隔路径：

```json
{
  "field_mapping": {
    "account_id": "data.user.id",
    "nickname": "data.user.display_name"
  }
}
```

## 安全机制

### PKCE

PKCE (Proof Key for Code Exchange) 提供额外的安全层，防止授权码被劫持：

| 平台 | PKCE 支持 |
|------|-----------|
| Google | ✅ 支持（推荐启用） |
| Apple | ✅ 支持（推荐启用） |
| GitHub | ❌ 不支持 |
| Discord | ✅ 支持（推荐启用） |
| Twitter | ✅ 支持（推荐启用） |

### State 验证

所有平台都使用 `state` 参数防止 CSRF 攻击，由 Bingo 自动生成和验证（Redis 存储，5 分钟有效期）。

## API 操作

### 获取已启用的平台

```bash
GET /v1/auth/providers
```

响应：

```json
{
  "code": 0,
  "data": {
    "providers": [
      {
        "name": "google",
        "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?client_id=...&state=..."
      },
      {
        "name": "github",
        "auth_url": "https://github.com/login/oauth/authorize?client_id=...&state=..."
      }
    ]
  }
}
```

### 通过 OAuth 登录

```bash
POST /v1/auth/login/{provider}
Content-Type: application/json

{
  "code": "授权码",
  "code_verifier": "PKCE verifier（如果启用 PKCE）"
}
```

## 常见问题

### redirect_uri 不匹配

确保配置的 `redirect_url` 与平台注册的回调 URL 完全一致，包括：
- 协议（http/https）
- 域名
- 端口
- 路径

### 获取不到邮箱

某些平台需要用户授权才能获取邮箱：
- **Apple**: 用户可选择隐藏邮箱
- **GitHub**: 需要 `user:email` scope，且邮箱需设为公开
- **Discord**: 需要 `email` scope

### PKCE 验证失败

确保客户端正确实现 PKCE 流程：
1. 客户端生成 `code_verifier`（43-128 位随机字符串）
2. 计算 `code_challenge`（S256 哈希）
3. 授权时发送 `code_challenge`
4. 换取 token 时发送原始 `code_verifier`

## 相关文档

- [统一认证授权](unified-auth.md) - 认证授权架构设计
- [统一错误处理](unified-error-handling.md) - 错误码规范
