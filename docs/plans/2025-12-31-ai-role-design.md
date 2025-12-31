# AI 角色预设功能设计

## 概述

为 AI 对话系统增加角色预设能力，支持用户快速切换不同的 AI 助手角色（如老师、医生、HR 等），每个角色有独立的系统提示词和行为配置。

**设计日期**: 2025-12-31

---

## 设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 角色存储方式 | 数据库独立表 | 支持动态管理，前端可查询可用角色 |
| 角色方案 | 多角色独立定制 | 真实度、专业度更高，符合 Prompt Engineering 最佳实践 |
| 变量支持 | 不支持模板变量 | 保持简单，5-10 个角色数量不大 |
| 分类管理 | category 字段 | 支持前端分组展示 |
| 调用方式 | role_id 参数 | 扩展现有 API，向后兼容 |

---

## 数据库设计

### ai_role 表

```sql
CREATE TABLE ai_role (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    role_id VARCHAR(32) NOT NULL COMMENT '外部角色ID',
    name VARCHAR(64) NOT NULL COMMENT '角色名称',
    description VARCHAR(255) COMMENT '角色描述',
    icon VARCHAR(255) COMMENT '角色图标URL',
    category VARCHAR(32) DEFAULT 'general' COMMENT '分类',
    system_prompt TEXT NOT NULL COMMENT '系统提示词',
    model VARCHAR(64) COMMENT '指定模型，NULL用系统默认',
    temperature DECIMAL(3,2) DEFAULT 0.70 COMMENT '温度参数(0.00-1.00)',
    max_tokens INT DEFAULT 2000 COMMENT '最大输出token数',
    sort INT DEFAULT 0 COMMENT '排序权重',
    status VARCHAR(16) DEFAULT 'active' COMMENT '状态',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_role_id (role_id),
    KEY idx_category_status (category, status),
    KEY idx_status_sort (status, sort)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI 角色预设';
```

### 状态常量

```go
const (
    AiRoleStatusActive   = "active"
    AiRoleStatusDisabled = "disabled"
)
```

### 分类常量（可选）

```go
const (
    AiRoleCategoryGeneral  = "general"   // 通用
    AiRoleCategoryEducation = "education" // 教育
    AiRoleCategoryMedical   = "medical"   // 医疗
    AiRoleCategoryWorkplace = "workplace" // 职场
    AiRoleCategoryCreative  = "creative"  // 创作
)
```

---

## API 设计

### 角色管理接口

| 方法 | 路径 | 说明 | 权限 |
|------|------|------|------|
| GET | `/v1/ai/roles` | 获取角色列表 | 公开 |
| GET | `/v1/ai/roles/:id` | 获取角色详情 | 公开 |
| POST | `/v1/ai/roles` | 创建角色 | 管理员 |
| PUT | `/v1/ai/roles/:id` | 更新角色 | 管理员 |
| DELETE | `/v1/ai/roles/:id` | 删除角色 | 管理员 |

### 获取角色列表

**GET /v1/ai/roles**

Query 参数:
- `category` (可选): 按分类筛选
- `status` (可选): 状态筛选，默认 active

Response:
```json
{
  "data": [
    {
      "role_id": "math_teacher",
      "name": "数学老师",
      "description": "擅长小学数学辅导",
      "icon": "https://...",
      "category": "education",
      "model": "gpt-4o",
      "sort": 1
    }
  ],
  "total": 4
}
```

### 获取角色详情

**GET /v1/ai/roles/:id**

Response:
```json
{
  "role_id": "math_teacher",
  "name": "数学老师",
  "description": "擅长小学数学辅导，耐心引导",
  "icon": "https://...",
  "category": "education",
  "system_prompt": "你是一位经验丰富的小学数学老师...",
  "model": "gpt-4o",
  "temperature": 0.7,
  "max_tokens": 2000,
  "sort": 1,
  "status": "active"
}
```

### 聊天接口扩展

**POST /v1/chat/completions**

新增字段:
```json
{
  "model": "gpt-4o",
  "messages": [{"role": "user", "content": "什么是分数？"}],
  "role_id": "math_teacher",  // 新增：指定角色
  "session_id": "sess_xxx"
}
```

**处理逻辑**:
1. 如果 `role_id` 为空，按原逻辑处理
2. 如果 `role_id` 有值：
   - 查询角色配置
   - 检查状态是否为 active
   - 将 `system_prompt` 作为第一条消息插入
   - 如果角色指定了 `model`，覆盖请求中的 model

---

## 业务逻辑设计

### 消息构建流程

```go
// internal/apiserver/biz/chat/chat.go

func (b *chatBiz) buildMessagesWithRole(ctx context.Context, req *ChatRequest) ([]ai.Message, error) {
    // 1. 如果指定了 role_id，加载角色预设
    if req.RoleID != "" {
        role, err := b.ds.AiRole().GetByRoleID(ctx, req.RoleID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil, errno.ErrAIRoleNotFound
            }
            return nil, errno.ErrOperationFailed.WithMessage("failed to get role: %v", err)
        }

        // 2. 检查角色状态
        if role.Status != model.AiRoleStatusActive {
            return nil, errno.ErrAIRoleDisabled
        }

        // 3. 构建 messages: system prompt + user messages
        messages := []ai.Message{
            {Role: ai.RoleSystem, Content: role.SystemPrompt},
        }
        messages = append(messages, req.Messages...)

        // 4. 如果角色指定了模型，覆盖请求模型
        if role.Model != "" {
            req.Model = role.Model
        }
        // 5. 如果角色指定了 temperature/max_tokens，覆盖请求参数
        if role.Temperature > 0 {
            req.Temperature = role.Temperature
        }
        if role.MaxTokens > 0 {
            req.MaxTokens = role.MaxTokens
        }

        return messages, nil
    }

    // 6. 没有指定角色，直接使用请求的 messages
    return req.Messages, nil
}
```

### Chat 方法修改

```go
func (b *chatBiz) Chat(ctx context.Context, uid string, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    if len(req.Messages) == 0 {
        return nil, errno.ErrAIEmptyMessages
    }

    // 新增：处理角色预设
    messages, err := b.buildMessagesWithRole(ctx, req)
    if err != nil {
        return nil, err
    }
    req.Messages = messages

    // ... 后续逻辑不变
}
```

---

## 文件改动清单

### 新增文件

| 文件 | 说明 |
|------|------|
| **internal/pkg/model/** | |
| `ai_role.go` | AiRoleM Model 定义 |
| **internal/pkg/store/** | |
| `ai_role.go` | AiRoleStore 接口和实现 |
| **internal/apiserver/biz/role/** | |
| `role.go` | RoleBiz 业务逻辑 |
| **internal/apiserver/handler/http/role/** | |
| `role.go` | Role HTTP Handler |
| **pkg/api/apiserver/v1/** | |
| `role.go` | Role API DTO |
| **internal/pkg/errno/** | |
| `ai.go` (修改) | 新增角色相关错误码 |
| **数据库迁移** | |
| `xxx_create_ai_role_table.go` | ai_role 表迁移 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `internal/apiserver/biz/chat/chat.go` | 新增 `buildMessagesWithRole` 方法 |
| `pkg/api/apiserver/v1/chat.go` | ChatRequest 新增 `role_id` 字段 |
| `internal/apiserver/router/` | 新增角色相关路由 |

### 新增错误码

```go
// internal/pkg/errno/ai.go
var (
    ErrAIRoleNotFound    = NewError(20001, "AI role not found")
    ErrAIRoleDisabled    = NewError(20002, "AI role is disabled")
)
```

---

## 示例数据

### 教育分类

```sql
INSERT INTO ai_role (role_id, name, description, category, system_prompt, model, temperature, max_tokens, sort) VALUES
('math_teacher', '数学老师', '擅长小学数学辅导，耐心引导', 'education',
'你是一位经验丰富的小学数学老师，擅长用简单易懂的语言解释概念。请耐心引导学生思考，不要直接给出答案，而是通过提问帮助学生自己找到答案。', 'gpt-4o', 0.7, 2000, 1),

('chinese_teacher', '语文老师', '擅长阅读写作指导', 'education',
'你是一位小学语文老师，擅长指导阅读理解和写作。请用温和鼓励的方式，帮助学生提高语文能力。', 'gpt-4o', 0.7, 2000, 2),

('english_teacher', '英语老师', '擅长英语口语和语法', 'education',
'你是一位英语外教，擅长纠正语法错误和表达方式。请用自然的英语与学生对话，必要时给出中文解释和示例。', 'gpt-4o', 0.7, 2000, 3);
```

### 医疗分类

```sql
INSERT INTO ai_role (role_id, name, description, category, system_prompt, model, temperature, max_tokens, sort) VALUES
('doctor_internal', '内科医生', '擅长内科疾病诊断建议', 'medical',
'你是一位内科医生，擅长消化、呼吸、心血管等常见疾病的诊断和治疗建议。请注意：你只能提供参考建议，不能替代线下就医。对于紧急情况，请立即建议患者就医。', 'gpt-4o', 0.6, 1500, 10),

('doctor_surgical', '外科医生', '擅长外科疾病诊断建议', 'medical',
'你是一位外科医生，擅长常见外科疾病的诊断和治疗建议。请注意：你只能提供参考建议，不能替代线下就医。对于紧急情况，请立即建议患者就医。', 'gpt-4o', 0.6, 1500, 11),

('doctor_pediatric', '儿科医生', '擅长儿童疾病诊断建议', 'medical',
'你是一位儿科医生，擅长儿童常见病的诊断和治疗建议。请注意：你只能提供参考建议，不能替代线下就医。对于紧急情况，请立即建议家长带孩子就医。', 'gpt-4o', 0.6, 1500, 12);
```

### 职场分类

```sql
INSERT INTO ai_role (role_id, name, description, category, system_prompt, model, temperature, max_tokens, sort) VALUES
('tech_hr', '科技HR', '科技行业招聘专家', 'workplace',
'你是一位科技行业HR，熟悉互联网、软件开发岗位。面试时关注：1. 技术栈匹配度（如 Go、Python、前端框架）2. 开源贡献和 GitHub 活动 3. 技术博客和社区参与 4. 敏捷开发经验。请用专业但亲切的语气交流。', 'gpt-4o', 0.7, 2000, 20),

('finance_hr', '金融HR', '金融行业招聘专家', 'workplace',
'你是一位金融行业HR，熟悉银行、证券、基金等机构。面试时关注：1. 持有证书（CPA、CFA、FRM）2. 合规意识和风控理解 3. 对金融产品的了解 4. 工作稳定性。请用严谨、专业的语气交流。', 'gpt-4o', 0.7, 2000, 21),

('code_reviewer', '代码审查', '代码质量专家', 'workplace',
'你是一位严谨的代码审查专家，专注于发现代码中的 bug、安全隐患、性能问题和代码规范问题。请给出具体的改进建议，并解释原因。', 'gpt-4o', 0.3, 2000, 22);
```

### 通用分类

```sql
INSERT INTO ai_role (role_id, name, description, category, system_prompt, model, temperature, max_tokens, sort) VALUES
('creative_writer', '创作助手', '帮助构思和润色', 'general',
'你是一位富有创意的写作助手，擅长帮助用户构思故事情节、润色文字、提供创作灵感。请保持开放和鼓励的态度，提供建设性的建议。', 'gpt-4o', 0.8, 2000, 30),

('interview_coach', '面试教练', '模拟面试和技巧指导', 'general',
'你是一位专业的面试教练，擅长帮助用户准备面试。你可以进行模拟面试、提供面试技巧、分析常见问题的回答策略。请以鼓励为主，给出具体可操作的建议。', 'gpt-4o', 0.7, 2000, 31);
```

---

## 前端集成

### 角色选择器

```
┌─────────────────────────────────────────┐
│  选择角色                               │
├─────────────────────────────────────────┤
│  📚 教育                                 │
│    🎓 数学老师    👩🏫 语文老师         │
│    🌐 英语老师                           │
├─────────────────────────────────────────┤
│  🏥 医疗                                 │
│    👨⚕️ 内科医生    👩⚕️ 儿科医生        │
├─────────────────────────────────────────┤
│  💼 职场                                 │
│    👔 科技HR      🔍 代码审查           │
└─────────────────────────────────────────┘
```

### 调用示例

```javascript
// 不使用角色
const response = await fetch('/v1/chat/completions', {
  method: 'POST',
  body: JSON.stringify({
    model: 'gpt-4o',
    messages: [{role: 'user', content: '什么是分数？'}]
  })
});

// 使用角色
const response = await fetch('/v1/chat/completions', {
  method: 'POST',
  body: JSON.stringify({
    role_id: 'math_teacher',  // 指定角色
    messages: [{role: 'user', content: '什么是分数？'}]
  })
});
```

---

## 实现检查清单

- [ ] 创建 `internal/pkg/model/ai_role.go`
- [ ] 创建 `internal/pkg/store/ai_role.go`
- [ ] 创建 `internal/apiserver/biz/role/role.go`
- [ ] 创建 `internal/apiserver/handler/http/role/role.go`
- [ ] 创建 `pkg/api/apiserver/v1/role.go`
- [ ] 创建数据库迁移文件
- [ ] 修改 `internal/apiserver/biz/chat/chat.go` 添加 `buildMessagesWithRole`
- [ ] 修改 `pkg/api/apiserver/v1/chat.go` 添加 `role_id` 字段
- [ ] 添加路由注册
- [ ] 执行数据库迁移
- [ ] 插入示例数据
- [ ] API 测试

---

## 后续扩展（暂不实现）

| 功能 | 说明 | 优先级 |
|------|------|--------|
| 用户自定义角色 | 允许用户创建私人角色 | P2 |
| 角色统计 | 记录每个角色的使用次数 | P2 |
| 角色评分 | 用户对角色质量评分 | P3 |
| 多模态角色 | 支持图像分析角色 | P3 |

---

## 参考文档

- [AI 对话功能设计](./2025-12-29-ai-chat-design.md)
- [AI 对话模块 Review](./2025-12-31-ai-chat-review.md)
