# 头脑风暴：用户忘记/重置密码功能设计

**日期**: 2025-12-26
**状态**: 进行中

## 当前理解

### 现有系统能力（已完整实现）

经过代码审查，**密码重置功能已经完整实现**：

| 组件 | 文件 | 状态 |
|------|------|------|
| API 请求结构 | `pkg/api/apiserver/v1/auth.go` | ✅ 完整 |
| 业务逻辑 | `internal/apiserver/biz/auth/reset_password.go` | ✅ 完整 |
| HTTP Handler | `internal/apiserver/handler/http/auth/auth.go:152-179` | ✅ 完整 |
| HTTP 路由 | `internal/apiserver/router/api.go:28` | ✅ 已注册 |
| Swagger 文档 | `api/swagger/apiserver/swagger.json` | ✅ 已生成 |

### 现有 API

**1. 发送验证码**
```
POST /v1/auth/code
{
  "account": "user@example.com",
  "scene": "reset_password"
}
```

**2. 重置密码**
```
POST /v1/auth/reset-password
{
  "account": "user@example.com",
  "code": "123456",
  "password": "newPassword123"
}
```

### 现有流程
1. 用户输入账号（邮箱/手机）
2. 调用 `/v1/auth/code` 发送验证码（场景: `reset_password`）
3. 用户收到验证码
4. 调用 `/v1/auth/reset-password` 验证码 + 新密码
5. 密码更新成功

---

## 关键发现：功能已存在！

Jesse，经过代码审查，**用户忘记/重置密码功能已经完整实现了**。

现有实现包括：
- ✅ 验证码发送（邮件/短信）
- ✅ 验证码校验（6位、5分钟有效、1分钟间隔限制）
- ✅ 账号类型自动检测（邮箱/手机）
- ✅ 密码加密（bcrypt）
- ✅ HTTP API 和路由
- ✅ Swagger 文档

---

## 问题分析：邮件场景区分

### 当前实现问题

当前 `sendEmail` 方法对所有场景使用**相同的邮件内容模板**：

```go
// internal/apiserver/biz/auth/code.go:94-113
func (b *codeBiz) sendEmail(ctx context.Context, email, code string) error {
    // 问题：所有场景使用相同的标题和内容
    subject := "Email Verification code " + code
    msg := fmt.Sprintf("Your verification code is: %s, please note that it will expire in %d minutes.", code, b.codeTTL)
    // ...
}
```

**问题**：
1. 邮件标题/内容不区分场景
2. 用户无法从邮件中识别验证码用途
3. 存在安全风险：用户可能误用验证码

### 现有验证码场景

| 场景 | CodeScene | 用途 |
|------|-----------|------|
| 注册 | `register` | 新用户注册验证 |
| 重置密码 | `reset_password` | 忘记密码时验证 |
| 绑定 | `bind` | 绑定新邮箱/手机时验证 |

### 需要解决的问题

1. **邮件内容需要根据场景定制**
   - 不同场景应有不同的邮件标题
   - 不同场景应有不同的邮件正文
   - 可能需要使用邮件模板

2. **邮件模板管理**
   - 硬编码 vs 模板文件 vs 数据库配置
   - 多语言支持（i18n）
   - 品牌定制（logo、样式）

---

## 设计方案选项

### 方案一：简单场景映射（推荐起步方案）

在 `sendEmail` 中根据 `scene` 参数选择不同的标题和内容：

```go
func (b *codeBiz) sendEmail(ctx context.Context, email, code string, scene CodeScene) error {
    var subject, msg string

    switch scene {
    case CodeSceneRegister:
        subject = "注册验证码"
        msg = fmt.Sprintf("您正在注册账号，验证码：%s，%d分钟内有效。", code, b.codeTTL)
    case CodeSceneResetPassword:
        subject = "密码重置验证码"
        msg = fmt.Sprintf("您正在重置密码，验证码：%s，%d分钟内有效。如非本人操作请忽略。", code, b.codeTTL)
    case CodeSceneBind:
        subject = "绑定验证码"
        msg = fmt.Sprintf("您正在绑定邮箱，验证码：%s，%d分钟内有效。", code, b.codeTTL)
    }
    // ...
}
```

**优点**：简单、快速实现
**缺点**：不支持多语言、修改需要改代码

### 方案二：邮件模板系统

使用 Go 模板引擎加载模板文件：

```
templates/
  emails/
    register.html
    reset_password.html
    bind.html
```

**优点**：灵活、支持 HTML 邮件、易于定制
**缺点**：需要额外的模板管理

### 方案三：数据库配置模板

将邮件模板存储在数据库中，支持后台管理：

**优点**：动态配置、无需重启
**缺点**：增加复杂度

---

## 确定的设计方向

### 用户决策

| 问题 | 决策 |
|------|------|
| 内容管理方式 | 简单场景映射（switch/case） |
| 多语言支持 | 需要，同步引入 i18n |
| 邮件格式 | 纯文本 |
| 短信场景区分 | 需要 |
| i18n 库 | **nicksnyder/go-i18n/v2** |
| 翻译文件格式 | YAML |
| 语言偏好来源 | Accept-Language Header |
| 默认语言 | 英文 (en) |

### i18n 库对比

| 指标 | go-i18n ✅ | universal-translator |
|------|---------|----------------------|
| Stars | 3,429 | 417 |
| 最近更新 | 24 天前 | 3 年前 |
| 活跃度 | 增长中 | 停滞 |
| YAML 支持 | ✅ | ❌ |
| CLI 工具 | ✅ goi18n | ❌ |

选择 **go-i18n**：活跃维护、支持 YAML、有 CLI 工具

---

## 设计方案（最终）

### 1. 新增依赖

```bash
go get github.com/nicksnyder/go-i18n/v2
```

### 2. 文件结构

```
internal/pkg/i18n/
├── i18n.go           # 初始化和翻译函数
└── locales/
    ├── en.yaml       # 英文（默认）
    └── zh.yaml       # 中文
```

### 3. 翻译文件内容

**en.yaml**:
```yaml
# Verification code messages
code_register_subject:
  other: "Registration Verification Code"
code_register_body:
  other: "You are registering an account. Your verification code is: {{.Code}}. Valid for {{.TTL}} minutes."

code_reset_password_subject:
  other: "Password Reset Verification Code"
code_reset_password_body:
  other: "You are resetting your password. Your verification code is: {{.Code}}. Valid for {{.TTL}} minutes. Ignore this if you didn't request it."

code_bind_subject:
  other: "Account Binding Verification Code"
code_bind_body:
  other: "You are binding your account. Your verification code is: {{.Code}}. Valid for {{.TTL}} minutes."
```

**zh.yaml**:
```yaml
# 验证码消息
code_register_subject:
  other: "注册验证码"
code_register_body:
  other: "您正在注册账号，验证码：{{.Code}}，{{.TTL}}分钟内有效。"

code_reset_password_subject:
  other: "密码重置验证码"
code_reset_password_body:
  other: "您正在重置密码，验证码：{{.Code}}，{{.TTL}}分钟内有效。如非本人操作请忽略。"

code_bind_subject:
  other: "账号绑定验证码"
code_bind_body:
  other: "您正在绑定账号，验证码：{{.Code}}，{{.TTL}}分钟内有效。"
```

### 4. i18n 包实现

```go
// internal/pkg/i18n/i18n.go
package i18n

import (
    "embed"
    "github.com/nicksnyder/go-i18n/v2/i18n"
    "golang.org/x/text/language"
    "gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var localeFS embed.FS

var bundle *i18n.Bundle

func Init() {
    bundle = i18n.NewBundle(language.English)
    bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
    bundle.LoadMessageFileFS(localeFS, "locales/en.yaml")
    bundle.LoadMessageFileFS(localeFS, "locales/zh.yaml")
}

func T(lang, messageID string, data map[string]interface{}) string {
    localizer := i18n.NewLocalizer(bundle, lang)
    msg, _ := localizer.Localize(&i18n.LocalizeConfig{
        MessageID:    messageID,
        TemplateData: data,
    })
    return msg
}
```

### 5. 修改 codeBiz

```go
// internal/apiserver/biz/auth/code.go

func (b *codeBiz) Send(ctx context.Context, account string, scene CodeScene) error {
    // ... 现有逻辑 ...

    switch accountType {
    case AccountTypeEmail:
        return b.sendEmail(ctx, account, code, scene)  // 添加 scene
    case AccountTypePhone:
        return b.sendSMS(ctx, account, code, scene)    // 添加 scene
    }
}

func (b *codeBiz) sendEmail(ctx context.Context, email, code string, scene CodeScene) error {
    lang := contextx.Lang(ctx)  // 从 context 获取语言

    subject := i18n.T(lang, string(scene)+"_subject", nil)
    body := i18n.T(lang, string(scene)+"_body", map[string]interface{}{
        "Code": code,
        "TTL":  b.codeTTL,
    })

    payload := &task.EmailVerificationCodePayload{
        To:      email,
        Subject: subject,
        Content: body,
    }
    // ... 发送 ...
}

func (b *codeBiz) sendSMS(ctx context.Context, phone, code string, scene CodeScene) error {
    lang := contextx.Lang(ctx)
    body := i18n.T(lang, string(scene)+"_body", map[string]interface{}{
        "Code": code,
        "TTL":  b.codeTTL,
    })
    // ... 发送短信 ...
}
```

### 6. 语言中间件

```go
// internal/pkg/middleware/lang.go
func LangMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        lang := c.GetHeader("Accept-Language")
        if lang == "" {
            lang = "en"  // 默认英文
        }
        ctx := contextx.WithLang(c.Request.Context(), lang)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

---

## 实施步骤

1. **添加 go-i18n 依赖**
2. **创建 i18n 包和翻译文件**
3. **添加语言中间件**
4. **修改 codeBiz.sendEmail/sendSMS 使用 i18n**
5. **在应用启动时初始化 i18n**
6. **编写测试**

---

## 探索进度

- [x] 了解现有认证系统架构
- [x] 确认现有实现的完整性
- [x] 识别问题：邮件内容不区分场景
- [x] 确定解决方案方向
- [x] 确定 i18n 库选择：go-i18n
- [x] 制定详细实施方案

## 头脑风暴完成 ✅

设计方案已确定，可以进入实施阶段。
