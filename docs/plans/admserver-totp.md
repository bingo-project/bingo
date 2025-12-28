# AdminServer TOTP 支持实现计划

## 概述

为 AdminServer 添加基于角色的 TOTP（Time-based One-Time Password）双因素认证支持。

## 设计决策

| 决策点 | 方案 |
|-------|------|
| TOTP 策略 | 角色级控制，角色可设置 `require_totp` 字段 |
| 登录流程 | 两步登录：密码验证 → TOTP 验证 |
| 角色判断 | 只看当前激活角色的 `require_totp` 设置 |
| 切换角色 | 切换到 `require_totp=true` 的角色时需验证 TOTP |
| 恢复机制 | 超级管理员可重置其他管理员的 TOTP 绑定 |

## 与 APIServer 保持一致

为降低学习成本，以下内容与 APIServer 完全一致：

| 项目 | 规范 |
|-----|------|
| 表字段 | `google_key VARCHAR(255) NOT NULL DEFAULT ''`<br>`google_status ENUM('unbind','disabled','enabled') NOT NULL DEFAULT 'unbind'` |
| 路由 | `/v1/auth/security/totp/status` (GET)<br>`/v1/auth/security/totp/setup` (POST)<br>`/v1/auth/security/totp/enable` (POST)<br>`/v1/auth/security/totp/verify` (POST)<br>`/v1/auth/security/totp/disable` (POST) |
| 请求/响应类型 | 复用 `pkg/api/apiserver/v1/security.go` |
| 错误码 | 复用 `internal/pkg/errno/user.go` 中已定义的 TOTP 错误码 |
| Biz 方法 | `GetTOTPStatus`, `SetupTOTP`, `EnableTOTP`, `VerifyTOTP`, `DisableTOTP` |

## 阶段划分

### 阶段 1：数据库变更

**目标**：为 TOTP 功能添加必要的数据库字段

> 遵循 CONVENTIONS.md 7.4：开发阶段合并到原 create table migration，避免临时 alter table

#### 1.1 更新 sys_auth_role 表的 create migration
- 文件：`internal/pkg/database/migration/2024_05_18_215212_create_sys_auth_role_table.go`
- 增加字段：`RequireTOTP bool gorm:"type:tinyint(1);not null;default:0;comment:是否强制TOTP"`
- 同步更新 `internal/pkg/model/role.go`

#### 1.2 更新 sys_auth_admin 表的 create migration
- 文件：`internal/pkg/database/migration/2024_05_18_215143_create_sys_auth_admin_table.go`
- 增加字段（与 uc_user 表一致）：
  - `GoogleKey string gorm:"column:google_key;type:varchar(255);not null;default:''"`
  - `GoogleStatus string gorm:"column:google_status;type:enum('unbind','disabled','enabled');not null;default:unbind"`
- 同步更新 `internal/pkg/model/admin.go`

#### 1.3 执行数据库重置验证
```bash
bingo migrate reset && bingo migrate up && bingo db seed
```

### 阶段 2：API 类型定义

**目标**：定义 AdminServer 特有的类型，复用通用类型

#### 2.1 复用 APIServer 的 TOTP 类型
直接 import `pkg/api/apiserver/v1` 中的：
- `TOTPStatusResponse`
- `TOTPSetupResponse`
- `TOTPEnableRequest`
- `TOTPVerifyRequest`
- `TOTPDisableRequest`

#### 2.2 更新登录相关类型
修改 `pkg/api/apiserver/v1/admin.go`（或创建 `pkg/api/admserver/v1/auth.go`）：
- `LoginResponse` 增加：
  - `RequireTOTP bool`：是否需要 TOTP 验证
  - `TOTPToken string`：两步登录临时 Token
- 新增 `TOTPLoginRequest`：
  - `TOTPToken string`：临时 Token
  - `Code string`：TOTP 验证码

#### 2.3 更新角色类型
修改 `pkg/api/apiserver/v1/role.go`：
- `CreateRoleRequest` 增加 `RequireTOTP *bool`
- `UpdateRoleRequest` 增加 `RequireTOTP *bool`
- `RoleInfo` 增加 `RequireTOTP bool`

#### 2.4 更新切换角色类型
修改 `pkg/api/apiserver/v1/admin.go`：
- `SwitchRoleRequest` 增加 `TOTPCode string`（可选）

#### 2.5 新增重置 TOTP 请求类型
- `ResetAdminTOTPRequest`（可能不需要 body，从 URL 参数获取 username）

### 阶段 3：业务逻辑层（Biz）

**目标**：实现 TOTP 核心业务逻辑

#### 3.1 创建 `internal/admserver/biz/auth/security.go`
复用 `internal/pkg/auth/totp.go` 核心函数，实现：
```go
type SecurityBiz interface {
    GetTOTPStatus(ctx context.Context, username string) (*v1.TOTPStatusResponse, error)
    SetupTOTP(ctx context.Context, username string) (*v1.TOTPSetupResponse, error)
    EnableTOTP(ctx context.Context, username string, code string) error
    VerifyTOTP(ctx context.Context, username string, code string) error
    DisableTOTP(ctx context.Context, username string, code string) error
}
```

注意：
- AdminServer 用 `username` 作为标识（APIServer 用 `uid`）
- DisableTOTP 简化：只需验证当前 TOTP 码（不需要邮箱验证码，因为管理员场景不同）

#### 3.2 更新 `internal/admserver/biz/system/login.go`
修改 `Login` 方法：
1. 验证密码
2. 获取当前角色，检查 `require_totp`
3. 如需 TOTP：
   - 检查 `google_status == 'enabled'`
   - 未启用：返回错误 `ErrTOTPRequired`（需新增）
   - 已启用：生成临时 `TOTPToken`（存 Redis，5分钟过期），返回 `RequireTOTP=true`
4. 不需要 TOTP：直接返回 JWT

新增 `LoginWithTOTP` 方法：
- 验证 `TOTPToken`（从 Redis 获取 username）
- 验证 TOTP 码
- 删除 Redis 中的 TOTPToken
- 生成并返回 JWT

#### 3.3 更新 `internal/admserver/biz/system/admin.go`
修改 `SwitchRole` 方法：
1. 获取目标角色
2. 检查 `require_totp`
3. 如需 TOTP：
   - 检查管理员 `google_status == 'enabled'`
   - 未启用：返回错误
   - 已启用：验证 `req.TOTPCode`
4. 执行角色切换

新增 `ResetTOTP` 方法：
- 检查调用者是否为 root
- 重置目标管理员的 `google_key=''`, `google_status='unbind'`

#### 3.4 更新 `internal/admserver/biz/system/role.go`
- `Create` / `Update` 方法处理 `RequireTOTP` 字段

### 阶段 4：Handler 层

**目标**：实现 HTTP 接口，与 APIServer 保持一致的风格

#### 4.1 创建 `internal/admserver/handler/http/auth/security.go`
参照 `internal/apiserver/handler/http/auth/security.go` 实现：
- `GetTOTPStatus`：GET /v1/auth/security/totp/status
- `SetupTOTP`：POST /v1/auth/security/totp/setup
- `EnableTOTP`：POST /v1/auth/security/totp/enable
- `VerifyTOTP`：POST /v1/auth/security/totp/verify
- `DisableTOTP`：POST /v1/auth/security/totp/disable

每个方法包含：
- ABOUTME 注释
- Swagger 注释
- 参数绑定和验证
- 调用 Biz 层
- 统一响应 `core.Response()`

#### 4.2 更新 `internal/admserver/handler/http/system/admin.go`
- 新增 `LoginWithTOTP`：POST /v1/auth/login/totp
- 新增 `ResetTOTP`：PUT /v1/admins/:name/reset-totp

#### 4.3 确认角色 Handler 已支持 RequireTOTP
检查 `internal/admserver/handler/http/system/role.go` 的 Create/Update

### 阶段 5：路由注册

**目标**：注册新接口路由

#### 5.1 更新 `internal/admserver/router/api.go`

登录相关（无需认证）：
```go
v1.POST("auth/login", adminHandler.Login)
v1.POST("auth/login/totp", adminHandler.LoginWithTOTP)  // 新增
```

安全设置（需认证，与 APIServer 路由一致）：
```go
// Security (TOTP)
securityHandler := auth.NewSecurityHandler(store.S)
securityGroup := v1.Group("auth/security")
{
    securityGroup.GET("/totp/status", securityHandler.GetTOTPStatus)
    securityGroup.POST("/totp/setup", securityHandler.SetupTOTP)
    securityGroup.POST("/totp/enable", securityHandler.EnableTOTP)
    securityGroup.POST("/totp/verify", securityHandler.VerifyTOTP)
    securityGroup.POST("/totp/disable", securityHandler.DisableTOTP)
}
```

管理员管理（需权限）：
```go
v1.PUT("admins/:name/reset-totp", adminHandler.ResetTOTP)  // 新增
```

### 阶段 6：错误码

**目标**：复用已有错误码，仅新增必要的

#### 6.1 复用已有错误码（`internal/pkg/errno/user.go`）
- `ErrTOTPNotEnabled`
- `ErrTOTPAlreadyEnabled`
- `ErrTOTPInvalid`
- `ErrTOTPCodeRequired`

#### 6.2 新增错误码（如需要）
在 `internal/pkg/errno/` 中添加：
- `ErrTOTPRequired`：该角色要求启用 TOTP
- `ErrTOTPTokenInvalid`：TOTP Token 无效或过期

### 阶段 7：测试

**目标**：确保功能正确性

#### 7.1 Biz 层单元测试
- `internal/admserver/biz/auth/security_test.go`
- 测试 TOTP 生命周期：setup → enable → verify → disable

#### 7.2 登录流程测试
- 测试普通登录（角色不需要 TOTP）
- 测试两步登录（角色需要 TOTP）
- 测试未绑定 TOTP 时登录到需要 TOTP 的角色
- 测试 TOTP Token 过期

#### 7.3 角色切换测试
- 测试切换到不需要 TOTP 的角色
- 测试切换到需要 TOTP 的角色（已绑定/未绑定）

#### 7.4 重置 TOTP 测试
- root 用户重置成功
- 非 root 用户重置被拒绝

### 阶段 8：文档和收尾

#### 8.1 更新 Swagger 文档
```bash
make swag
```

#### 8.2 运行 lint 检查
```bash
make lint
```

#### 8.3 构建验证
```bash
make build
```

## 文件变更清单

### 新增文件
| 文件路径 | 说明 |
|---------|------|
| `internal/admserver/biz/auth/security.go` | TOTP 业务逻辑 |
| `internal/admserver/handler/http/auth/security.go` | TOTP HTTP Handler |
| `internal/admserver/biz/auth/security_test.go` | 单元测试 |

### 修改文件
| 文件路径 | 说明 |
|---------|------|
| `internal/pkg/database/migration/2024_05_18_215212_create_sys_auth_role_table.go` | 增加 require_totp |
| `internal/pkg/database/migration/2024_05_18_215143_create_sys_auth_admin_table.go` | 增加 TOTP 字段 |
| `internal/pkg/model/role.go` | 增加 RequireTOTP 字段 |
| `internal/pkg/model/admin.go` | 增加 GoogleKey、GoogleStatus 字段 |
| `pkg/api/apiserver/v1/role.go` | 角色请求/响应增加 RequireTOTP |
| `pkg/api/apiserver/v1/admin.go` | 登录响应增加 TOTP 相关字段，切换角色增加 TOTPCode |
| `internal/admserver/biz/system/login.go` | 两步登录逻辑 |
| `internal/admserver/biz/system/admin.go` | 角色切换 TOTP 验证、重置 TOTP |
| `internal/admserver/biz/system/role.go` | 角色创建/更新支持 RequireTOTP |
| `internal/admserver/handler/http/system/admin.go` | 新增 LoginWithTOTP、ResetTOTP |
| `internal/admserver/router/api.go` | 注册新路由 |
| `internal/pkg/errno/user.go` 或新文件 | 新增 TOTP 错误码（如需要）|

## 依赖关系

```
阶段 1 (数据库) ──→ 执行 migrate reset 验证
    ↓
阶段 2 (API 类型) + 阶段 6 (错误码)
    ↓
阶段 3 (Biz 层)
    ↓
阶段 4 (Handler 层)
    ↓
阶段 5 (路由)
    ↓
阶段 7 (测试)
    ↓
阶段 8 (文档收尾)
```

## 注意事项

1. **复用现有代码**：
   - `internal/pkg/auth/totp.go` 已有 TOTP 核心实现
   - `pkg/api/apiserver/v1/security.go` 已有请求/响应类型
   - `internal/pkg/errno/user.go` 已有 TOTP 错误码

2. **TOTP Token 存储**：两步登录的临时 Token 存 Redis，Key 格式如 `admin:totp_token:{token}`，5 分钟过期

3. **root 用户特殊处理**：
   - root 的虚拟角色不应用 TOTP 策略
   - root 可以重置任何管理员的 TOTP

4. **向后兼容**：现有管理员默认 `google_status=unbind`，现有角色默认 `require_totp=false`，不影响登录

5. **Migration 管理**：遵循 CONVENTIONS.md 7.4，直接修改 create table migration，执行 reset 验证

6. **`/totp/verify` 接口说明**：
   - 这是可选的辅助接口，供前端做"预验证"（如弹窗中输入 TOTP 码后先检查对不对）
   - 真正的安全保障在业务接口内部（如 `switch-role` 内部验证 `totpCode` 参数）
   - 前端不调用 `/totp/verify` 不影响安全性，因为业务接口会强制验证
