# Admin TOTP 双因素认证

AdminServer 支持基于 Time-based One-Time Password (TOTP) 的双因素认证，为管理员账号提供额外的安全保护。

## 功能特性

- **角色级强制策略**：可配置特定角色必须开启 TOTP（如 `super_admin`）
- **两步登录流程**：密码验证 -> TOTP 验证
- **安全切换**：切换到高权限角色时需验证 TOTP
- **自助管理**：管理员可自行绑定/解绑 Google Authenticator

## 管理指南

### 强制开启 TOTP

在创建或更新角色时，可以设置 `require_totp` 为 `true`。

- 拥有该角色的用户登录时，如果未绑定 TOTP，将被拒绝登录并提示绑定。
- 绑定后，登录流程将分为两步：先验证密码，再输入 6 位验证码。

### 重置管理员 TOTP

如果管理员丢失了验证设备，超级管理员可以重置其 TOTP 设置。

```http
PUT /v1/admins/:username/reset-totp
```

## API 流程

### 1. 绑定 TOTP

**步骤 1：获取密钥和二维码**

```http
POST /v1/auth/security/totp/setup
```

响应包含 `secret` 和 `otpauth_url`（可生成二维码供 App 扫描）。

**步骤 2：启用 TOTP**

用户在 App 中扫描后，输入生成的 6 位验证码进行确认。

```http
POST /v1/auth/security/totp/enable
{
  "code": "123456"
}
```

### 2. 两步登录

当角色要求 TOTP 时，普通登录接口会返回中间状态：

**请求：**
```http
POST /v1/auth/login
{
  "account": "admin",
  "password": "..."
}
```

**响应：**
```json
{
  "require_totp": true,
  "totp_token": "tmp_token_xyz"  // 临时 Token，有效期 5 分钟
}
```

前端检测到 `require_totp: true` 后，弹出验证码输入框，调用第二步：

```http
POST /v1/auth/login/totp
{
  "totp_token": "tmp_token_xyz",
  "code": "123456"
}
```

验证通过后返回正式 `access_token`。
