# 智能体预设 (AI Agents)

智能体预设 (AI Agents) 允许用户快速切换不同的 AI 助手人格（如"资深翻译"、"Python 专家"、"心理咨询师"）。每个智能体都有独立的 System Prompt 和模型参数配置。

## 1. 智能体模型

### 1.1 核心字段

AI 智能体存储在 `ai_agent` 表中，主要字段包括：

| 字段 | 说明 | 示例 |
|------|------|------|
| `agent_id` | 唯一标识符，用于 API 调用 | `math_teacher` |
| `name` | 显示名称 | `数学老师` |
| `system_prompt` | 核心提示词，定义智能体行为 | `你是一位小学数学老师...` |
| `model` | (可选) 强制绑定的模型 | `gpt-4o` (若为空则使用用户选定的模型) |
| `temperature` | (可选) 思维发散度 | `0.7` |
| `category` | 分类标签 | `education`, `coding`, `general` |
| `status` | 状态 | `active`, `disabled` |

### 1.2 参数覆盖逻辑

当用户指定 `agent_id` 发起对话时，系统按以下优先级合并参数：

1. **Agent 强制指定**: 如果智能体配置了 `model`, `temperature`, `max_tokens`，则**覆盖**用户请求中的对应参数。
2. **用户请求参数**: 如果智能体未配置，则使用用户请求中的参数。
3. **系统默认值**: 如果用户也未指定，则使用系统默认配置。

并且，智能体的 `system_prompt` 会作为 Messages 列表的第一条自动插入。

## 2. 权限隔离

为了保证系统安全，我们对智能体的**管理**和**使用**进行了严格的权限隔离。

### 2.1 管理端 (Admin)

- **入口**: `bingo-admserver`
- **权限**: 仅管理员可访问。
- **能力**: 完整的 CRUD 操作。可以创建、更新、软删除智能体，可以查看任意状态（包括 disabled）的智能体。
- **接口**:
  - `POST /v1/ai/agents`: 创建智能体
  - `PUT /v1/ai/agents/:id`: 更新智能体
  - `DELETE /v1/ai/agents/:id`: 删除智能体

### 2.2 用户端 (User)

- **入口**: `bingo-apiserver`
- **权限**: 所有已登录用户。
- **能力**: **只读**。只能查询状态为 `active` 的智能体。
- **接口**:
  - `GET /v1/ai/agents`: 获取智能体列表 (强制过滤 `status=active`)
  - `GET /v1/ai/agents/:id`: 获取智能体详情
  - `POST /v1/chat/completions`: 调用对话接口时指定 `agent_id`

## 3. 请求示例

### 3.1 查询可用智能体

```bash
GET /v1/ai/agents?category=coding
```

响应:
```json
{
  "data": [
    {
      "agentId": "python_expert",
      "name": "Python 专家",
      "description": "精通 Python 编程和性能优化",
      "category": "coding",
      "model": "gpt-4o"
    }
  ]
}
```

### 3.2 使用智能体对话

```bash
POST /v1/chat/completions
{
  "agentId": "python_expert",
  "messages": [
    {"role": "user", "content": "如何优化列表推导式？"}
  ]
}
```

系统实际发送给 LLM 的请求（自动注入 System Prompt）:

```json
{
  "model": "gpt-4o",
  "messages": [
    {"role": "system", "content": "你是一位拥有10年经验的 Python 专家..."},
    {"role": "user", "content": "如何优化列表推导式？"}
  ]
}
```
