# Multi-Auth Type Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 扩展 apiserver 认证系统，支持 email/phone 多认证类型，默认 email，可配置验证码验证，通用化 OAuth provider。

**Architecture:**
- 统一入口 `/auth/register` 和 `/auth/login`，通过 `account` 字段自动识别 email/phone
- 验证码通过配置项控制是否启用
- OAuth provider 通用化，通过 `user_info_url` + `field_mapping` 配置支持任意标准 OAuth2 provider

**Tech Stack:** Go, Gin, GORM, Redis (缓存验证码)

**Design Doc:** `docs/plans/2025-12-26-multi-auth-type-design.md`

---

## Phase 1: 基础设施

### Task 1: 添加 Auth 配置结构

**Files:**
- Create: `internal/pkg/config/auth.go`
- Modify: `configs/bingo-apiserver.example.yaml`

**Step 1: 创建 auth 配置结构**

```go
// internal/pkg/config/auth.go
package config

// AuthConfig 认证配置
type AuthConfig struct {
	DefaultType       string   `mapstructure:"default-type"`
	AllowedTypes      []string `mapstructure:"allowed-types"`
	EmailVerification bool     `mapstructure:"email-verification"`
	PhoneVerification bool     `mapstructure:"phone-verification"`
}
```

**Step 2: 更新 example 配置**

在 `configs/bingo-apiserver.example.yaml` 添加：

```yaml
auth:
  default-type: email
  allowed-types:
    - email
    - phone
  email-verification: true
  phone-verification: true
```

**Step 3: Commit**

```bash
git add internal/pkg/config/auth.go configs/bingo-apiserver.example.yaml
git commit -m "feat(config): add auth configuration for multi-auth types"
```

---

### Task 2: 添加新错误码

**Files:**
- Modify: `internal/pkg/errno/user.go`

**Step 1: 添加错误码定义**

在 `internal/pkg/errno/user.go` 末尾添加：

```go
	// 账号格式错误
	ErrInvalidAccountFormat = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.InvalidAccountFormat",
		Message: "Invalid account format, please enter email or phone number.",
	}

	// 验证码错误
	ErrInvalidCode = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.InvalidCode",
		Message: "Verification code is invalid or expired.",
	}

	// 注册方式未开放
	ErrAuthTypeNotAllowed = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AuthTypeNotAllowed",
		Message: "This registration method is not allowed.",
	}

	// 已绑定该类型账号
	ErrAlreadyBound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AlreadyBound",
		Message: "Already bound to this account type.",
	}

	// 账号已被占用
	ErrAccountOccupied = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AccountOccupied",
		Message: "This account is already in use by another user.",
	}

	// 短信服务未配置
	ErrSMSNotConfigured = &errorsx.ErrorX{
		Code:    http.StatusServiceUnavailable,
		Reason:  "InternalError.SMSNotConfigured",
		Message: "SMS service is not configured.",
	}

	// 未绑定该账号
	ErrNotBound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.NotBound",
		Message: "This provider is not bound to your account.",
	}

	// 不能解绑唯一登录方式
	ErrCannotUnbindLastLogin = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.CannotUnbindLastLogin",
		Message: "Cannot unbind the only login method.",
	}
```

**Step 2: Commit**

```bash
git add internal/pkg/errno/user.go
git commit -m "feat(errno): add error codes for multi-auth system"
```

---

### Task 3: 创建账号类型识别模块

**Files:**
- Create: `internal/apiserver/biz/auth/account.go`

**Step 1: 创建账号类型识别**

```go
// internal/apiserver/biz/auth/account.go
package auth

import (
	"regexp"
	"strings"

	"github.com/bingo-project/bingo/internal/pkg/errno"
)

// AccountType 账号类型
type AccountType string

const (
	AccountTypeEmail AccountType = "email"
	AccountTypePhone AccountType = "phone"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`) // 中国手机号格式
)

// DetectAccountType 自动检测账号类型
func DetectAccountType(account string) (AccountType, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return "", errno.ErrInvalidAccountFormat
	}

	// 包含 @ 且符合邮箱格式 → email
	if strings.Contains(account, "@") && emailRegex.MatchString(account) {
		return AccountTypeEmail, nil
	}

	// 符合手机号格式 → phone
	if phoneRegex.MatchString(account) {
		return AccountTypePhone, nil
	}

	return "", errno.ErrInvalidAccountFormat
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPhone 验证手机号格式
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/account.go
git commit -m "feat(auth): add account type detection for email/phone"
```

---

### Task 4: 预留 SMS 接口

**Files:**
- Create: `internal/pkg/sms/sms.go`

**Step 1: 创建 SMS 接口**

```go
// internal/pkg/sms/sms.go
package sms

import "github.com/bingo-project/bingo/internal/pkg/errno"

// SMS 短信发送接口
type SMS interface {
	Send(phone string, content string) error
}

// nopSMS 空实现（未配置时使用）
type nopSMS struct{}

// NewNopSMS 创建空实现
func NewNopSMS() SMS {
	return &nopSMS{}
}

func (n *nopSMS) Send(phone, content string) error {
	return errno.ErrSMSNotConfigured
}

// IsConfigured 检查 SMS 是否已配置
// TODO: 实际接入时根据配置判断
func IsConfigured() bool {
	return false
}
```

**Step 2: Commit**

```bash
git add internal/pkg/sms/sms.go
git commit -m "feat(sms): add SMS interface placeholder for future implementation"
```

---

## Phase 2: API 结构更新

### Task 5: 更新 Auth API 请求/响应结构

**Files:**
- Modify: `pkg/api/apiserver/v1/auth.go`

**Step 1: 添加新的请求结构**

在 `pkg/api/apiserver/v1/auth.go` 中添加/修改：

```go
// RegisterRequest 注册请求（新）
type RegisterRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Password string `json:"password" binding:"required,min=6,max=18"`
	Code     string `json:"code"`                                          // 验证码（验证开启时必填）
	Nickname string `json:"nickname" binding:"omitempty,min=2,max=255"`
}

// LoginRequest 登录请求（新）
type LoginRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Account string `json:"account" binding:"required,min=5,max=255"`
	Scene   string `json:"scene" binding:"required,oneof=register reset_password bind"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Account  string `json:"account" binding:"required,min=5,max=255"`
	Code     string `json:"code" binding:"required,len=6"`
	Password string `json:"password" binding:"required,min=6,max=18"`
}

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	Phone    *string `json:"phone" binding:"omitempty,min=11,max=11"`
	Code     string  `json:"code"`                                      // 修改 email/phone 时必填
	Nickname *string `json:"nickname" binding:"omitempty,min=2,max=255"`
}

// BindingInfo 社交账号绑定信息
type BindingInfo struct {
	Provider  string `json:"provider"`
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Avatar    string `json:"avatar"`
	BindTime  string `json:"bindTime"`
}

// ListBindingsResponse 社交账号列表响应
type ListBindingsResponse struct {
	Data []BindingInfo `json:"data"`
}
```

**Step 2: Commit**

```bash
git add pkg/api/apiserver/v1/auth.go
git commit -m "feat(api): add request/response structs for multi-auth APIs"
```

---

### Task 6: 更新 AuthProvider API 结构

**Files:**
- Modify: `pkg/api/apiserver/v1/auth_provider.go`

**Step 1: 添加新字段**

在 `AuthProviderInfo` 和相关结构中添加：

```go
// 在 CreateAuthProviderRequest 和 UpdateAuthProviderRequest 中添加
UserInfoURL  *string `json:"userInfoUrl"`
FieldMapping *string `json:"fieldMapping"`
TokenInQuery *bool   `json:"tokenInQuery"`
ExtraHeaders *string `json:"extraHeaders"`

// 在 AuthProviderInfo 中添加
UserInfoURL  string `json:"userInfoUrl"`
FieldMapping string `json:"fieldMapping"`
TokenInQuery bool   `json:"tokenInQuery"`
ExtraHeaders string `json:"extraHeaders"`
```

**Step 2: Commit**

```bash
git add pkg/api/apiserver/v1/auth_provider.go
git commit -m "feat(api): add OAuth provider generalization fields"
```

---

## Phase 3: 数据层更新

### Task 7: 更新 AuthProvider 模型

**Files:**
- Modify: `internal/pkg/model/uc_auth_provider.go`

**Step 1: 添加新字段到模型**

```go
// 在 AuthProvider 结构体中添加
UserInfoURL  string `gorm:"column:user_info_url;type:varchar(500)"`
FieldMapping string `gorm:"column:field_mapping;type:text"`
TokenInQuery bool   `gorm:"column:token_in_query;default:false"`
ExtraHeaders string `gorm:"column:extra_headers;type:text"`
```

**Step 2: Commit**

```bash
git add internal/pkg/model/uc_auth_provider.go
git commit -m "feat(model): add OAuth generalization fields to AuthProvider"
```

---

### Task 8: 添加 User Store 方法

**Files:**
- Modify: `internal/pkg/store/user.go`

**Step 1: 添加接口方法**

在 `UserStore` 接口中添加：

```go
FindByEmail(ctx context.Context, email string) (*model.UserM, error)
FindByPhone(ctx context.Context, phone string) (*model.UserM, error)
FindByAccount(ctx context.Context, account string, accountType string) (*model.UserM, error)
```

**Step 2: 实现方法**

```go
func (s *userStore) FindByEmail(ctx context.Context, email string) (*model.UserM, error) {
	var user model.UserM
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userStore) FindByPhone(ctx context.Context, phone string) (*model.UserM, error) {
	var user model.UserM
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userStore) FindByAccount(ctx context.Context, account string, accountType string) (*model.UserM, error) {
	switch accountType {
	case "email":
		return s.FindByEmail(ctx, account)
	case "phone":
		return s.FindByPhone(ctx, phone)
	default:
		return nil, errno.ErrInvalidAccountFormat
	}
}
```

**Step 3: Commit**

```bash
git add internal/pkg/store/user.go
git commit -m "feat(store): add FindByEmail/FindByPhone/FindByAccount methods"
```

---

## Phase 4: 业务层重构

### Task 9: 重构验证码逻辑

**Files:**
- Create: `internal/apiserver/biz/auth/code.go`
- Modify: `internal/apiserver/biz/common/email.go` (参考现有实现)

**Step 1: 创建统一验证码业务逻辑**

```go
// internal/apiserver/biz/auth/code.go
package auth

import (
	"context"
	"fmt"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/sms"
	"github.com/bingo-project/bingo/pkg/random"
)

// CodeScene 验证码场景
type CodeScene string

const (
	CodeSceneRegister      CodeScene = "register"
	CodeSceneResetPassword CodeScene = "reset_password"
	CodeSceneBind          CodeScene = "bind"
)

// CodeBiz 验证码业务接口
type CodeBiz interface {
	Send(ctx context.Context, account string, scene CodeScene) error
	Verify(ctx context.Context, account string, scene CodeScene, code string) error
}

type codeBiz struct {
	codeLength int
	codeTTL    int // 分钟
	codeWait   int // 分钟
}

func NewCodeBiz() CodeBiz {
	return &codeBiz{
		codeLength: 6,
		codeTTL:    5,
		codeWait:   1,
	}
}

func (b *codeBiz) Send(ctx context.Context, account string, scene CodeScene) error {
	accountType, err := DetectAccountType(account)
	if err != nil {
		return err
	}

	// 检查发送频率
	waitKey := fmt.Sprintf("verify_code_waiting:%s:%s", scene, account)
	if _, err := facade.Cache.Get(waitKey); err == nil {
		return errno.ErrTooManyRequests
	}

	// 生成验证码
	code := random.RandNumeral(b.codeLength)

	// 存储验证码
	codeKey := fmt.Sprintf("verify_code:%s:%s", scene, account)
	facade.Cache.Set(codeKey, code, b.codeTTL*60)

	// 设置发送间隔
	facade.Cache.Set(waitKey, "1", b.codeWait*60)

	// 发送验证码
	switch accountType {
	case AccountTypeEmail:
		return b.sendEmail(ctx, account, code)
	case AccountTypePhone:
		return b.sendSMS(ctx, account, code)
	}

	return nil
}

func (b *codeBiz) Verify(ctx context.Context, account string, scene CodeScene, code string) error {
	codeKey := fmt.Sprintf("verify_code:%s:%s", scene, account)
	stored, err := facade.Cache.Get(codeKey)
	if err != nil || stored != code {
		return errno.ErrInvalidCode
	}

	// 验证成功后删除
	facade.Cache.Del(codeKey)
	return nil
}

func (b *codeBiz) sendEmail(ctx context.Context, email, code string) error {
	// 复用现有 email 发送逻辑
	// TODO: 调用 email job
	return nil
}

func (b *codeBiz) sendSMS(ctx context.Context, phone, code string) error {
	if !sms.IsConfigured() {
		return errno.ErrSMSNotConfigured
	}
	// TODO: 实际发送
	return nil
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/code.go
git commit -m "feat(auth): add unified verification code business logic"
```

---

### Task 10: 重构 Register 逻辑

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 更新 Register 方法**

重构 `Register` 方法以支持 email/phone：

```go
func (b *authBiz) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.LoginResponse, error) {
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return nil, err
	}

	// TODO: 检查 allowed_types 配置

	// 验证码检查（如果需要）
	// TODO: 根据配置判断是否需要验证码
	// if b.isVerificationRequired(accountType) { ... }

	// 构建用户
	user := &model.UserM{
		Nickname: req.Nickname,
		Password: req.Password,
		Status:   model.UserStatusEnabled,
	}

	// 根据类型设置 email 或 phone
	switch accountType {
	case AccountTypeEmail:
		user.Email = req.Account
	case AccountTypePhone:
		user.Phone = req.Account
	}

	// 检查用户是否存在
	exist, err := b.ds.User().IsExist(ctx, user)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errno.ErrUserAlreadyExist
	}

	// 创建用户
	err = b.ds.User().Create(ctx, user)
	if err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*'", err.Error()); match {
			return nil, errno.ErrUserAlreadyExist
		}
		return nil, err
	}

	// 生成 token
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

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go
git commit -m "refactor(auth): update Register to support email/phone"
```

---

### Task 11: 重构 Login 逻辑

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 更新 Login 方法**

```go
func (b *authBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return nil, err
	}

	// 查找用户
	var user *model.UserM
	switch accountType {
	case AccountTypeEmail:
		user, err = b.ds.User().FindByEmail(ctx, req.Account)
	case AccountTypePhone:
		user, err = b.ds.User().FindByPhone(ctx, req.Account)
	}
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 验证密码
	if err := auth.Compare(user.Password, req.Password); err != nil {
		return nil, errno.ErrPasswordInvalid
	}

	// 更新登录信息
	user.LastLoginTime = pointer.Of(time.Now())
	user.LastLoginIP = "" // TODO: 从 context 获取
	user.LastLoginType = string(accountType)
	b.ds.User().Update(ctx, user)

	// 生成 token
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

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go
git commit -m "refactor(auth): update Login to support email/phone"
```

---

### Task 12: 添加重置密码功能

**Files:**
- Create: `internal/apiserver/biz/auth/reset_password.go`

**Step 1: 创建重置密码逻辑**

```go
// internal/apiserver/biz/auth/reset_password.go
package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type ResetPasswordBiz interface {
	ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) error
}

type resetPasswordBiz struct {
	ds      store.IStore
	codeBiz CodeBiz
}

func NewResetPasswordBiz(ds store.IStore, codeBiz CodeBiz) ResetPasswordBiz {
	return &resetPasswordBiz{ds: ds, codeBiz: codeBiz}
}

func (b *resetPasswordBiz) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) error {
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return err
	}

	// 查找用户
	var user *model.UserM
	switch accountType {
	case AccountTypeEmail:
		user, err = b.ds.User().FindByEmail(ctx, req.Account)
	case AccountTypePhone:
		user, err = b.ds.User().FindByPhone(ctx, req.Account)
	}
	if err != nil {
		return errno.ErrUserNotFound
	}

	// 验证码检查
	if err := b.codeBiz.Verify(ctx, req.Account, CodeSceneResetPassword, req.Code); err != nil {
		return err
	}

	// 更新密码
	user.Password, _ = auth.Encrypt(req.Password)
	return b.ds.User().Update(ctx, user)
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/reset_password.go
git commit -m "feat(auth): add password reset functionality"
```

---

### Task 13: 添加更新用户信息功能

**Files:**
- Create: `internal/apiserver/biz/auth/user.go`

**Step 1: 创建更新用户信息逻辑**

```go
// internal/apiserver/biz/auth/user.go
package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type UserBiz interface {
	UpdateUser(ctx context.Context, uid string, req *v1.UpdateUserRequest) error
}

type userBiz struct {
	ds      store.IStore
	codeBiz CodeBiz
}

func NewUserBiz(ds store.IStore, codeBiz CodeBiz) UserBiz {
	return &userBiz{ds: ds, codeBiz: codeBiz}
}

func (b *userBiz) UpdateUser(ctx context.Context, uid string, req *v1.UpdateUserRequest) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	// 更新 email
	if req.Email != nil && *req.Email != user.Email {
		// 检查是否被占用
		if existing, _ := b.ds.User().FindByEmail(ctx, *req.Email); existing != nil {
			return errno.ErrAccountOccupied
		}
		// 验证码检查
		if err := b.codeBiz.Verify(ctx, *req.Email, CodeSceneBind, req.Code); err != nil {
			return err
		}
		user.Email = *req.Email
	}

	// 更新 phone
	if req.Phone != nil && *req.Phone != user.Phone {
		// 检查是否被占用
		if existing, _ := b.ds.User().FindByPhone(ctx, *req.Phone); existing != nil {
			return errno.ErrAccountOccupied
		}
		// 验证码检查
		if err := b.codeBiz.Verify(ctx, *req.Phone, CodeSceneBind, req.Code); err != nil {
			return err
		}
		user.Phone = *req.Phone
	}

	// 更新 nickname
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}

	return b.ds.User().Update(ctx, user)
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/user.go
git commit -m "feat(auth): add user info update functionality"
```

---

### Task 14: 添加社交账号管理功能

**Files:**
- Create: `internal/apiserver/biz/auth/bindings.go`

**Step 1: 创建社交账号管理逻辑**

```go
// internal/apiserver/biz/auth/bindings.go
package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type BindingsBiz interface {
	ListBindings(ctx context.Context, uid string) (*v1.ListBindingsResponse, error)
	Unbind(ctx context.Context, uid string, provider string) error
}

type bindingsBiz struct {
	ds store.IStore
}

func NewBindingsBiz(ds store.IStore) BindingsBiz {
	return &bindingsBiz{ds: ds}
}

func (b *bindingsBiz) ListBindings(ctx context.Context, uid string) (*v1.ListBindingsResponse, error) {
	accounts, err := b.ds.UserAccount().FindByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	data := make([]v1.BindingInfo, 0, len(accounts))
	for _, acc := range accounts {
		data = append(data, v1.BindingInfo{
			Provider:  acc.Provider,
			AccountID: acc.AccountID,
			Username:  acc.Username,
			Avatar:    acc.Avatar,
			BindTime:  acc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return &v1.ListBindingsResponse{Data: data}, nil
}

func (b *bindingsBiz) Unbind(ctx context.Context, uid string, provider string) error {
	// 检查是否绑定了该 provider
	account, err := b.ds.UserAccount().FindByUIDAndProvider(ctx, uid, provider)
	if err != nil {
		return errno.ErrNotBound
	}

	// 检查是否为唯一登录方式
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	hasPassword := user.Email != "" || user.Phone != ""
	accountCount, _ := b.ds.UserAccount().CountByUID(ctx, uid)

	if !hasPassword && accountCount <= 1 {
		return errno.ErrCannotUnbindLastLogin
	}

	// 删除绑定
	return b.ds.UserAccount().Delete(ctx, account.ID)
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/bindings.go
git commit -m "feat(auth): add social account bindings management"
```

---

### Task 15: 通用化 OAuth GetUserInfo

**Files:**
- Modify: `internal/apiserver/biz/auth/auth.go`

**Step 1: 重写 GetUserInfo 方法**

```go
// GetUserInfo 通用化获取用户信息
func (b *authBiz) GetUserInfo(ctx context.Context, provider *model.AuthProvider, accessToken string) (*model.UserAccount, error) {
	url := provider.UserInfoURL

	// Facebook: token 放在 query parameter
	if provider.TokenInQuery {
		if strings.Contains(url, "?") {
			url += "&access_token=" + accessToken
		} else {
			url += "?access_token=" + accessToken
		}
	}

	req, _ := http.NewRequest("GET", url, nil)

	// 标准 Bearer token
	if !provider.TokenInQuery {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	// 额外 headers
	if provider.ExtraHeaders != "" {
		var headers map[string]string
		json.Unmarshal([]byte(provider.ExtraHeaders), &headers)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]any
	json.NewDecoder(resp.Body).Decode(&data)

	var mapping map[string]string
	json.Unmarshal([]byte(provider.FieldMapping), &mapping)

	account := &model.UserAccount{
		Provider:  provider.Name,
		AccountID: getNestedString(data, mapping["account_id"]),
		Username:  getNestedString(data, mapping["username"]),
		Nickname:  getNestedString(data, mapping["nickname"]),
		Email:     getNestedString(data, mapping["email"]),
		Avatar:    getNestedString(data, mapping["avatar"]),
		Bio:       getNestedString(data, mapping["bio"]),
	}

	return account, nil
}

// getNestedString 支持嵌套路径如 "data.id"
func getNestedString(data map[string]any, path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, ".")
	current := data
	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return ""
		}
		if i == len(parts)-1 {
			return cast.ToString(val)
		}
		if next, ok := val.(map[string]any); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/biz/auth/auth.go
git commit -m "refactor(auth): generalize OAuth GetUserInfo with field mapping"
```

---

## Phase 5: HTTP Handler 和路由

### Task 16: 添加新的 HTTP Handler

**Files:**
- Modify: `internal/apiserver/handler/http/auth/auth.go`

**Step 1: 添加新的 handler 方法**

```go
// SendCode 发送验证码
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req v1.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}
	// TODO: 调用 codeBiz.Send
	core.WriteResponse(c, nil, gin.H{"message": "验证码已发送"})
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req v1.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}
	// TODO: 调用 resetPasswordBiz.ResetPassword
	core.WriteResponse(c, nil, gin.H{"message": "密码重置成功"})
}

// UpdateUser 更新用户信息
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	var req v1.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}
	uid := contextx.UserID(c)
	// TODO: 调用 userBiz.UpdateUser
	core.WriteResponse(c, nil, gin.H{"message": "更新成功"})
}

// ListBindings 查询社交账号绑定
func (h *AuthHandler) ListBindings(c *gin.Context) {
	uid := contextx.UserID(c)
	// TODO: 调用 bindingsBiz.ListBindings
	core.WriteResponse(c, nil, resp)
}

// Unbind 解绑社交账号
func (h *AuthHandler) Unbind(c *gin.Context) {
	provider := c.Param("provider")
	uid := contextx.UserID(c)
	// TODO: 调用 bindingsBiz.Unbind
	core.WriteResponse(c, nil, gin.H{"message": "解绑成功"})
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/handler/http/auth/auth.go
git commit -m "feat(handler): add HTTP handlers for new auth APIs"
```

---

### Task 17: 更新路由配置

**Files:**
- Modify: `internal/apiserver/router/api.go`

**Step 1: 更新路由**

```go
// Auth routes
authGroup := v1Group.Group("/auth")
{
	// 公开接口
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/code", authHandler.SendCode)           // 新增
	authGroup.POST("/reset-password", authHandler.ResetPassword) // 新增

	// OAuth
	authGroup.GET("/providers", authHandler.Providers)
	authGroup.GET("/login/:provider", authHandler.ProviderAuthURL)
	authGroup.POST("/login/:provider", authHandler.LoginByProvider)

	// 需要登录
	authGroup.Use(authMiddleware)
	authGroup.GET("/user-info", authHandler.UserInfo)
	authGroup.PUT("/user", authHandler.UpdateUser)          // 新增
	authGroup.PUT("/change-password", authHandler.ChangePassword)

	// 社交账号管理
	authGroup.GET("/bindings", authHandler.ListBindings)           // 新增
	authGroup.POST("/bindings/:provider", authHandler.BindProvider) // 修改路径
	authGroup.DELETE("/bindings/:provider", authHandler.Unbind)    // 新增
}
```

**Step 2: Commit**

```bash
git add internal/apiserver/router/api.go
git commit -m "feat(router): update routes for multi-auth APIs"
```

---

## Phase 6: 数据库迁移

### Task 18: 添加 AuthProvider 字段迁移

**Files:**
- Create: `internal/pkg/database/migration/YYYY_MM_DD_HHMMSS_add_auth_provider_oauth_fields.go`

**Step 1: 创建迁移文件**

```go
package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func init() {
	Migrations = append(Migrations, &gormigrate.Migration{
		ID: "2025_12_26_120000_add_auth_provider_oauth_fields",
		Migrate: func(tx *gorm.DB) error {
			type AuthProvider struct {
				UserInfoURL  string `gorm:"column:user_info_url;type:varchar(500)"`
				FieldMapping string `gorm:"column:field_mapping;type:text"`
				TokenInQuery bool   `gorm:"column:token_in_query;default:false"`
				ExtraHeaders string `gorm:"column:extra_headers;type:text"`
			}
			return tx.Migrator().AutoMigrate(&AuthProvider{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropColumn("uc_auth_provider", "user_info_url", "field_mapping", "token_in_query", "extra_headers")
		},
	})
}
```

**Step 2: Commit**

```bash
git add internal/pkg/database/migration/
git commit -m "feat(migration): add OAuth generalization fields to auth_provider"
```

---

## Phase 7: 集成测试

### Task 19: 手动测试核心流程

**Step 1: 启动服务**

```bash
make run-apiserver
```

**Step 2: 测试注册**

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"account": "test@example.com", "password": "123456"}'
```

**Step 3: 测试登录**

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"account": "test@example.com", "password": "123456"}'
```

**Step 4: 验证功能正常后提交最终整理**

```bash
git add -A
git commit -m "feat: complete multi-auth type implementation"
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 1-4 | 基础设施：配置、错误码、账号识别、SMS 预留 |
| 2 | 5-6 | API 结构：Request/Response 更新 |
| 3 | 7-8 | 数据层：模型和 Store 方法 |
| 4 | 9-15 | 业务层：核心逻辑重构 |
| 5 | 16-17 | HTTP Handler 和路由 |
| 6 | 18 | 数据库迁移 |
| 7 | 19 | 集成测试 |

**Total: 19 tasks**
