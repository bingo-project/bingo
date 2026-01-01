# SIWE 钱包登录

Bingo 支持符合 EIP-4361 标准的 Sign-In with Ethereum (SIWE) 登录方式，允许用户使用加密货币钱包安全地登录应用。

## 功能特性

- **EIP-4361 标准**：使用行业标准的签名消息格式
- **Domain 绑定**：防止跨站点重放攻击
- **Nonce 机制**：一次性随机数，防止重放
- **自动注册**：首次登录自动创建账号

## 配置指南

在 `config.yaml` 或 `bingo-apiserver.yaml` 中配置 `auth.siwe` 部分：

```yaml
auth:
  siwe:
    enabled: true                     # 启用钱包登录
    domains:                          # 允许的域名白名单（防止钓鱼/重放）
      - "localhost:3000"
      - "bingo.example.com"
    statement: "Sign in to Bingo"     # 签名提示文案
    chainId: 1                        # 默认链 ID (1=Ethereum Mainnet)
    nonceExpiration: 5m               # Nonce 有效期
```

## API 使用流程

### 1. 获取 Nonce

前端首先调用接口获取包含 nonce 的 SIWE 消息。

**请求：**

```http
GET /v1/auth/nonce?address=0x...
Origin: https://bingo.example.com
```

**响应：**

```json
{
  "message": "bingo.example.com wants you to sign in with your Ethereum account:\n0x...\n\nSign in to Bingo\n...",
  "nonce": "k8j2m4n6p8q0r2s4"
}
```

### 2. 钱包签名

前端使用钱包（如 MetaMask）对返回的 `message` 进行签名。

```javascript
// Ethers.js 示例
const signature = await signer.signMessage(message);
```

### 3. 提交登录

将消息和签名提交给后端进行验证。

**请求：**

```http
POST /v1/auth/login/address
Content-Type: application/json

{
  "message": "完整的 SIWE 消息字符串",
  "signature": "0x..."
}
```

**响应：**

```json
{
  "accessToken": "eyJhb...",
  "expiresAt": "2025-12-30T10:00:00Z"
}
```

### 4. 绑定钱包

已登录用户可以将钱包绑定到当前账号。

**请求：**

```http
POST /v1/auth/bindings/wallet
Authorization: Bearer <token>

{
  "message": "...",
  "signature": "..."
}
```

## 安全机制

- **Nonce 验证**：生成的 Nonce 存储在 Redis 中并设置 TTL，验证一次后立即删除，确保一次性使用。
- **Domain 校验**：严格校验请求头中的 `Origin` 和消息中的 `Domain` 字段是否在白名单中。
- **过期检查**：验证消息中的 `Issued At` 和 `Expiration Time`。
