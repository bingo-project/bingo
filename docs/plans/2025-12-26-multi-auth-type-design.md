# 多认证类型设计文档

## 概述

将 apiserver 的注册登录系统从 username-only 扩展为支持 email/phone 多认证类型，默认 email，支持验证码验证（可配置关闭）。同时通用化 OAuth provider 实现，支持纯配置添加新 provider。

## 设计决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 认证类型 | email + phone（移除 username） | email 不验证时同样简单 |
| 入口设计 | 统一入口 + 自动识别 | 用户体验流畅，格式识别可靠 |
| 字段命名 | `account` | 国内通用，比 identity 更直观 |
| 验证流程 | 注册时一并验证 | 一步完成，避免状态管理 |
| 配置粒度 | email/phone 分开配置 | 灵活且不过度复杂 |
| 登录方式 | 仅密码登录 | YAGNI，验证码登录后续可加 |
| 绑定功能 | 支持 | 用户体验好，实现不复杂 |
| 密码重置 | 任意已绑定方式 | 更灵活的恢复途径 |
| SMS 服务 | 预留接口，延后实现 | 设计完整，实现务实 |
| OAuth 通用化 | 标准化 UserInfo 端点 | 新增 provider 无需改代码 |

## 配置结构

```yaml
# configs/bingo-apiserver.yaml

auth:
  default_type: email           # 默认认证类型: email | phone
  allowed_types: [email, phone] # 允许的认证类型
  email_verification: true      # email 注册是否需要验证码
  phone_verification: true      # phone 注册是否需要验证码

# 复用现有配置
code:
  length: 6    # 验证码长度
  ttl: 5       # 验证码有效期（分钟）
  waiting: 1   # 发送间隔（分钟）

# 预留 SMS 配置（延后实现）
sms:
  driver: ""   # 短信服务商: aliyun | twilio | ...
  # ... 具体配置待接入时定义
```

## API 设计

### 注册

```
POST /v1/auth/register
Request:
{
  "account": "jesse@example.com",  // email 或 phone，自动识别
  "password": "123456",
  "code": "123456",                // 验证码（验证开启时必填）
  "nickname": "Jesse"              // 可选，显示名称
}
Response:
{
  "accessToken": "xxx",
  "expiresAt": 1234567890
}
```

### 登录

```
POST /v1/auth/login
Request:
{
  "account": "jesse@example.com",
  "password": "123456"
}
Response:
{
  "accessToken": "xxx",
  "expiresAt": 1234567890
}
```

### 发送验证码

```
POST /v1/auth/code
Request:
{
  "account": "jesse@example.com",  // email 或 phone
  "scene": "register"              // register | reset_password | bind
}
Response:
{
  "message": "验证码已发送"
}
```

### 重置密码

```
POST /v1/auth/reset-password
Request:
{
  "account": "jesse@example.com",
  "code": "123456",
  "password": "newpassword"
}
Response:
{
  "message": "密码重置成功"
}
```

### 绑定账号

```
POST /v1/auth/bind (需登录)
Request:
{
  "account": "13800138000",  // 要绑定的 email 或 phone
  "code": "123456"
}
Response:
{
  "message": "绑定成功"
}
```

## 数据模型

### 用户模型

```go
// internal/pkg/model/user.go
type User struct {
    // 现有字段保持
    UID      string  // 主键标识
    Email    string  // 可为空，唯一
    Phone    string  // 可为空，唯一
    Nickname string  // 显示名称
    Password string  // 加密存储
    // ...

    // username 字段保留但不再用于认证
}
```

约束：email 和 phone 至少有一个非空

### AuthProvider 模型新增

```go
// internal/pkg/model/uc_auth_provider.go 新增字段

type AuthProvider struct {
    // ... 现有字段
    UserInfoURL  string `gorm:"column:user_info_url"`  // 用户信息端点
    FieldMapping string `gorm:"column:field_mapping"`  // 字段映射（JSON）
}
```

字段映射示例：

```json
// GitHub
{
  "account_id": "id",
  "username": "login",
  "nickname": "name",
  "email": "email",
  "avatar": "avatar_url",
  "bio": "bio"
}

// Google
{
  "account_id": "sub",
  "username": "email",
  "nickname": "name",
  "email": "email",
  "avatar": "picture",
  "bio": ""
}
```

## 类型识别逻辑

```go
// internal/apiserver/biz/auth/account.go

type AccountType string

const (
    AccountTypeEmail AccountType = "email"
    AccountTypePhone AccountType = "phone"
)

func DetectAccountType(account string) (AccountType, error) {
    // 包含 @ 且符合邮箱格式 → email
    if strings.Contains(account, "@") && isValidEmail(account) {
        return AccountTypeEmail, nil
    }
    // 纯数字且符合手机号格式 → phone
    if isValidPhone(account) {
        return AccountTypePhone, nil
    }
    return "", ErrInvalidAccountFormat
}
```

## 验证码流程

### 场景定义

```go
type CodeScene string

const (
    CodeSceneRegister      CodeScene = "register"
    CodeSceneResetPassword CodeScene = "reset_password"
    CodeSceneBind          CodeScene = "bind"
)
```

### 缓存 Key 设计

```
验证码存储：verify_code:{scene}:{account}
例：verify_code:register:jesse@example.com
TTL: 5 分钟（配置 code.ttl）

发送频率限制：verify_code_waiting:{scene}:{account}
TTL: 1 分钟（配置 code.waiting）
```

### 发送流程

```
POST /v1/auth/code
    ↓
检查 allowed_types 是否包含该类型
    ↓
检查频率限制 (verify_code_waiting)
    ↓ 通过
生成验证码 → 存入缓存 → 发送
    ↓
Email: 现有 email job 发送
Phone: 调用 SMS 接口（未实现时返回错误）
```

### 验证逻辑

```go
func (s *codeBiz) Verify(account string, scene CodeScene, code string) error {
    key := fmt.Sprintf("verify_code:%s:%s", scene, account)
    stored, err := facade.Cache.Get(key)
    if err != nil || stored != code {
        return ErrInvalidCode
    }
    // 验证成功后删除，防止重复使用
    facade.Cache.Del(key)
    return nil
}
```

## 核心业务流程

### 注册流程

```
POST /v1/auth/register
    ↓
DetectAccountType(account)
    ↓
检查 allowed_types 配置
    ↓
检查用户是否已存在 (email/phone)
    ↓
IsVerificationRequired(accountType)?
    ├─ 是 → 验证 code，失败则返回错误
    └─ 否 → 跳过验证
    ↓
创建用户 (email 或 phone 字段填入 account)
    ↓
生成 JWT Token 返回
```

### 登录流程

```
POST /v1/auth/login
    ↓
DetectAccountType(account)
    ↓
FindByAccount(account)
    ↓
验证密码 auth.Compare(password, user.Password)
    ↓
更新 last_login_time, last_login_ip, last_login_type
    ↓
生成 JWT Token 返回
```

### 重置密码流程

```
POST /v1/auth/reset-password
    ↓
DetectAccountType(account)
    ↓
FindByAccount(account) → 用户不存在则报错
    ↓
Verify(account, "reset_password", code)
    ↓
更新用户密码
    ↓
返回成功
```

### 绑定流程

```
POST /v1/auth/bind (需登录)
    ↓
DetectAccountType(account)
    ↓
当前用户已绑定该类型? → 报错 "已绑定"
    ↓
该 account 已被其他用户使用? → 报错 "已被占用"
    ↓
Verify(account, "bind", code)
    ↓
更新用户 email/phone 字段
    ↓
返回成功
```

## OAuth Provider 通用化

### 通用 GetUserInfo 实现

```go
func (b *authBiz) GetUserInfo(ctx context.Context, provider *model.AuthProvider, token string) (*model.UserAccount, error) {
    // 1. 调用 user_info_url
    req, _ := http.NewRequest("GET", provider.UserInfoURL, nil)
    req.Header.Set("Authorization", "Bearer "+token)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // 2. 解析 JSON 响应
    var data map[string]any
    json.NewDecoder(resp.Body).Decode(&data)

    // 3. 根据 field_mapping 提取字段
    var mapping map[string]string
    json.Unmarshal([]byte(provider.FieldMapping), &mapping)

    account := &model.UserAccount{
        Provider:  provider.Name,
        AccountID: getString(data, mapping["account_id"]),
        Username:  getString(data, mapping["username"]),
        Nickname:  getString(data, mapping["nickname"]),
        Email:     getString(data, mapping["email"]),
        Avatar:    getString(data, mapping["avatar"]),
        Bio:       getString(data, mapping["bio"]),
    }

    return account, nil
}
```

### 常见 Provider 配置参考

| Provider | UserInfoURL | AuthURL | TokenURL |
|----------|-------------|---------|----------|
| GitHub | `https://api.github.com/user` | `https://github.com/login/oauth/authorize` | `https://github.com/login/oauth/access_token` |
| Google | `https://www.googleapis.com/oauth2/v3/userinfo` | `https://accounts.google.com/o/oauth2/v2/auth` | `https://oauth2.googleapis.com/token` |
| Facebook | `https://graph.facebook.com/me?fields=id,name,email,picture` | `https://www.facebook.com/v18.0/dialog/oauth` | `https://graph.facebook.com/v18.0/oauth/access_token` |

## 错误码

```go
// internal/pkg/errno/user.go 新增

var (
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
)

// 复用现有：ErrTooManyRequests、ErrUserAlreadyExist、ErrUserNotFound
```

## SMS 接口预留

```go
// internal/pkg/sms/sms.go

type SMS interface {
    Send(phone string, content string) error
}

type nopSMS struct{}

func (n *nopSMS) Send(phone, content string) error {
    return ErrSMSNotConfigured
}

// facade 中注册
// facade.SMS = sms.New(config.SMS)
```

## 文件变更清单

### 配置相关

```
internal/pkg/config/auth.go          # 新增：Auth 配置结构
internal/pkg/config/sms.go           # 新增：SMS 配置结构（预留）
configs/bingo-apiserver.example.yaml # 修改：添加 auth 配置示例
```

### API 层

```
pkg/api/apiserver/v1/auth.go         # 修改：更新 Request/Response 结构
pkg/api/apiserver/v1/auth_provider.go # 修改：新增 UserInfoURL、FieldMapping 字段
internal/apiserver/router/api.go     # 修改：调整路由
internal/apiserver/http/auth.go      # 修改：更新 handler
```

### 业务层

```
internal/apiserver/biz/auth/auth.go           # 修改：重构 Register/Login，通用化 GetUserInfo
internal/apiserver/biz/auth/account.go        # 新增：账号类型识别
internal/apiserver/biz/auth/code.go           # 新增：验证码业务逻辑
internal/apiserver/biz/auth/bind.go           # 新增：绑定逻辑
internal/apiserver/biz/auth/reset_password.go # 新增：密码重置逻辑
```

### 数据层

```
internal/pkg/store/user.go                  # 修改：新增 FindByEmail、FindByPhone、FindByAccount
internal/pkg/model/uc_auth_provider.go      # 修改：新增 UserInfoURL、FieldMapping 字段
internal/pkg/errno/user.go                  # 修改：新增错误码
internal/pkg/database/migration/xxx.go      # 新增：migration 添加 AuthProvider 字段
```

### 基础设施

```
internal/pkg/sms/sms.go              # 新增：SMS 接口定义
internal/pkg/facade/sms.go           # 新增：SMS facade（预留）
```
