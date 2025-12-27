# OAuth 多平台支持实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 扩展 OAuth 支持 Google、Apple、GitHub、Discord、Twitter 五个平台，增加 State 验证和 PKCE 安全机制。

**Architecture:** 通过数据库迁移添加新字段，使用 Seeder 预置平台配置模板，改造现有 OAuth 流程支持 State 和 PKCE。

**Tech Stack:** Go 1.24+, Gin, GORM, Redis, golang.org/x/oauth2

---

## Task 1: 数据库迁移 - 添加新字段

**Files:**
- Create: `internal/pkg/database/migration/2025_12_27_010000_add_auth_provider_security_fields.go`

**Step 1: 创建迁移文件**

```go
// ABOUTME: Migration to add scopes and pkce_enabled fields to uc_auth_provider table.
// ABOUTME: Supports OAuth security enhancements (PKCE) and configurable scopes.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type addAuthProviderSecurityFields struct {
	Scopes      string `gorm:"column:scopes;type:varchar(500)"`
	PKCEEnabled bool   `gorm:"column:pkce_enabled;default:false"`
}

func (addAuthProviderSecurityFields) TableName() string {
	return "uc_auth_provider"
}

func (addAuthProviderSecurityFields) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&addAuthProviderSecurityFields{})
}

func (addAuthProviderSecurityFields) Down(migrator gorm.Migrator) {
	_ = migrator.DropColumn(&addAuthProviderSecurityFields{}, "scopes")
	_ = migrator.DropColumn(&addAuthProviderSecurityFields{}, "pkce_enabled")
}

func init() {
	migrate.Add("2025_12_27_010000_add_auth_provider_security_fields", addAuthProviderSecurityFields{}.Up, addAuthProviderSecurityFields{}.Down)
}
```

**Step 2: 验证迁移文件编译通过**

Run: `go build ./internal/pkg/database/migration/...`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/pkg/database/migration/2025_12_27_010000_add_auth_provider_security_fields.go
git commit -m "feat(auth): add migration for scopes and pkce_enabled fields"
```

---

## Task 2: 数据库迁移 - Status 字段类型变更

**Files:**
- Create: `internal/pkg/database/migration/2025_12_27_020000_change_auth_provider_status_type.go`

**Step 1: 创建迁移文件**

```go
// ABOUTME: Migration to change status field type from tinyint to varchar.
// ABOUTME: Values change from 1/2 to enabled/disabled for consistency.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type changeAuthProviderStatusType struct{}

func (changeAuthProviderStatusType) TableName() string {
	return "uc_auth_provider"
}

func (changeAuthProviderStatusType) Up(migrator gorm.Migrator) {
	db := migrator.DB()
	// 1. 添加临时列
	_ = db.Exec("ALTER TABLE uc_auth_provider ADD COLUMN status_new VARCHAR(20) NOT NULL DEFAULT 'disabled'")
	// 2. 迁移数据
	_ = db.Exec("UPDATE uc_auth_provider SET status_new = CASE WHEN status = 1 THEN 'enabled' ELSE 'disabled' END")
	// 3. 删除旧列
	_ = db.Exec("ALTER TABLE uc_auth_provider DROP COLUMN status")
	// 4. 重命名新列
	_ = db.Exec("ALTER TABLE uc_auth_provider CHANGE COLUMN status_new status VARCHAR(20) NOT NULL DEFAULT 'disabled'")
}

func (changeAuthProviderStatusType) Down(migrator gorm.Migrator) {
	db := migrator.DB()
	_ = db.Exec("ALTER TABLE uc_auth_provider ADD COLUMN status_old TINYINT NOT NULL DEFAULT 2")
	_ = db.Exec("UPDATE uc_auth_provider SET status_old = CASE WHEN status = 'enabled' THEN 1 ELSE 2 END")
	_ = db.Exec("ALTER TABLE uc_auth_provider DROP COLUMN status")
	_ = db.Exec("ALTER TABLE uc_auth_provider CHANGE COLUMN status_old status TINYINT NOT NULL DEFAULT 2")
}

func init() {
	migrate.Add("2025_12_27_020000_change_auth_provider_status_type", changeAuthProviderStatusType{}.Up, changeAuthProviderStatusType{}.Down)
}
```

**Step 2: 验证迁移文件编译通过**

Run: `go build ./internal/pkg/database/migration/...`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/pkg/database/migration/2025_12_27_020000_change_auth_provider_status_type.go
git commit -m "feat(auth): add migration to change status from tinyint to varchar"
```

---

## Task 3: 更新 Model - AuthProvider

**Files:**
- Modify: `internal/pkg/model/uc_auth_provider.go`

**Step 1: 更新 AuthProvider 模型**

将整个文件内容替换为：

```go
// ABOUTME: AuthProvider model defines OAuth provider configuration.
// ABOUTME: Supports multiple OAuth platforms with configurable endpoints and PKCE.

package model

type AuthProvider struct {
	Base

	Name         string             `gorm:"type:varchar(255);not null;default:'';comment:Auth provider name"`
	Status       AuthProviderStatus `gorm:"type:varchar(20);not null;default:'disabled';comment:Status: enabled/disabled"`
	IsDefault    int                `gorm:"type:tinyint;not null;default:0;comment:Is default provider, 0-not, 1-yes"`
	AppID        string             `gorm:"type:varchar(255);not null;default:'';comment:App ID"`
	ClientID     string             `gorm:"type:varchar(255);not null;default:'';comment:Client ID"`
	ClientSecret string             `gorm:"type:varchar(1024);not null;default:'';comment:Client secret"`
	TokenType    string             `gorm:"type:varchar(1024);not null;default:'';comment:Token type"`
	RedirectURL  string             `gorm:"type:varchar(1024);not null;default:'';comment:Redirect URL"`
	AuthURL      string             `gorm:"type:varchar(1024);not null;default:'';comment:Auth URL"`
	TokenURL     string             `gorm:"type:varchar(1024);not null;default:'';comment:Token URL"`
	LogoutURI    string             `gorm:"type:varchar(1024);not null;default:'';comment:Logout URI"`
	Info         string             `gorm:"type:json;comment:Ext info"`
	UserInfoURL  string             `gorm:"column:user_info_url;type:varchar(500)"`
	FieldMapping string             `gorm:"column:field_mapping;type:text"`
	TokenInQuery bool               `gorm:"column:token_in_query;default:false"`
	ExtraHeaders string             `gorm:"column:extra_headers;type:text"`
	Scopes       string             `gorm:"column:scopes;type:varchar(500)"`
	PKCEEnabled  bool               `gorm:"column:pkce_enabled;default:false"`
}

func (*AuthProvider) TableName() string {
	return "uc_auth_provider"
}

// AuthProviderStatus enabled/disabled.
type AuthProviderStatus string

const (
	AuthProviderStatusEnabled  AuthProviderStatus = "enabled"
	AuthProviderStatusDisabled AuthProviderStatus = "disabled"

	AuthProviderGoogle  = "google"
	AuthProviderApple   = "apple"
	AuthProviderGithub  = "github"
	AuthProviderDiscord = "discord"
	AuthProviderTwitter = "twitter"
	AuthProviderWallet  = "wallet"
)
```

**Step 2: 验证编译通过**

Run: `go build ./internal/pkg/model/...`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add internal/pkg/model/uc_auth_provider.go
git commit -m "feat(auth): update AuthProvider model with new fields and status type"
```

---

## Task 4: 更新 Store - AuthProvider 查询方法

**Files:**
- Modify: `internal/pkg/store/auth_provider.go`

**Step 1: 查看当前 store 实现**

首先读取文件内容，然后更新 FirstEnabled 和 FindEnabled 方法使用新的 status 类型。

**Step 2: 更新查询条件**

将 `Status: model.AuthProviderStatusEnabled` (int) 改为 `Status: model.AuthProviderStatusEnabled` (string)。

查找并替换：
- 旧: `Status: 1` 或 `Status: model.AuthProviderStatusEnabled`（作为 int）
- 新: `Status: model.AuthProviderStatusEnabled`（作为 string）

**Step 3: 验证编译通过**

Run: `go build ./internal/pkg/store/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/pkg/store/auth_provider.go
git commit -m "fix(auth): update store queries for new status type"
```

---

## Task 5: 更新 API 类型定义

**Files:**
- Modify: `pkg/api/apiserver/v1/auth_provider.go`

**Step 1: 更新 AuthProviderInfo 和相关结构体**

添加 Scopes 和 PKCEEnabled 字段，将 Status 类型改为 string：

在 `AuthProviderInfo` 中添加：
```go
Scopes      string `json:"scopes"`      // OAuth scopes (space-separated)
PKCEEnabled bool   `json:"pkceEnabled"` // PKCE support enabled
```

将所有 `Status int` 改为 `Status string`。

**Step 2: 验证编译通过**

Run: `go build ./pkg/api/apiserver/v1/...`
Expected: 无错误输出

**Step 3: Commit**

```bash
git add pkg/api/apiserver/v1/auth_provider.go
git commit -m "feat(auth): update API types with scopes, pkce_enabled and status string"
```

---

## Task 6: 更新 Biz 层 - AuthProvider CRUD

**Files:**
- Modify: `internal/apiserver/biz/auth/auth_provider.go`

**Step 1: 更新 Update 方法**

在 Update 方法中添加对新字段的处理：

```go
if req.Scopes != nil {
	authProviderM.Scopes = *req.Scopes
}
if req.PKCEEnabled != nil {
	authProviderM.PKCEEnabled = *req.PKCEEnabled
}
```

**Step 2: 更新 FindEnabled 方法**

在构建 oauth2.Config 时使用 Scopes 字段：

```go
scopes := strings.Split(item.Scopes, " ")
if len(scopes) == 0 || scopes[0] == "" {
	scopes = []string{"user"}
}

conf := oauth2.Config{
	// ... existing fields ...
	Scopes: scopes,
}
```

需要添加 `import "strings"`。

**Step 3: 验证编译通过**

Run: `go build ./internal/apiserver/biz/auth/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/apiserver/biz/auth/auth_provider.go
git commit -m "feat(auth): add scopes and pkce_enabled handling in biz layer"
```

---

## Task 7: 创建 OAuth 工具函数 - PKCE 和 State

**Files:**
- Create: `internal/pkg/auth/oauth.go`

**Step 1: 创建 PKCE 和 State 生成函数**

```go
// ABOUTME: OAuth security utilities for PKCE and state validation.
// ABOUTME: Provides code verifier/challenge generation and state management via Redis.

package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	stateKeyPrefix = "oauth:state:"
	stateTTL       = 5 * time.Minute
)

// GenerateCodeVerifier generates a random code verifier for PKCE (43-128 chars).
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge computes S256 code challenge from verifier.
func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GenerateState generates a random state string.
func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SaveState stores state in Redis with TTL.
func SaveState(ctx context.Context, rdb *redis.Client, state string) error {
	key := stateKeyPrefix + state
	return rdb.Set(ctx, key, "1", stateTTL).Err()
}

// ValidateAndDeleteState validates state exists and deletes it (one-time use).
func ValidateAndDeleteState(ctx context.Context, rdb *redis.Client, state string) error {
	key := stateKeyPrefix + state
	result, err := rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if result == 0 {
		return fmt.Errorf("invalid or expired state")
	}
	return nil
}
```

**Step 2: 编写测试**

Create: `internal/pkg/auth/oauth_test.go`

```go
// ABOUTME: Tests for OAuth PKCE and state utilities.
// ABOUTME: Verifies code verifier/challenge generation correctness.

package auth

import (
	"testing"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("GenerateCodeVerifier failed: %v", err)
	}
	if len(verifier) < 43 {
		t.Errorf("code verifier too short: got %d chars", len(verifier))
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := GenerateCodeChallenge(verifier)
	if challenge == "" {
		t.Error("code challenge should not be empty")
	}
	if challenge == verifier {
		t.Error("code challenge should differ from verifier")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("GenerateState failed: %v", err)
	}
	state2, _ := GenerateState()
	if state1 == state2 {
		t.Error("states should be unique")
	}
}
```

**Step 3: 运行测试**

Run: `go test ./internal/pkg/auth/... -v -run TestGenerate`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/pkg/auth/oauth.go internal/pkg/auth/oauth_test.go
git commit -m "feat(auth): add PKCE and state generation utilities"
```

---

## Task 8: 更新 API 请求/响应结构

**Files:**
- Modify: `pkg/api/apiserver/v1/auth.go`

**Step 1: 更新 LoginByProviderRequest**

```go
type LoginByProviderRequest struct {
	Code         string `json:"code" form:"code"`                   // Auth code
	State        string `json:"state" form:"state"`                 // State for CSRF protection
	CodeVerifier string `json:"codeVerifier" form:"codeVerifier"`   // PKCE code verifier
}
```

**Step 2: 添加 GetAuthCodeResponse**

```go
type GetAuthCodeResponse struct {
	AuthURL      string `json:"authUrl"`               // OAuth authorization URL
	State        string `json:"state"`                 // State parameter
	CodeVerifier string `json:"codeVerifier,omitempty"` // PKCE code verifier (if PKCE enabled)
}
```

**Step 3: 验证编译通过**

Run: `go build ./pkg/api/apiserver/v1/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add pkg/api/apiserver/v1/auth.go
git commit -m "feat(auth): add state and PKCE fields to OAuth request/response"
```

---

## Task 9: 更新 Handler 层 - GetAuthCode

**Files:**
- Modify: `internal/apiserver/handler/http/auth/login.go`

**Step 1: 更新 GetAuthCode 方法**

需要调用 biz 层的新方法来生成授权 URL 和安全参数：

```go
func (ctrl *AuthHandler) GetAuthCode(c *gin.Context) {
	log.C(c).Infow("GetAuthCode function called")

	provider := c.Param("provider")
	resp, err := ctrl.b.Auth().GetAuthCode(c, provider)
	if err != nil {
		core.Response(c, nil, err)
		return
	}

	core.Response(c, resp, nil)
}
```

**Step 2: 验证编译通过**

Run: `go build ./internal/apiserver/handler/http/auth/...`
Expected: 编译错误（biz 方法尚未实现，预期失败）

**Step 3: Commit**

```bash
git add internal/apiserver/handler/http/auth/login.go
git commit -m "feat(auth): update GetAuthCode handler to use biz layer"
```

---

## Task 10: 更新 Biz 层接口 - 添加 GetAuthCode 方法

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 更新 AuthBiz 接口**

在接口中添加：
```go
GetAuthCode(ctx *gin.Context, provider string) (*v1.GetAuthCodeResponse, error)
```

**Step 2: 实现 GetAuthCode 方法**

```go
func (b *authBiz) GetAuthCode(ctx *gin.Context, providerName string) (*v1.GetAuthCodeResponse, error) {
	providerName = strings.ToLower(providerName)
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, providerName)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Parse scopes
	scopes := strings.Split(oauthProvider.Scopes, " ")
	if len(scopes) == 0 || scopes[0] == "" {
		scopes = []string{"user"}
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: scopes,
	}

	// Generate state
	state, err := auth.GenerateState()
	if err != nil {
		return nil, err
	}

	// Save state to Redis
	if err := auth.SaveState(ctx, facade.Redis, state); err != nil {
		return nil, err
	}

	resp := &v1.GetAuthCodeResponse{
		State: state,
	}

	// Build auth URL options
	opts := []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("state", state)}

	// PKCE support
	if oauthProvider.PKCEEnabled {
		codeVerifier, err := auth.GenerateCodeVerifier()
		if err != nil {
			return nil, err
		}
		codeChallenge := auth.GenerateCodeChallenge(codeVerifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		resp.CodeVerifier = codeVerifier
	}

	resp.AuthURL = conf.AuthCodeURL(state, opts...)

	return resp, nil
}
```

需要添加导入：
```go
"github.com/bingo-project/bingo/internal/pkg/auth"
"github.com/bingo-project/bingo/internal/pkg/facade"
```

**Step 3: 验证编译通过**

Run: `go build ./internal/apiserver/biz/auth/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go
git commit -m "feat(auth): implement GetAuthCode with state and PKCE support"
```

---

## Task 11: 更新 Biz 层 - LoginByProvider 添加安全验证

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 更新 LoginByProvider 方法**

在方法开头添加 state 验证，在 token 交换时添加 PKCE 支持：

```go
func (b *authBiz) LoginByProvider(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest) (*v1.LoginResponse, error) {
	// Validate state
	if req.State != "" {
		if err := auth.ValidateAndDeleteState(ctx, facade.Redis, req.State); err != nil {
			return nil, errno.ErrInvalidState
		}
	}

	// Get provider
	provider = strings.ToLower(provider)
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, provider)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Parse scopes
	scopes := strings.Split(oauthProvider.Scopes, " ")
	if len(scopes) == 0 || scopes[0] == "" {
		scopes = []string{"user"}
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: scopes,
	}

	// Exchange options
	var opts []oauth2.AuthCodeOption
	if oauthProvider.PKCEEnabled && req.CodeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", req.CodeVerifier))
	}

	// Get Access Token
	oauthToken, err := conf.Exchange(ctx, req.Code, opts...)
	if err != nil {
		return nil, err
	}

	// ... rest of existing implementation
```

**Step 2: 添加错误码**

在 `internal/pkg/errno/` 中需要添加 ErrInvalidState 错误码（如果不存在）。

**Step 3: 验证编译通过**

Run: `go build ./internal/apiserver/biz/auth/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go internal/pkg/errno/
git commit -m "feat(auth): add state and PKCE validation to LoginByProvider"
```

---

## Task 12: 添加错误码 ErrInvalidState

**Files:**
- Modify: `internal/pkg/errno/auth.go` (或合适的错误码文件)

**Step 1: 查看现有错误码结构**

首先读取错误码文件确定格式。

**Step 2: 添加 ErrInvalidState**

```go
ErrInvalidState = &errorsx.Errx{Code: 401004, Message: "Invalid or expired OAuth state"}
```

**Step 3: 验证编译通过**

Run: `go build ./internal/pkg/errno/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/pkg/errno/
git commit -m "feat(auth): add ErrInvalidState error code"
```

---

## Task 13: 创建 OAuth Provider Seeder

**Files:**
- Create: `internal/pkg/database/seeder/auth_provider_seeder.go`

**Step 1: 创建 Seeder 文件**

```go
// ABOUTME: Seeder for OAuth provider templates.
// ABOUTME: Pre-populates Google, Apple, GitHub, Discord, Twitter configurations.

package seeder

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type AuthProviderSeeder struct{}

func (AuthProviderSeeder) Signature() string {
	return "AuthProviderSeeder"
}

func (AuthProviderSeeder) Run() error {
	ctx := context.Background()
	providers := getOAuthProviderTemplates()

	for _, p := range providers {
		// Check if exists
		_, err := store.S.AuthProvider().FirstEnabled(ctx, p.Name)
		if err == nil {
			// Already exists, skip
			continue
		}

		// Check if exists but disabled
		existing, _ := store.S.AuthProvider().FindByName(ctx, p.Name)
		if existing != nil {
			continue
		}

		// Create new provider
		if err := store.S.AuthProvider().Create(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

func getOAuthProviderTemplates() []*model.AuthProvider {
	return []*model.AuthProvider{
		{
			Name:         model.AuthProviderGoogle,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
			Scopes:       "openid email profile",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "sub",
				"email":      "email",
				"nickname":   "name",
				"avatar":     "picture",
			}),
		},
		{
			Name:         model.AuthProviderApple,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://appleid.apple.com/auth/authorize",
			TokenURL:     "https://appleid.apple.com/auth/token",
			Scopes:       "name email",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "sub",
				"email":      "email",
			}),
			Info: mustJSON(map[string]string{
				"team_id":     "",
				"key_id":      "",
				"private_key": "",
			}),
		},
		{
			Name:         model.AuthProviderGithub,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scopes:       "read:user user:email",
			PKCEEnabled:  false,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "id",
				"username":   "login",
				"nickname":   "name",
				"email":      "email",
				"avatar":     "avatar_url",
				"bio":        "bio",
			}),
		},
		{
			Name:         model.AuthProviderDiscord,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://discord.com/api/oauth2/authorize",
			TokenURL:     "https://discord.com/api/oauth2/token",
			UserInfoURL:  "https://discord.com/api/users/@me",
			Scopes:       "identify email",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "id",
				"username":   "username",
				"nickname":   "global_name",
				"email":      "email",
				"avatar":     "avatar",
			}),
		},
		{
			Name:         model.AuthProviderTwitter,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://twitter.com/i/oauth2/authorize",
			TokenURL:     "https://api.twitter.com/2/oauth2/token",
			UserInfoURL:  "https://api.twitter.com/2/users/me",
			Scopes:       "users.read tweet.read",
			PKCEEnabled:  true, // Twitter OAuth 2.0 requires PKCE
			FieldMapping: mustJSON(map[string]string{
				"account_id": "data.id",
				"username":   "data.username",
				"nickname":   "data.name",
			}),
			ExtraHeaders: mustJSON(map[string]string{
				"User-Agent": "BingoApp/1.0",
			}),
		},
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
```

**Step 2: 注册 Seeder**

在 `internal/pkg/database/seeder/database_seeder.go` 的 Seeders 切片中添加 `AuthProviderSeeder{}`。

**Step 3: 添加 FindByName 方法到 Store**

需要在 `internal/pkg/store/auth_provider.go` 添加：

```go
func (s *authProviderStore) FindByName(ctx context.Context, name string) (*model.AuthProvider, error) {
	var provider model.AuthProvider
	if err := s.db.Where("name = ?", name).First(&provider).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}
```

同时在接口中添加方法签名。

**Step 4: 验证编译通过**

Run: `go build ./internal/pkg/database/seeder/...`
Expected: 无错误输出

**Step 5: Commit**

```bash
git add internal/pkg/database/seeder/auth_provider_seeder.go internal/pkg/database/seeder/database_seeder.go internal/pkg/store/auth_provider.go
git commit -m "feat(auth): add OAuth provider seeder with 5 platform templates"
```

---

## Task 14: Apple 登录特殊处理 - JWT Client Secret 生成

**Files:**
- Create: `internal/pkg/auth/apple.go`

**Step 1: 创建 Apple JWT 生成函数**

```go
// ABOUTME: Apple Sign In specific utilities.
// ABOUTME: Generates JWT client_secret required for Apple OAuth token exchange.

package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppleConfig holds Apple Sign In configuration from provider.Info.
type AppleConfig struct {
	TeamID     string `json:"team_id"`
	KeyID      string `json:"key_id"`
	PrivateKey string `json:"private_key"`
}

// GenerateAppleClientSecret generates a JWT client_secret for Apple OAuth.
// The JWT is valid for 6 months (Apple's maximum).
func GenerateAppleClientSecret(clientID string, config AppleConfig) (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(config.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse private key PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not ECDSA")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": config.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour * 24 * 180).Unix(), // 6 months
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = config.KeyID

	return token.SignedString(ecdsaKey)
}

// ParseAppleConfig parses AppleConfig from provider.Info JSON string.
func ParseAppleConfig(info string) (AppleConfig, error) {
	var config AppleConfig
	if err := json.Unmarshal([]byte(info), &config); err != nil {
		return config, err
	}
	return config, nil
}
```

**Step 2: 编写测试**

Create: `internal/pkg/auth/apple_test.go`

```go
// ABOUTME: Tests for Apple Sign In JWT generation.
// ABOUTME: Uses a test ECDSA key to verify JWT structure.

package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAppleClientSecret(t *testing.T) {
	// Generate a test key
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pkcs8Bytes, _ := x509.MarshalPKCS8PrivateKey(privateKey)
	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Bytes,
	})

	config := AppleConfig{
		TeamID:     "TEAM123",
		KeyID:      "KEY456",
		PrivateKey: string(pemBlock),
	}

	secret, err := GenerateAppleClientSecret("com.example.app", config)
	if err != nil {
		t.Fatalf("GenerateAppleClientSecret failed: %v", err)
	}

	// Parse the JWT to verify structure
	token, _, err := jwt.NewParser().ParseUnverified(secret, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Failed to parse JWT: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)
	if claims["iss"] != "TEAM123" {
		t.Errorf("Expected iss=TEAM123, got %v", claims["iss"])
	}
	if claims["sub"] != "com.example.app" {
		t.Errorf("Expected sub=com.example.app, got %v", claims["sub"])
	}
}

func TestParseAppleConfig(t *testing.T) {
	json := `{"team_id":"TEAM123","key_id":"KEY456","private_key":"-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----"}`
	config, err := ParseAppleConfig(json)
	if err != nil {
		t.Fatalf("ParseAppleConfig failed: %v", err)
	}
	if config.TeamID != "TEAM123" {
		t.Errorf("Expected TeamID=TEAM123, got %s", config.TeamID)
	}
}
```

**Step 3: 添加依赖**

Run: `go get github.com/golang-jwt/jwt/v5`

**Step 4: 运行测试**

Run: `go test ./internal/pkg/auth/... -v -run TestApple`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/pkg/auth/apple.go internal/pkg/auth/apple_test.go go.mod go.sum
git commit -m "feat(auth): add Apple Sign In JWT client_secret generation"
```

---

## Task 15: 更新 Biz 层 - Apple 特殊处理

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 在 LoginByProvider 中添加 Apple 特殊处理**

在构建 oauth2.Config 之前，检查是否为 Apple 提供商并生成动态 client_secret：

```go
clientSecret := oauthProvider.ClientSecret

// Apple: Generate JWT client_secret dynamically
if provider == model.AuthProviderApple && oauthProvider.Info != "" {
	appleConfig, err := auth.ParseAppleConfig(oauthProvider.Info)
	if err == nil && appleConfig.PrivateKey != "" {
		generatedSecret, err := auth.GenerateAppleClientSecret(oauthProvider.ClientID, appleConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Apple client secret: %w", err)
		}
		clientSecret = generatedSecret
	}
}

conf := oauth2.Config{
	ClientID:     oauthProvider.ClientID,
	ClientSecret: clientSecret,
	// ... rest
}
```

需要添加 `import "fmt"` 和更新 auth 包导入。

**Step 2: 同样更新 Bind 方法**

复制相同的 Apple 处理逻辑到 Bind 方法。

**Step 3: 验证编译通过**

Run: `go build ./internal/apiserver/biz/auth/...`
Expected: 无错误输出

**Step 4: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go
git commit -m "feat(auth): add Apple dynamic JWT client_secret generation"
```

---

## Task 16: 更新 admserver Biz 层（如果需要）

**Files:**
- Modify: `internal/admserver/biz/auth/auth_provider.go` (如果存在且需要更新)

**Step 1: 检查并更新管理后台的 AuthProvider 处理**

更新管理后台 biz 层以支持新字段的处理。

**Step 2: 验证编译通过**

Run: `go build ./internal/admserver/...`
Expected: 无错误输出

**Step 3: Commit（如有修改）**

```bash
git add internal/admserver/
git commit -m "feat(auth): update admserver biz for new auth provider fields"
```

---

## Task 17: 运行完整构建和测试

**Step 1: 运行 make build**

Run: `make build`
Expected: 所有二进制文件构建成功

**Step 2: 运行测试**

Run: `go test ./... -v`
Expected: 所有测试通过

**Step 3: 最终检查**

- 确认所有迁移文件存在
- 确认 seeder 已注册
- 确认 model、store、biz、handler 层都已更新

**Step 4: Commit（如有遗漏修复）**

---

## Task 18: 更新 Swagger 文档

**Step 1: 生成 Swagger 文档**

Run: `swag init -g cmd/bingo-apiserver/main.go -o api/swagger/apiserver --parseDependency`
Expected: 无错误

**Step 2: Commit**

```bash
git add api/swagger/
git commit -m "docs: update swagger for OAuth enhancements"
```

---

## 完成检查清单

- [ ] 数据库迁移：scopes、pkce_enabled 字段
- [ ] 数据库迁移：status 类型变更
- [ ] Model 更新：AuthProvider
- [ ] Store 更新：查询方法适配新 status 类型
- [ ] API 类型更新：请求/响应结构
- [ ] Biz 层：GetAuthCode 实现
- [ ] Biz 层：LoginByProvider 安全增强
- [ ] Biz 层：Apple JWT 动态生成
- [ ] OAuth 工具函数：PKCE、State
- [ ] Seeder：5 个平台模板
- [ ] 错误码：ErrInvalidState
- [ ] 构建通过
- [ ] 测试通过
- [ ] Swagger 更新
