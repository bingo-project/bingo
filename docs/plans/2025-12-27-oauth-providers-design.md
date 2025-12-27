# OAuth 多平台支持设计

## 背景

当前系统已有灵活的 OAuth 架构，通过数据库配置支持任意 OAuth 2.0 提供商。本设计扩展支持主流海外平台，并增强安全机制。

## 目标平台

| 优先级 | 平台 | 说明 |
|--------|------|------|
| P0 | Google | 海外用户最常用 |
| P0 | Apple | iOS 应用可能需要，海外用户信任度高 |
| P0 | GitHub | 开发者必备，已有常量定义 |
| P1 | Discord | Web3/游戏社区主流 |
| P1 | Twitter/X | 社交传播、Web3 社区常用 |

## 安全增强

### State 参数验证

防止 CSRF 攻击：

1. 用户请求登录 → 后端生成随机 state（UUID）
2. state 存入 Redis（key: `oauth:state:{state}`, ttl: 5 分钟）
3. 重定向时带上 state 参数
4. OAuth 回调时验证 state 是否存在且未过期
5. 验证通过后删除 state，防止重放

### PKCE（Proof Key for Code Exchange）

增强移动端和 SPA 安全性：

1. 客户端生成随机 code_verifier（43-128 字符）
2. 计算 code_challenge = BASE64URL(SHA256(code_verifier))
3. 授权请求带上 code_challenge 和 code_challenge_method=S256
4. Token 请求时带上原始 code_verifier
5. OAuth 服务器验证 SHA256(code_verifier) == code_challenge

## 平台配置模板

### 配置数据

| 平台 | AuthURL | Scopes | PKCE |
|------|---------|--------|------|
| Google | accounts.google.com/o/oauth2/v2/auth | openid email profile | 支持 |
| Apple | appleid.apple.com/auth/authorize | name email | 支持 |
| GitHub | github.com/login/oauth/authorize | read:user user:email | 不支持 |
| Discord | discord.com/api/oauth2/authorize | identify email | 支持 |
| Twitter | twitter.com/i/oauth2/authorize | users.read tweet.read | 必须 |

### 配置方式

使用 Seeder 预置平台配置记录（status=disabled），管理员只需：
1. 填写 client_id、client_secret、redirect_url
2. 启用平台

Seeder 幂等处理：已存在的记录跳过，保留管理员修改。

## API 改造

### GET /v1/auth/login/:provider

当前返回：
```json
{ "auth_url": "https://..." }
```

改造后：
```json
{
  "auth_url": "https://...&state=abc&code_challenge=xyz&code_challenge_method=S256",
  "state": "abc123",
  "code_verifier": "原始verifier供前端保存"
}
```

### POST /v1/auth/login/:provider

当前参数：`code`

改造后参数：`code`, `state`, `code_verifier`

验证流程：
1. 验证 state 存在于 Redis 且未过期
2. 删除 state（防重放）
3. 用 code + code_verifier 换取 token
4. 获取用户信息，创建/登录用户

## 数据库变更

### 新增字段

`uc_auth_provider` 表：

| 字段 | 类型 | 说明 |
|------|------|------|
| scopes | varchar(500) | OAuth 权限范围，空格分隔 |
| pkce_enabled | tinyint(1) | 是否启用 PKCE，默认 0 |

### 字段类型变更

`status`: tinyint → varchar(20)，值从 1/2 改为 enabled/disabled

### Apple 特殊配置

Apple 的 client_secret 是动态生成的 JWT，需要额外配置存入 `info` JSON 字段：

```json
{
  "team_id": "Apple开发者团队ID",
  "key_id": "私钥ID",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----"
}
```

## 实现清单

1. **数据库迁移**
   - 新增 scopes、pkce_enabled 字段
   - status 字段类型变更

2. **Seeder**
   - 预置 Google、Apple、GitHub、Discord、Twitter 配置
   - 幂等处理

3. **业务逻辑**
   - GetAuthCode() — 生成 state 存入 Redis，支持 PKCE
   - LoginByProvider() — 验证 state，支持 code_verifier
   - Apple JWT client_secret 动态生成

4. **模型更新**
   - AuthProvider 模型新增字段
   - 新增平台常量
