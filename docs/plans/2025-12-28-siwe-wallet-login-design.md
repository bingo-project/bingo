# SIWE 钱包登录后端改造设计

## 概述

将现有的 Web3 钱包登录升级为 SIWE (Sign-In with Ethereum, EIP-4361) 标准，修复安全漏洞，提升用户体验。

## 现有实现的安全问题

| 问题 | 风险等级 | 说明 |
|-----|---------|-----|
| Nonce 无过期时间 | 中 | nonce 永久有效，泄露可被长期利用 |
| Nonce 未在使用后更新 | **高** | 同一签名可无限次重放登录 |
| 无 domain 绑定 | **高** | 钓鱼网站可重放用户签名 |
| 签名消息语义不明 | 低 | 用户只看到 UUID，不知在签什么 |

## SIWE 标准优势

- **Domain 绑定**：消息包含请求来源，防止跨站重放
- **时效性**：消息有 `expirationTime`，过期失效
- **一次性 Nonce**：使用后立即失效
- **用户可读**：清晰展示签名意图

## 技术方案

### 1. 配置结构

在 `bingo-apiserver.yaml` 新增配置：

```yaml
auth:
  # ... 现有配置

  siwe:
    enabled: true                     # 是否启用钱包登录
    domains:                          # 允许的域名白名单
      - "localhost:3000"              # 开发环境
      - "bingo.example.com"           # 生产环境
    statement: "Sign in to Bingo"     # 签名提示文案
    chainId: 1                        # 默认链 ID (1=Ethereum mainnet)
    nonceExpiration: 5m               # Nonce 有效期
```

### 2. API 改造

#### GET /v1/auth/nonce

**请求**：
```
GET /v1/auth/nonce?address=0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B
Origin: https://bingo.example.com
```

**响应**：
```json
{
  "message": "bingo.example.com wants you to sign in with your Ethereum account:\n0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B\n\nSign in to Bingo\n\nURI: https://bingo.example.com\nVersion: 1\nChain ID: 1\nNonce: k8j2m4n6p8q0r2s4\nIssued At: 2025-12-28T10:30:00.000Z\nExpiration Time: 2025-12-28T10:35:00.000Z",
  "nonce": "k8j2m4n6p8q0r2s4"
}
```

**后端逻辑**：

```go
func (b *authBiz) Nonce(ctx *gin.Context, req *v1.AddressRequest) (*v1.NonceResponse, error) {
    // 1. 校验 Origin 在白名单
    origin := ctx.GetHeader("Origin")
    domain, err := b.validateAndExtractDomain(origin)
    if err != nil {
        return nil, errno.ErrInvalidOrigin
    }

    // 2. 生成随机 nonce
    nonce := generateSecureNonce()

    // 3. 构造 SIWE 消息
    now := time.Now().UTC()
    msg := siwe.Message{
        Domain:         domain,
        Address:        common.HexToAddress(req.Address),
        Statement:      b.cfg.SIWE.Statement,
        URI:            origin,
        Version:        "1",
        ChainID:        b.cfg.SIWE.ChainID,
        Nonce:          nonce,
        IssuedAt:       now,
        ExpirationTime: now.Add(b.cfg.SIWE.NonceExpiration),
    }

    // 4. 存储 nonce 到 Redis（带 TTL，自动过期）
    key := fmt.Sprintf("siwe:nonce:%s", nonce)
    b.redis.Set(ctx, key, req.Address, b.cfg.SIWE.NonceExpiration)

    return &v1.NonceResponse{
        Message: msg.String(),
        Nonce:   nonce,
    }, nil
}
```

#### POST /v1/auth/login/address

**请求**：
```json
{
  "message": "完整的 SIWE 消息",
  "signature": "0x..."
}
```

**响应**：
```json
{
  "accessToken": "eyJhbGciOiJIUzI1NiIs...",
  "expiresAt": "2025-12-29T10:30:00.000Z"
}
```

**后端逻辑**：

```go
func (b *authBiz) LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (*v1.LoginResponse, error) {
    // 1. 解析 SIWE 消息
    msg, err := siwe.ParseMessage(req.Message)
    if err != nil {
        return nil, errno.ErrInvalidSIWEMessage
    }

    // 2. 校验 domain 在白名单
    if !b.isDomainAllowed(msg.Domain) {
        return nil, errno.ErrInvalidDomain
    }

    // 3. 校验消息未过期
    if time.Now().After(msg.ExpirationTime) {
        return nil, errno.ErrNonceExpired
    }

    // 4. 校验 nonce 有效（取出即删除，保证一次性）
    key := fmt.Sprintf("siwe:nonce:%s", msg.Nonce)
    storedAddress, err := b.redis.GetDel(ctx, key)
    if err != nil || !strings.EqualFold(storedAddress, msg.Address.Hex()) {
        return nil, errno.ErrInvalidNonce
    }

    // 5. 验证签名
    valid, err := msg.Verify(req.Signature, nil, nil, nil)
    if err != nil || !valid {
        return nil, errno.ErrSignatureInvalid
    }

    // 6. 创建或获取用户
    user, err := b.getOrCreateWalletUser(ctx, msg.Address.Hex())
    if err != nil {
        return nil, err
    }

    // 7. 生成 JWT
    t, err := token.Sign(user.UID, known.RoleUser)
    if err != nil {
        return nil, errno.ErrSignToken
    }

    return &v1.LoginResponse{
        AccessToken: t.AccessToken,
        ExpiresAt:   t.ExpiresAt,
    }, nil
}
```

#### GET /v1/auth/providers

Wallet 不需要在 `uc_auth_provider` 表中配置（该表字段为 OAuth 设计），通过配置文件控制启用：

```go
func (b *authProviderBiz) FindEnabled(ctx context.Context) ([]*v1.AuthProviderBrief, error) {
    // 1. 从数据库获取 OAuth 提供商
    list, err := b.ds.AuthProvider().FindEnabled(ctx)
    // ... 现有逻辑

    // 2. 如果配置启用了 SIWE，追加 wallet
    if b.cfg.SIWE.Enabled {
        data = append(data, &v1.AuthProviderBrief{
            Name:      "wallet",
            IsDefault: 0,
        })
    }

    return data, nil
}
```

前端通过 `GET /v1/auth/providers` 返回值判断是否显示钱包登录按钮。

#### POST /v1/auth/bindings/wallet

绑定钱包，复用登录的验证逻辑：

```go
func (b *authBiz) BindWallet(ctx *gin.Context, req *v1.LoginByAddressRequest, user *v1.UserInfo) error {
    // 1-5 与登录相同的 SIWE 验证
    msg, err := b.verifySIWE(ctx, req)
    if err != nil {
        return err
    }

    // 6. 检查地址是否已被绑定
    existing, _ := b.ds.UserAccount().GetAccount(ctx, model.AuthProviderWallet, msg.Address.Hex())
    if existing != nil {
        if existing.UID == user.UID {
            return errno.ErrAlreadyBound
        }
        return errno.ErrAddressBoundToOther
    }

    // 7. 创建绑定记录
    account := &model.UserAccount{
        UID:       user.UID,
        Provider:  model.AuthProviderWallet,
        AccountID: msg.Address.Hex(),
    }
    return b.ds.UserAccount().Create(ctx, account)
}
```

### 3. 数据结构改动

#### Request/Response 类型

```go
// pkg/api/apiserver/v1/auth.go

type NonceResponse struct {
    Message string `json:"message"` // 完整的 SIWE 消息
    Nonce   string `json:"nonce"`   // Nonce（便于前端调试）
}

type LoginByAddressRequest struct {
    Message   string `json:"message" binding:"required"`   // SIWE 消息
    Signature string `json:"signature" binding:"required"` // 钱包签名
}
```

#### 新增错误码

```go
// internal/pkg/errno/code.go

var (
    ErrInvalidOrigin      = &errno.Errno{HTTP: 400, Code: "InvalidOrigin", Message: "Invalid request origin"}
    ErrInvalidDomain      = &errno.Errno{HTTP: 400, Code: "InvalidDomain", Message: "Domain not allowed"}
    ErrInvalidSIWEMessage = &errno.Errno{HTTP: 400, Code: "InvalidSIWEMessage", Message: "Invalid SIWE message format"}
    ErrNonceExpired       = &errno.Errno{HTTP: 400, Code: "NonceExpired", Message: "Nonce has expired"}
    ErrInvalidNonce       = &errno.Errno{HTTP: 400, Code: "InvalidNonce", Message: "Invalid or already used nonce"}
    ErrSignatureInvalid   = &errno.Errno{HTTP: 401, Code: "SignatureInvalid", Message: "Signature verification failed"}
    ErrAlreadyBound       = &errno.Errno{HTTP: 400, Code: "AlreadyBound", Message: "Wallet already bound to this account"}
    ErrAddressBoundToOther = &errno.Errno{HTTP: 400, Code: "AddressBoundToOther", Message: "Address already bound to another account"}
)
```

### 4. 存储改动

| 改动 | 说明 |
|-----|------|
| Nonce 存储位置 | 数据库 → Redis（带 TTL） |
| `uc_user_account.nonce` | 字段废弃，可保留不删除 |

### 5. 依赖

```go
// go.mod
require github.com/spruceid/siwe-go v0.2.0
```

## 文件改动清单

| 文件 | 改动类型 | 说明 |
|-----|---------|------|
| `configs/bingo-apiserver.example.yaml` | 修改 | 新增 `auth.siwe` 配置块 |
| `pkg/api/apiserver/v1/auth.go` | 修改 | 更新 `NonceResponse`、`LoginByAddressRequest` |
| `internal/apiserver/biz/auth/auth_address.go` | 重写 | SIWE 验证逻辑 |
| `internal/apiserver/biz/auth/auth_provider.go` | 修改 | `FindEnabled` 追加 wallet provider |
| `internal/pkg/errno/code.go` | 修改 | 新增错误码 |
| `go.mod` | 修改 | 新增 siwe-go 依赖 |

## 安全对比

| 特性 | 旧实现 | SIWE |
|-----|-------|------|
| 签名内容 | 裸 UUID | 结构化消息 |
| Domain 绑定 | ❌ | ✅ |
| 过期时间 | ❌ | ✅ |
| Nonce 一次性 | ❌ | ✅ |
| 用户可读性 | ❌ | ✅ |
| 行业标准 | ❌ | ✅ EIP-4361 |

## 参考

- [EIP-4361: Sign-In with Ethereum](https://eips.ethereum.org/EIPS/eip-4361)
- [spruceid/siwe-go](https://github.com/spruceid/siwe-go)
