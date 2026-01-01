# AI 核心模块

Bingo 框架内置了强大的 AI 对话模块，提供开箱即用的多厂商支持、流式响应、角色预设和高可用保障。

## 🌟 核心能力

- **多厂商支持**：统一接口对接 OpenAI, Claude, Gemini, Qwen (通义千问) 等主流模型。
- **流式响应 (SSE)**：支持打字机效果的实时流式输出，提升用户体验。
- **动态配置**：无需重启服务，即可通过数据库动态开启/禁用模型或调整优先级。
- **角色预设 (Personas)**：支持创建不同的 AI 角色（如翻译官、代码助手），定制 System Prompt 和参数。
- **高可用机制**：内置自动重试（瞬时错误）和故障熔断保护。
- **状态管理**：智能滑动窗口管理历史记录，自动持久化会话。

## 🚀 快速开始

### 1. 基础配置

在 `configs/bingo-apiserver.yaml` 中配置必要的 API Key（出于安全考虑，敏感信息仅通过文件或环境变量配置）：

```yaml
ai:
  credentials:
    openai:
      api_key: "${OPENAI_API_KEY}"
    claude:
      api_key: "${CLAUDE_API_KEY}"
    gemini:
      api_key: "${GEMINI_API_KEY}"
    qwen:
      api_key: "${QWEN_API_KEY}"
      base_url: "https://dashscope.aliyuncs.com/compatible-mode/v1"
```

### 2. 启用模型

AI 模块启动时会自动从数据库加载启用的 Provider 和 Model。你可以通过 SQL 快速启用：

```sql
-- 启用 OpenAI
UPDATE ai_provider SET status = 'active' WHERE name = 'openai';

-- 启用 GPT-4o 模型
UPDATE ai_model SET status = 'active' WHERE model = 'gpt-4o';
```

### 3. API 调用示例

**普通对话:**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "你好，Bingo"}]
  }'
```

**流式对话 (SSE):**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Accept: text/event-stream" \
  ...
```

**使用特定角色:**

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -d '{
    "role_id": "math_teacher",
    "messages": [{"role": "user", "content": "1+1等于几？"}]
  }'
```

## 📚 文档导航

- [供应商与模型管理 (Provider & Models)](./provider.md): 了解如何动态管理 AI 厂商和模型。
- [角色预设 (AI Roles)](./role.md): 了解如何创建和管理 AI 角色及其权限体系。
- [架构与机制 (Architecture)](./architecture.md): 深入了解流式处理、重试机制、上下文管理等核心实现。
