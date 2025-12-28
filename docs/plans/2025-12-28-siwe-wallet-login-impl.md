# SIWE 钱包登录实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将现有 Web3 钱包登录升级为 SIWE (EIP-4361) 标准，修复签名重放和钓鱼攻击漏洞。

**Architecture:** 通过配置文件控制 SIWE 启用和域名白名单。Nonce 存储从数据库迁移到 Redis（带 TTL）。复用现有的三层架构，只修改 Biz 层逻辑。

**Tech Stack:** Go 1.24+, siwe-go, Redis, Gin

**Design Doc:** `docs/plans/2025-12-28-siwe-wallet-login-design.md`

---

## Phase 1: 基础设施

### Task 1: 添加 siwe-go 依赖

**Files:**
- Modify: `go.mod`

**Step 1: 添加依赖**

```bash
go get github.com/spruceid/siwe-go
```

**Step 2: 验证依赖安装**

```bash
go mod tidy && go build ./...
```

Expected: 编译成功

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add siwe-go dependency for EIP-4361 wallet login"
```

---

### Task 2: 添加 SIWE 配置结构

**Files:**
- Modify: `internal/pkg/config/auth.go`
- Modify: `internal/pkg/config/config.go`
- Modify: `configs/bingo-apiserver.example.yaml`

**Step 1: 扩展 Auth 配置**

编辑 `internal/pkg/config/auth.go`，添加 SIWE 配置：

```go
// ABOUTME: Authentication configuration for multi-auth types.
// ABOUTME: Defines supported auth types (email/phone) and verification settings.

package config

import "time"

// Auth holds authentication configuration.
type Auth struct {
	DefaultType       string   `mapstructure:"defaulttype" json:"defaulttype" yaml:"defaulttype"`
	AllowedTypes      []string `mapstructure:"allowedtypes" json:"allowedtypes" yaml:"allowedtypes"`
	EmailVerification bool     `mapstructure:"emailverification" json:"emailverification" yaml:"emailverification"`
	PhoneVerification bool     `mapstructure:"phoneverification" json:"phoneverification" yaml:"phoneverification"`
	SIWE              SIWE     `mapstructure:"siwe" json:"siwe" yaml:"siwe"`
}

// SIWE holds Sign-In with Ethereum configuration.
type SIWE struct {
	Enabled         bool          `mapstructure:"enabled" json:"enabled" yaml:"enabled"`
	Domains         []string      `mapstructure:"domains" json:"domains" yaml:"domains"`
	Statement       string        `mapstructure:"statement" json:"statement" yaml:"statement"`
	ChainID         int           `mapstructure:"chainId" json:"chainId" yaml:"chainId"`
	NonceExpiration time.Duration `mapstructure:"nonceExpiration" json:"nonceExpiration" yaml:"nonceExpiration"`
}
```

**Step 2: 将 Auth 添加到 Config 结构体**

编辑 `internal/pkg/config/config.go`，在 Config 结构体中添加 Auth 字段：

```go
type Config struct {
	App       *App             `mapstructure:"app" json:"app" yaml:"app"`
	HTTP      *HTTP            `mapstructure:"http" json:"http" yaml:"http"`
	GRPC      *GRPC            `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	WebSocket *WebSocket       `mapstructure:"websocket" json:"websocket" yaml:"websocket"`
	Bot       *Bot             `mapstructure:"bot" json:"bot" yaml:"bot"`
	Auth      *Auth            `mapstructure:"auth" json:"auth" yaml:"auth"`  // 新增
	JWT       *JWT             `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Feature   *Feature         `mapstructure:"feature" json:"feature" yaml:"feature"`
	Mysql     *db.MySQLOptions `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Redis     *Redis           `mapstructure:"redis" json:"redis" yaml:"redis"`
	Log       *log.Options     `mapstructure:"log" json:"log" yaml:"log"`
	Mail      *mail.Options    `mapstructure:"mail" json:"mail" yaml:"mail"`
	Code      Code             `mapstructure:"code" json:"code" yaml:"code"`
	OpenAPI   OpenAPI          `mapstructure:"openapi" json:"openapi" yaml:"openapi"`
}
```

**Step 3: 更新示例配置文件**

编辑 `configs/bingo-apiserver.example.yaml`，将 auth 部分改为：

```yaml
# 认证配置
auth:
  defaulttype: email # 默认认证类型
  allowedtypes: # 允许的认证类型
    - email
    - phone
  emailverification: true # 邮箱验证
  phoneverification: true # 手机验证
  siwe:
    enabled: true                    # 是否启用 Web3 钱包登录
    domains:                         # 允许的域名白名单
      - "localhost:3000"
      - "localhost:5173"
    statement: "Sign in to Bingo"    # 签名提示文案
    chainId: 1                       # 链 ID (1=Ethereum mainnet)
    nonceExpiration: 5m              # Nonce 有效期
```

**Step 4: 验证配置加载**

```bash
make build
```

Expected: 编译成功

**Step 5: Commit**

```bash
git add internal/pkg/config/auth.go internal/pkg/config/config.go configs/bingo-apiserver.example.yaml
git commit -m "feat(config): add SIWE configuration for wallet login"
```

---

### Task 3: 添加 SIWE 相关错误码

**Files:**
- Modify: `internal/pkg/errno/code.go`

**Step 1: 添加错误码**

在 `internal/pkg/errno/code.go` 中添加：

```go
	// SIWE wallet login errors
	ErrInvalidOrigin       = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidOrigin", Message: "Invalid request origin."}
	ErrInvalidDomain       = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidDomain", Message: "Domain not allowed."}
	ErrInvalidSIWEMessage  = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidSIWEMessage", Message: "Invalid SIWE message format."}
	ErrNonceExpired        = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.NonceExpired", Message: "Nonce has expired."}
	ErrInvalidNonce        = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.InvalidNonce", Message: "Invalid or already used nonce."}
	ErrSignatureInvalid    = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignatureInvalid", Message: "Signature verification failed."}
	ErrWalletAlreadyBound  = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.WalletAlreadyBound", Message: "Wallet already bound to this account."}
	ErrWalletBoundToOther  = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.WalletBoundToOther", Message: "Wallet address already bound to another account."}
```

**Step 2: 验证编译**

```bash
make build
```

**Step 3: Commit**

```bash
git add internal/pkg/errno/code.go
git commit -m "feat(errno): add SIWE-related error codes"
```

---

### Task 4: 更新 API 请求/响应结构

**Files:**
- Modify: `pkg/api/apiserver/v1/auth.go`

**Step 1: 更新 NonceResponse 和 LoginByAddressRequest**

编辑 `pkg/api/apiserver/v1/auth.go`：

```go
type NonceResponse struct {
	Message string `json:"message"` // 完整的 SIWE 消息
	Nonce   string `json:"nonce"`   // Nonce
}

type LoginByAddressRequest struct {
	Message   string `json:"message" binding:"required"`   // SIWE 消息
	Signature string `json:"signature" binding:"required"` // 钱包签名
}
```

注意：删除原有的 `AddressRequest` 嵌入和 `Sign` 字段。

**Step 2: 更新 Swagger 文档**

```bash
make swag
```

**Step 3: 验证编译**

```bash
make build
```

**Step 4: Commit**

```bash
git add pkg/api/apiserver/v1/auth.go docs/apiserver/
git commit -m "feat(api): update NonceResponse and LoginByAddressRequest for SIWE"
```

---

## Phase 2: 核心实现

### Task 5: 重写 Nonce 方法

**Files:**
- Modify: `internal/apiserver/biz/auth/auth_address.go`

**Step 1: 重写 Nonce 函数**

完全替换 `internal/apiserver/biz/auth/auth_address.go`：

```go
// ABOUTME: SIWE (Sign-In with Ethereum) wallet authentication.
// ABOUTME: Implements EIP-4361 standard for secure wallet login.

package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	siwe "github.com/spruceid/siwe-go"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

const siweNoncePrefix = "siwe:nonce:"

func (b *authBiz) Nonce(ctx *gin.Context, req *v1.AddressRequest) (*v1.NonceResponse, error) {
	cfg := facade.Config.Auth
	if cfg == nil || !cfg.SIWE.Enabled {
		return nil, errno.ErrNotFound
	}

	// 1. 校验 Origin 在白名单
	origin := ctx.GetHeader("Origin")
	domain, err := b.validateAndExtractDomain(origin, cfg.SIWE.Domains)
	if err != nil {
		log.C(ctx).Warnw("SIWE invalid origin", "origin", origin, "allowed", cfg.SIWE.Domains)
		return nil, errno.ErrInvalidOrigin
	}

	// 2. 生成随机 nonce
	nonce := siwe.GenerateNonce()

	// 3. 构造 SIWE 消息
	uri, _ := url.Parse(origin)
	options := map[string]interface{}{
		"statement":      cfg.SIWE.Statement,
		"chainId":        cfg.SIWE.ChainID,
		"issuedAt":       time.Now().UTC().Format(time.RFC3339),
		"expirationTime": time.Now().UTC().Add(cfg.SIWE.NonceExpiration).Format(time.RFC3339),
	}

	msg, err := siwe.InitMessage(domain, req.Address, uri.String(), nonce, options)
	if err != nil {
		log.C(ctx).Errorw("SIWE init message failed", "err", err)
		return nil, errno.ErrInternal
	}

	// 4. 存储 nonce 到 Redis（带 TTL）
	key := siweNoncePrefix + nonce
	if err := facade.Redis.Set(ctx, key, req.Address, cfg.SIWE.NonceExpiration).Err(); err != nil {
		log.C(ctx).Errorw("SIWE save nonce failed", "err", err)
		return nil, errno.ErrInternal
	}

	return &v1.NonceResponse{
		Message: msg.String(),
		Nonce:   nonce,
	}, nil
}

func (b *authBiz) LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (*v1.LoginResponse, error) {
	cfg := facade.Config.Auth
	if cfg == nil || !cfg.SIWE.Enabled {
		return nil, errno.ErrNotFound
	}

	// 1. 解析 SIWE 消息
	msg, err := siwe.ParseMessage(req.Message)
	if err != nil {
		log.C(ctx).Warnw("SIWE parse message failed", "err", err)
		return nil, errno.ErrInvalidSIWEMessage
	}

	// 2. 校验 domain 在白名单
	if !b.isDomainAllowed(msg.GetDomain(), cfg.SIWE.Domains) {
		log.C(ctx).Warnw("SIWE domain not allowed", "domain", msg.GetDomain())
		return nil, errno.ErrInvalidDomain
	}

	// 3. 校验 nonce 有效（取出即删除，保证一次性）
	key := siweNoncePrefix + msg.GetNonce()
	storedAddress, err := facade.Redis.GetDel(ctx, key).Result()
	if err != nil || !strings.EqualFold(storedAddress, msg.GetAddress().Hex()) {
		log.C(ctx).Warnw("SIWE invalid nonce", "nonce", msg.GetNonce(), "stored", storedAddress, "msg_addr", msg.GetAddress().Hex())
		return nil, errno.ErrInvalidNonce
	}

	// 4. 验证签名（包含过期时间检查）
	_, err = msg.Verify(req.Signature, nil, nil, nil)
	if err != nil {
		log.C(ctx).Warnw("SIWE signature verification failed", "err", err)
		return nil, errno.ErrSignatureInvalid
	}

	// 5. 创建或获取用户
	address := msg.GetAddress().Hex()
	account, user, err := b.getOrCreateWalletUser(ctx, address)
	if err != nil {
		return nil, err
	}

	// 6. 更新登录信息
	user.LastLoginTime = pointer.Of(time.Now())
	user.LastLoginIP = ctx.ClientIP()
	user.LastLoginType = account.Provider
	_ = b.ds.User().Update(ctx, user, "last_login_time", "last_login_ip", "last_login_type")

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

// validateAndExtractDomain validates origin against whitelist and extracts domain.
func (b *authBiz) validateAndExtractDomain(origin string, allowedDomains []string) (string, error) {
	if origin == "" {
		return "", fmt.Errorf("empty origin")
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		return "", err
	}

	domain := parsed.Host
	for _, allowed := range allowedDomains {
		if strings.EqualFold(domain, allowed) {
			return domain, nil
		}
	}

	return "", fmt.Errorf("domain not in whitelist")
}

// isDomainAllowed checks if domain is in the allowed list.
func (b *authBiz) isDomainAllowed(domain string, allowedDomains []string) bool {
	for _, allowed := range allowedDomains {
		if strings.EqualFold(domain, allowed) {
			return true
		}
	}
	return false
}

// getOrCreateWalletUser finds or creates user by wallet address.
func (b *authBiz) getOrCreateWalletUser(ctx context.Context, address string) (*model.UserAccount, *model.UserM, error) {
	// 查找已有账号
	account, err := b.ds.UserAccount().GetAccount(ctx, model.AuthProviderWallet, address)
	if err == nil && account != nil {
		// 账号存在，获取用户
		user, err := b.ds.User().GetByUID(ctx, account.UID)
		if err != nil {
			return nil, nil, errno.ErrUserNotFound
		}
		return account, user, nil
	}

	// 创建新用户和账号
	uid := facade.Snowflake.Generate().String()
	user := &model.UserM{
		UID:    uid,
		Status: model.UserStatusEnabled,
	}

	account = &model.UserAccount{
		UID:       uid,
		Provider:  model.AuthProviderWallet,
		AccountID: address,
	}

	if err := b.ds.User().CreateWithAccount(ctx, user, account); err != nil {
		return nil, nil, errno.ErrDBWrite.WithMessage("create wallet user: %v", err)
	}

	return account, user, nil
}
```

**Step 2: 验证编译**

```bash
make build
```

**Step 3: Commit**

```bash
git add internal/apiserver/biz/auth/auth_address.go
git commit -m "feat(auth): implement SIWE nonce and login with EIP-4361"
```

---

### Task 6: 更新 Handler 层

**Files:**
- Modify: `internal/apiserver/handler/http/auth/address.go`

**Step 1: 检查并更新 Handler**

查看现有 handler，确保参数绑定与新的 request 结构一致。

编辑 `internal/apiserver/handler/http/auth/address.go`，更新 `LoginByAddress` 方法的参数绑定：

```go
// LoginByAddress
// @Summary    Wallet login
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.LoginByAddressRequest  true  "Param"
// @Success    200      {object}  v1.LoginResponse
// @Failure    400      {object}  core.ErrResponse
// @Failure    401      {object}  core.ErrResponse
// @Router     /v1/auth/login/address [POST].
func (h *AuthHandler) LoginByAddress(c *gin.Context) {
	var req v1.LoginByAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	resp, err := h.b.Auth().LoginByAddress(c, &req)
	core.Response(c, resp, err)
}
```

**Step 2: 更新 Swagger 文档**

```bash
make swag
```

**Step 3: 验证编译**

```bash
make build
```

**Step 4: Commit**

```bash
git add internal/apiserver/handler/http/auth/address.go docs/apiserver/
git commit -m "feat(handler): update wallet login handler for SIWE"
```

---

### Task 7: 修改 FindEnabled 追加 wallet provider

**Files:**
- Modify: `internal/apiserver/biz/auth/auth_provider.go`

**Step 1: 修改 FindEnabled 方法**

在 `internal/apiserver/biz/auth/auth_provider.go` 的 `FindEnabled` 方法末尾，data 构建完成后，追加 wallet provider：

```go
func (b *authProviderBiz) FindEnabled(ctx context.Context) (ret []*v1.AuthProviderBrief, err error) {
	list, err := b.ds.AuthProvider().FindEnabled(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*v1.AuthProviderBrief, 0)
	for _, item := range list {
		var authProvider v1.AuthProviderBrief
		_ = copier.Copy(&authProvider, item)

		// Get oauth config
		conf := oauth2.Config{
			ClientID:     item.ClientID,
			ClientSecret: item.ClientSecret,
			RedirectURL:  item.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  item.AuthURL,
				TokenURL: item.TokenURL,
			},
		}

		// Get Auth URL
		authProvider.AuthURL = conf.AuthCodeURL(uuid.New().String())

		data = append(data, &authProvider)
	}

	// 追加 wallet provider（如果 SIWE 已启用）
	if facade.Config.Auth != nil && facade.Config.Auth.SIWE.Enabled {
		data = append(data, &v1.AuthProviderBrief{
			Name:      model.AuthProviderWallet,
			IsDefault: 0,
		})
	}

	return data, err
}
```

需要添加 import：
```go
import "github.com/bingo-project/bingo/internal/pkg/facade"
```

**Step 2: 验证编译**

```bash
make build
```

**Step 3: Commit**

```bash
git add internal/apiserver/biz/auth/auth_provider.go
git commit -m "feat(auth): append wallet to providers when SIWE enabled"
```

---

## Phase 3: 钱包绑定

### Task 8: 实现钱包绑定接口

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go` (接口定义)
- Modify: `internal/apiserver/biz/auth/auth_address.go` (实现)
- Modify: `internal/apiserver/handler/http/auth/address.go` (Handler)
- Modify: `internal/apiserver/router/api.go` (路由)

**Step 1: 在 AuthBiz 接口添加 BindWallet 方法**

编辑 `internal/apiserver/biz/auth/auth.go`，在 AuthBiz 接口中添加：

```go
type AuthBiz interface {
	// ... 现有方法

	BindWallet(ctx *gin.Context, req *v1.LoginByAddressRequest, uid string) error
}
```

**Step 2: 实现 BindWallet 方法**

在 `internal/apiserver/biz/auth/auth_address.go` 添加：

```go
// BindWallet binds a wallet address to an existing user.
func (b *authBiz) BindWallet(ctx *gin.Context, req *v1.LoginByAddressRequest, uid string) error {
	cfg := facade.Config.Auth
	if cfg == nil || !cfg.SIWE.Enabled {
		return errno.ErrNotFound
	}

	// 1. 解析 SIWE 消息
	msg, err := siwe.ParseMessage(req.Message)
	if err != nil {
		return errno.ErrInvalidSIWEMessage
	}

	// 2. 校验 domain 在白名单
	if !b.isDomainAllowed(msg.GetDomain(), cfg.SIWE.Domains) {
		return errno.ErrInvalidDomain
	}

	// 3. 校验 nonce 有效
	key := siweNoncePrefix + msg.GetNonce()
	storedAddress, err := facade.Redis.GetDel(ctx, key).Result()
	if err != nil || !strings.EqualFold(storedAddress, msg.GetAddress().Hex()) {
		return errno.ErrInvalidNonce
	}

	// 4. 验证签名
	_, err = msg.Verify(req.Signature, nil, nil, nil)
	if err != nil {
		return errno.ErrSignatureInvalid
	}

	// 5. 检查地址是否已被绑定
	address := msg.GetAddress().Hex()
	existing, _ := b.ds.UserAccount().GetAccount(ctx, model.AuthProviderWallet, address)
	if existing != nil {
		if existing.UID == uid {
			return errno.ErrWalletAlreadyBound
		}
		return errno.ErrWalletBoundToOther
	}

	// 6. 创建绑定记录
	account := &model.UserAccount{
		UID:       uid,
		Provider:  model.AuthProviderWallet,
		AccountID: address,
	}

	if err := b.ds.UserAccount().Create(ctx, account); err != nil {
		return errno.ErrDBWrite.WithMessage("bind wallet: %v", err)
	}

	return nil
}
```

**Step 3: 添加 Handler 方法**

在 `internal/apiserver/handler/http/auth/address.go` 添加（与 `Nonce`、`LoginByAddress` 放在同一文件保持一致）：

```go
// BindWallet
// @Summary    Bind wallet
// @Security   Bearer
// @Tags       Auth
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.LoginByAddressRequest  true  "Param"
// @Success    200      {object}  nil
// @Failure    400      {object}  core.ErrResponse
// @Failure    401      {object}  core.ErrResponse
// @Router     /v1/auth/bindings/wallet [POST].
func (h *AuthHandler) BindWallet(c *gin.Context) {
	var req v1.LoginByAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	uid := contextx.UserID(c)
	err := h.b.Auth().BindWallet(c, &req, uid)
	core.Response(c, nil, err)
}
```

需要添加 import：
```go
import "github.com/bingo-project/bingo/pkg/contextx"
```

**Step 4: 添加路由**

编辑 `internal/apiserver/router/api.go`，在受保护的 auth 路由组中添加：

```go
// 找到 authGroup.POST("/bindings/:provider", authHandler.BindProvider) 附近
authGroup.POST("/bindings/wallet", authHandler.BindWallet)
```

**Step 5: 更新 Swagger 文档**

```bash
make swag
```

**Step 6: 验证编译**

```bash
make build
```

**Step 7: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go internal/apiserver/biz/auth/auth_address.go internal/apiserver/handler/http/auth/address.go internal/apiserver/router/api.go docs/apiserver/
git commit -m "feat(auth): add wallet binding endpoint"
```

---

## Phase 4: 单元测试

### Task 9: SIWE 辅助函数测试

**Files:**
- Create: `internal/apiserver/biz/auth/auth_address_test.go`

**Step 1: 创建测试文件**

创建 `internal/apiserver/biz/auth/auth_address_test.go`：

```go
// ABOUTME: Unit tests for SIWE wallet authentication helpers.
// ABOUTME: Tests domain validation and whitelist checking logic.

package auth

import (
	"testing"
)

func Test_validateAndExtractDomain(t *testing.T) {
	b := &authBiz{}

	tests := []struct {
		name           string
		origin         string
		allowedDomains []string
		wantDomain     string
		wantErr        bool
	}{
		{
			name:           "valid origin in whitelist",
			origin:         "http://localhost:3000",
			allowedDomains: []string{"localhost:3000", "example.com"},
			wantDomain:     "localhost:3000",
			wantErr:        false,
		},
		{
			name:           "valid https origin",
			origin:         "https://example.com",
			allowedDomains: []string{"localhost:3000", "example.com"},
			wantDomain:     "example.com",
			wantErr:        false,
		},
		{
			name:           "case insensitive match",
			origin:         "http://LOCALHOST:3000",
			allowedDomains: []string{"localhost:3000"},
			wantDomain:     "LOCALHOST:3000",
			wantErr:        false,
		},
		{
			name:           "origin not in whitelist",
			origin:         "http://evil.com",
			allowedDomains: []string{"localhost:3000", "example.com"},
			wantDomain:     "",
			wantErr:        true,
		},
		{
			name:           "empty origin",
			origin:         "",
			allowedDomains: []string{"localhost:3000"},
			wantDomain:     "",
			wantErr:        true,
		},
		{
			name:           "empty whitelist",
			origin:         "http://localhost:3000",
			allowedDomains: []string{},
			wantDomain:     "",
			wantErr:        true,
		},
		{
			name:           "invalid url",
			origin:         "not-a-valid-url",
			allowedDomains: []string{"localhost:3000"},
			wantDomain:     "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.validateAndExtractDomain(tt.origin, tt.allowedDomains)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAndExtractDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantDomain {
				t.Errorf("validateAndExtractDomain() = %v, want %v", got, tt.wantDomain)
			}
		})
	}
}

func Test_isDomainAllowed(t *testing.T) {
	b := &authBiz{}

	tests := []struct {
		name           string
		domain         string
		allowedDomains []string
		want           bool
	}{
		{
			name:           "domain in whitelist",
			domain:         "localhost:3000",
			allowedDomains: []string{"localhost:3000", "example.com"},
			want:           true,
		},
		{
			name:           "domain not in whitelist",
			domain:         "evil.com",
			allowedDomains: []string{"localhost:3000", "example.com"},
			want:           false,
		},
		{
			name:           "case insensitive",
			domain:         "LOCALHOST:3000",
			allowedDomains: []string{"localhost:3000"},
			want:           true,
		},
		{
			name:           "empty whitelist",
			domain:         "localhost:3000",
			allowedDomains: []string{},
			want:           false,
		},
		{
			name:           "empty domain",
			domain:         "",
			allowedDomains: []string{"localhost:3000"},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.isDomainAllowed(tt.domain, tt.allowedDomains); got != tt.want {
				t.Errorf("isDomainAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Step 2: 运行测试**

```bash
go test ./internal/apiserver/biz/auth/... -v -run "Test_validateAndExtractDomain|Test_isDomainAllowed"
```

Expected: 所有测试通过

**Step 3: Commit**

```bash
git add internal/apiserver/biz/auth/auth_address_test.go
git commit -m "test(auth): add unit tests for SIWE domain validation"
```

---

## Phase 5: 手动测试

### Task 10: 手动集成测试

**Step 1: 启动服务**

确保本地配置文件 `configs/bingo-apiserver.yaml` 包含 SIWE 配置：

```yaml
auth:
  siwe:
    enabled: true
    domains:
      - "localhost:3000"
      - "localhost:5173"
    statement: "Sign in to Bingo"
    chainId: 1
    nonceExpiration: 5m
```

启动服务：

```bash
make build && ./_output/platforms/darwin/arm64/bingo-apiserver
```

**Step 2: 测试获取 providers**

```bash
curl -s http://localhost:8080/v1/auth/providers | jq
```

Expected: 返回的列表中包含 `{"name": "wallet", ...}`

**Step 3: 测试获取 nonce**

```bash
curl -s -H "Origin: http://localhost:3000" \
  "http://localhost:8080/v1/auth/nonce?address=0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B" | jq
```

Expected: 返回 `{"message": "localhost:3000 wants you to sign in...", "nonce": "..."}`

**Step 4: 测试无效 Origin**

```bash
curl -s -H "Origin: http://evil.com" \
  "http://localhost:8080/v1/auth/nonce?address=0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B" | jq
```

Expected: 返回错误 `InvalidOrigin`

**Step 5: Commit 最终状态**

```bash
git add -A
git status
# 如果有未提交的更改，提交它们
```

---

## Phase 6: 清理

### Task 11: 代码检查与清理

**Step 1: 运行 lint**

```bash
make lint
```

修复任何 lint 错误。

**Step 2: 确认所有更改已提交**

```bash
git status
git log --oneline -10
```

**Step 3: 最终 Commit（如有需要）**

```bash
git add -A
git commit -m "chore: fix lint issues"
```

---

## 文件改动汇总

| 文件 | 改动类型 |
|-----|---------|
| `go.mod`, `go.sum` | 新增 siwe-go 依赖 |
| `internal/pkg/config/auth.go` | 添加 SIWE 配置结构 |
| `internal/pkg/config/config.go` | 添加 Auth 字段 |
| `configs/bingo-apiserver.example.yaml` | 添加 siwe 配置示例 |
| `internal/pkg/errno/code.go` | 添加 SIWE 错误码 |
| `pkg/api/apiserver/v1/auth.go` | 更新请求/响应结构 |
| `internal/apiserver/biz/auth/auth_address.go` | 重写 SIWE 逻辑 |
| `internal/apiserver/biz/auth/auth_address_test.go` | 新增单元测试 |
| `internal/apiserver/biz/auth/auth.go` | 添加 BindWallet 接口 |
| `internal/apiserver/biz/auth/auth_provider.go` | FindEnabled 追加 wallet |
| `internal/apiserver/handler/http/auth/address.go` | 更新 Handler，添加 BindWallet |
| `internal/apiserver/router/api.go` | 添加绑定路由 |
