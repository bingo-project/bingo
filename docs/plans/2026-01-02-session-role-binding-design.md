# Session 绑定角色设计方案

**目标**: 将 AI 角色与 Session 绑定,而不是在每次聊天请求时指定角色。

**日期**: 2026-01-02
**状态**: 设计完成,待实施

---

## 1. 背景和问题

### 当前实现
- 聊天请求中通过 `role_id` 参数指定角色
- 每次聊天都需要传递角色 ID
- Session 与角色无关联

### 问题分析
1. **用户体验**: 每次聊天都指定角色不符合直觉
2. **语义混乱**: 同一会话可能混用多个角色,历史记录不清晰
3. **配置冲突**: Session 的模型配置与角色模型可能不一致

### 方案选择

经过分析,**一个会话中切换角色的场景极少**,原因是:
- 上下文污染: 不同角色的系统提示词会相互干扰
- 历史混乱: 混合角色的记录难以理解
- 用户习惯: 更自然的做法是"换话题=创建新会话"

**决定**: Session 在创建时绑定角色,聊天时自动应用。

---

## 2. 核心设计

### 2.1 设计原则

1. **Session 创建时绑定 Role ID** - 作为会话的默认角色
2. **聊天时自动使用 Session 的角色** - 无需在请求中指定
3. **移除聊天请求中的 `role_id` 参数** - 简化 API
4. **向后兼容** - 现有无角色的 Session 继续工作

### 2.2 数据模型变更

```go
// internal/pkg/model/ai_session.go

type AiSessionM struct {
    ID           uint   `gorm:"primaryKey" json:"id"`
    SessionID    string `gorm:"column:session_id;type:varchar(64);uniqueIndex:uk_session_id;not null" json:"sessionId"`
    UID          string `gorm:"column:uid;type:varchar(64);index:idx_uid;not null" json:"uid"`
    RoleID       string `gorm:"column:role_id;type:varchar(64);index:idx_role_id" json:"roleId"` // 新增
    Title        string `gorm:"column:title;type:varchar(255);not null;default:''" json:"title"`
    Model        string `gorm:"column:model;type:varchar(64);not null" json:"model"`
    MessageCount int    `gorm:"column:message_count;type:int;not null;default:0" json:"messageCount"`
    TotalTokens  int    `gorm:"column:total_tokens;type:int;not null;default:0" json:"totalTokens"`
    Status       string `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`

    CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
    UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}
```

### 2.3 API 变更

#### 创建 Session 请求

```go
// pkg/api/apiserver/v1/chat.go

type CreateSessionRequest struct {
    RoleID string `json:"roleId"`            // 新增: 绑定的角色ID
    Title  string `json:"title,omitempty"`   // 可选, 默认为角色名称或"新对话"
    Model  string `json:"model,omitempty"`  // 可选: 覆盖角色默认模型
}
```

#### Session 信息响应

```go
type SessionInfo struct {
    SessionID    string    `json:"sessionId"`
    Title        string    `json:"title"`
    RoleID       string    `json:"roleId"`       // 新增
    RoleName     string    `json:"roleName"`     // 新增: 关联查询角色名称
    Model        string    `json:"model"`
    MessageCount int       `json:"messageCount"`
    TotalTokens  int       `json:"totalTokens"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"createdAt"`
    UpdatedAt    time.Time `json:"updatedAt"`
}
```

#### 聊天请求变更

```go
// 移除 role_id 字段
type ChatCompletionRequest struct {
    Model       string        `json:"model" binding:"required"`
    Messages    []ChatMessage `json:"messages" binding:"required,min=1"`
    MaxTokens   int           `json:"maxTokens,omitempty"`
    Temperature float64       `json:"temperature,omitempty"`
    Stream      bool          `json:"stream,omitempty"`
    SessionID   string        `json:"sessionId"` // 从 session 获取 role_id
    // 移除: RoleID string `json:"role_id,omitempty"`
}
```

---

## 3. 业务逻辑实现

### 3.1 创建 Session 流程

```go
// internal/apiserver/biz/chat/session.go

func (b *sessionBiz) Create(ctx context.Context, uid, title, roleID string) (*model.AiSessionM, error) {
    var model string

    // 如果指定了 role_id, 从角色获取默认配置
    if roleID != "" {
        role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
        if err != nil {
            return nil, err
        }
        if role.Status == model.AiRoleStatusDisabled {
            return nil, errno.ErrAIRoleDisabled
        }
        model = role.Model  // 使用角色的模型
        if title == "" {
            title = role.Name  // 默认标题使用角色名称
        }
    }

    // 回退到默认模型
    if model == "" {
        model = facade.Config.AI.DefaultModel
    }
    if title == "" {
        title = "新对话"
    }

    // 创建 Session
    session := &model.AiSessionM{
        SessionID: generateSessionID(),
        UID:       uid,
        Title:     title,
        RoleID:    roleID,  // 绑定角色
        Model:     model,
        Status:    model.AiSessionStatusActive,
    }

    return b.ds.AiSession().Create(ctx, session)
}
```

### 3.2 聊天流程变更

```go
// internal/apiserver/biz/chat/chat.go

func (b *chatBiz) Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    // ... 前置验证

    // 【新增】从 Session 获取 RoleID
    if req.SessionID != "" {
        session, err := b.ds.AiSession().GetBySessionID(ctx, req.SessionID)
        if err == nil && session.RoleID != "" {
            req.RoleID = session.RoleID  // 使用 Session 绑定的角色
        }
    }

    // 【修改】buildMessagesWithRole 现在使用 Session 的 RoleID
    if err := b.buildMessagesWithRole(ctx, req); err != nil {
        return nil, err
    }

    // ... 后续逻辑保持不变
}
```

### 3.3 角色提示词加载(保持不变)

```go
// buildMessagesWithRole 从 Session 的 RoleID 加载角色提示词
func (b *chatBiz) buildMessagesWithRole(ctx context.Context, req *ai.ChatRequest) error {
    if req.RoleID == "" {
        return nil  // 没有角色, 不添加系统提示词
    }

    // 获取角色详情
    role, err := b.ds.AiRole().GetByRoleID(ctx, req.RoleID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return errno.ErrAIRoleNotFound
        }
        return errno.ErrDBRead.WithMessage("get ai role: %v", err)
    }

    if role.Status == model.AiRoleStatusDisabled {
        return errno.ErrAIRoleDisabled
    }

    // 应用角色配置
    if req.Model == "" || req.Model == facade.Config.AI.DefaultModel {
        if role.Model != "" {
            req.Model = role.Model
        }
    }
    if req.Temperature == 0 && role.Temperature > 0 {
        req.Temperature = role.Temperature
    }
    if req.MaxTokens == 0 && role.MaxTokens > 0 {
        req.MaxTokens = role.MaxTokens
    }

    // 注入系统提示词
    hasSystem := false
    if len(req.Messages) > 0 && req.Messages[0].Role == ai.RoleSystem {
        hasSystem = true
    }

    if !hasSystem && role.SystemPrompt != "" {
        systemMsg := ai.Message{
            Role:    ai.RoleSystem,
            Content: role.SystemPrompt,
        }
        req.Messages = append([]ai.Message{systemMsg}, req.Messages...)
    }

    return nil
}
```

---

## 4. 数据库迁移

### 4.1 Migration 文件

```go
// internal/pkg/database/migration/2026_01_02_100000_add_role_id_to_ai_session.go
// ABOUTME: Database migration for adding role_id to ai_session table.
// ABOUTME: Adds foreign key to ai_role for session-role binding.

package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type AddRoleIDToAISession struct {
	ID           uint64    `gorm:"primaryKey"`
	SessionID    string    `gorm:"type:varchar(64);uniqueIndex:uk_session_id;not null"`
	UID          string    `gorm:"type:varchar(64);index:idx_uid;not null"`
	RoleID       string    `gorm:"type:varchar(64);index:idx_role_id"`
	Title        string    `gorm:"type:varchar(255);not null;default:''"`
	Model        string    `gorm:"type:varchar(64);not null"`
	MessageCount int       `gorm:"type:int;not null;default:0"`
	TotalTokens  int       `gorm:"type:int;not null;default:0"`
	Status       string    `gorm:"type:varchar(16);not null;default:active"`
	CreatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt    time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (AddRoleIDToAISession) TableName() string {
	return "ai_session"
}

func (AddRoleIDToAISession) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&AddRoleIDToAISession{})
}

func (AddRoleIDToAISession) Down(migrator gorm.Migrator) {
	_ = migrator.DB().Exec("ALTER TABLE ai_session DROP COLUMN role_id")
	_ = migrator.DB().Exec("ALTER TABLE ai_session DROP INDEX idx_role_id")
}

func init() {
	migrate.Add("2026_01_02_100000_add_role_id_to_ai_session", AddRoleIDToAISession{}.Up, AddRoleIDToAISession{}.Down)
}
```

### 4.2 同步更新原表创建 Migration

根据 CONVENTIONS.md 第 7.4 节,需要同步更新原表的创建 migration:

```go
// internal/pkg/database/migration/2025_12_29_100004_create_ai_session_table.go

type CreateAISessionTable struct {
    // ... 原有字段
    RoleID string `gorm:"type:varchar(64);index:idx_role_id"` // 新增
    // ... 其他字段
}
```

---

## 5. 实施计划

### 5.1 实施步骤

#### 第 1 步: 数据库迁移
- [ ] 创建 `2026_01_02_100000_add_role_id_to_ai_session.go`
- [ ] 更新 `AiSessionM` 模型添加 `RoleID` 字段
- [ ] 更新原表创建 migration
- [ ] 执行 `bingo migrate up` 验证
- [ ] 执行 `bingo migrate rollback` 测试回滚

#### 第 2 步: API 结构更新
- [ ] 更新 `CreateSessionRequest` 添加 `roleId` 字段
- [ ] 更新 `SessionInfo` 添加 `roleId` 和 `roleName` 字段
- [ ] 移除 `ChatCompletionRequest.RoleID` 字段
- [ ] 执行 `make swag` 生成 Swagger 文档

#### 第 3 步: 业务逻辑实现
- [ ] 修改 `sessionBiz.Create` 支持 `role_id` 参数
- [ ] 修改 `chatBiz.Chat` 从 Session 获取 `role_id`
- [ ] 修改 `chatBiz.ChatStream` 从 Session 获取 `role_id`
- [ ] 更新 `buildMessagesWithRole` 逻辑

#### 第 4 步: Handler 层适配
- [ ] 更新 `SessionHandler.Create` 处理 `role_id`
- [ ] 更新 `SessionHandler.List` 关联查询角色信息
- [ ] 更新 `SessionHandler.Get` 返回角色信息

#### 第 5 步: 测试验证
- [ ] 单元测试: Session 创建、角色绑定
- [ ] 集成测试: 聊天时自动应用角色提示词
- [ ] 向后兼容性测试: 无 role_id 的 Session

### 5.2 验证清单

```bash
# 1. 数据库迁移
bingo migrate up
✓ 检查 ai_session 表是否有 role_id 字段
✓ 检查索引 idx_role_id 是否创建

# 2. 构建验证
make build
✓ 编译无错误

# 3. Swagger 文档
make swag
✓ API 文档已更新

# 4. 功能测试
# 创建带角色的 Session
POST /v1/ai/sessions
{
  "roleId": "math_teacher"
}
✓ Session 创建成功, title 为角色名称
✓ model 为角色配置的模型

# 聊天测试
POST /v1/chat/completions
{
  "sessionId": "xxx",
  "messages": [{"role": "user", "content": "1+1=?"}]
}
✓ 自动使用 Session 的角色
✓ 系统提示词已注入
```

---

## 6. 向后兼容性

### 6.1 数据层面
- ✅ 现有 Session (role_id=NULL): 正常工作,无系统提示词
- ✅ 新字段允许 NULL: 不影响现有数据

### 6.2 API 层面
- ✅ 不指定 role_id 创建 Session: 使用默认模型,标题为"新对话"
- ⚠️ 移除 `ChatCompletionRequest.RoleID`: 破坏性变更
  - **选项 A**: 直接移除(推荐,如果前端可以同步更新)
  - **选项 B**: 标记为 `deprecated`,保留一段时间

### 6.3 行为层面
- ✅ 无角色 Session: 聊天时无系统提示词,行为与之前一致
- ✅ 有角色 Session: 自动应用角色提示词,新功能

---

## 7. 风险和注意事项

### 7.1 风险点
1. **破坏性变更**: 移除聊天请求的 `role_id` 字段
   - 缓解: 前端同步更新,或保留字段标记废弃
2. **现有会话**: 用户已有的无角色 Session
   - 缓解: 允许 role_id 为 NULL,不影响现有功能
3. **性能**: 关联查询角色信息可能影响性能
   - 缓解: 添加索引,使用缓存

### 7.2 测试重点
1. **角色配置正确应用**: 模型、温度、最大 tokens
2. **系统提示词注入**: 验证提示词在消息开头
3. **历史消息加载**: 验证系统提示词在滑动窗口中保留
4. **向后兼容**: 无角色 Session 正常工作

---

## 8. 总结

### 8.1 优势
- ✅ **语义清晰**: Session 与角色绑定,会话用途明确
- ✅ **用户友好**: 创建时选角色,聊天时无需关心
- ✅ **代码简化**: 聊天接口参数减少
- ✅ **配置统一**: Session 和角色模型配置一致

### 8.2 劣势
- ❌ **灵活性降低**: 无法中途切换角色
- ❌ **破坏性变更**: 需要前端同步更新

### 8.3 决定
**采用此方案**,理由:
1. 符合用户使用习惯("一个会话 = 一个场景")
2. 避免上下文混乱
3. 简化 API 和前端逻辑
4. 灵活性损失可接受(用户可以创建新会话)

---

**文档状态**: ✅ 设计完成,待实施
