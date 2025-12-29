# AI 对话功能设计

## 概述

为 Bingo 脚手架添加通用 AI 对话能力，支持多 Provider 切换、会话管理、流式输出。采用 Eino 作为底层框架，对外提供 OpenAI 兼容 API。

## 设计决策

| 决策点 | 选择 | 理由 |
|--------|------|------|
| 底层框架 | Eino (CloudWeGo) | 字节开源，Go 风格，类型安全，生产验证 |
| API 风格 | OpenAI 兼容 + 扩展字段 | 行业标准，前端可用 openai SDK |
| Provider 配置 | 三层覆盖 | 系统默认 → 用户偏好 → 请求指定 |
| 凭证管理 | 配置文件(环境变量) + DB | 敏感信息环境变量，业务配置存 DB |
| 会话管理 | 后端存储 | 多端同步，可审计，Token 控制 |
| BYOK | 预留扩展点，暂不实现 | 产品定位未定，符合 YAGNI |

## 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Layer                                │
│  POST /v1/chat/completions (OpenAI 兼容)                        │
│  GET  /v1/models                                                 │
│  GET  /v1/ai/sessions                                           │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    internal/apiserver                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   handler   │→ │     biz     │→ │    store    │              │
│  │  /chat/*    │  │  chat.go    │  │ session.go  │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│         │                │                                       │
│         │                ▼                                       │
│         │         ┌─────────────┐                                │
│         │         │  provider   │ ← 选择 Provider                │
│         │         │  resolver   │                                │
│         │         └─────────────┘                                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         pkg/ai                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   client    │  │  provider   │  │   config    │              │
│  │  (Eino)     │  │  registry   │  │             │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
│         │                                                        │
│         ▼                                                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐            │
│  │  OpenAI  │ │ DeepSeek │ │  Claude  │ │  Ollama  │  ...       │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

### 职责划分

| 层级 | 职责 |
|------|------|
| `pkg/ai` | 通用 AI 能力，无业务依赖，可被多个 server 复用 |
| `internal/apiserver/biz/chat` | 会话管理、Provider 选择、用量统计等业务逻辑 |
| `internal/apiserver/handler/chat` | HTTP 处理、流式响应、OpenAI 格式转换 |

## pkg/ai 设计

### 目录结构

```
pkg/ai/
├── client.go          # 统一客户端，封装 Eino
├── config.go          # 配置结构
├── provider.go        # Provider 接口定义
├── registry.go        # Provider 注册表
├── message.go         # Message、Request/Response 结构
├── errors.go          # 错误定义
└── providers/
    ├── openai/        # OpenAI 及兼容服务（DeepSeek、Moonshot 等）
    ├── anthropic/     # Claude（后续添加）
    └── ollama/        # 本地模型（后续添加）
```

### 核心接口

```go
// pkg/ai/provider.go

// Provider 定义 AI 服务提供商
type Provider interface {
    // Name 返回 provider 标识，如 "openai", "deepseek"
    Name() string

    // Chat 非流式调用
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

    // ChatStream 流式调用
    ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error)

    // Models 返回该 provider 支持的模型列表
    Models() []string
}

// ChatRequest OpenAI 兼容的请求结构
type ChatRequest struct {
    Model       string     `json:"model"`
    Messages    []Message  `json:"messages"`
    MaxTokens   int        `json:"max_tokens,omitempty"`
    Temperature float64    `json:"temperature,omitempty"`
    Stream      bool       `json:"stream,omitempty"`
}

// ChatResponse OpenAI 兼容的响应结构
type ChatResponse struct {
    ID      string   `json:"id"`
    Model   string   `json:"model"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
}
```

### Provider 注册表

```go
// pkg/ai/registry.go

type Registry struct {
    providers map[string]Provider  // name -> provider
    models    map[string]string    // model -> provider name
}

// Get 根据 model 名称获取对应的 Provider
func (r *Registry) Get(model string) (Provider, error) {
    providerName, ok := r.models[model]
    if !ok {
        return nil, ErrModelNotFound
    }
    return r.providers[providerName], nil
}

// Register 注册 provider 及其支持的 models
func (r *Registry) Register(p Provider) {
    r.providers[p.Name()] = p
    for _, model := range p.Models() {
        r.models[model] = p.Name()
    }
}
```

### 使用示例

```go
// 初始化
registry := ai.NewRegistry()
registry.Register(openai.New(cfg.OpenAI))
registry.Register(openai.NewCompatible("deepseek", cfg.DeepSeek))
registry.Register(anthropic.New(cfg.Anthropic))

// 使用（根据 model 自动路由到对应 provider）
provider, _ := registry.Get("gpt-4")
resp, _ := provider.Chat(ctx, &ai.ChatRequest{
    Model:    "gpt-4",
    Messages: messages,
})
```

## 配置结构

### YAML 配置

```yaml
# configs/bingo-apiserver.yaml

ai:
  # 系统默认模型（用户未指定时使用）
  default_model: "gpt-4o"

  # Provider 凭证配置（敏感信息通过环境变量注入）
  credentials:
    openai:
      api_key: "${OPENAI_API_KEY}"
      base_url: "https://api.openai.com/v1"

    deepseek:
      api_key: "${DEEPSEEK_API_KEY}"
      base_url: "https://api.deepseek.com/v1"

    moonshot:
      api_key: "${MOONSHOT_API_KEY}"
      base_url: "https://api.moonshot.cn/v1"

    anthropic:
      api_key: "${ANTHROPIC_API_KEY}"
      base_url: "https://api.anthropic.com"

    ollama:
      base_url: "http://localhost:11434"

  # 会话配置
  session:
    max_history: 50              # 最多保留消息条数
    max_tokens: 8000             # 历史消息最大 token 数
    ttl: 24h                     # 会话过期时间

  # 限流配置（按用户）
  rate_limit:
    requests_per_minute: 20
    tokens_per_day: 100000
```

### Go 配置结构

```go
// pkg/ai/config.go

type Config struct {
    DefaultModel string                       `mapstructure:"default_model"`
    Credentials  map[string]*CredentialConfig `mapstructure:"credentials"`
    Session      SessionConfig                `mapstructure:"session"`
    RateLimit    RateLimitConfig              `mapstructure:"rate_limit"`
}

type CredentialConfig struct {
    APIKey  string `mapstructure:"api_key"`
    BaseURL string `mapstructure:"base_url"`
}

type SessionConfig struct {
    MaxHistory int           `mapstructure:"max_history"`
    MaxTokens  int           `mapstructure:"max_tokens"`
    TTL        time.Duration `mapstructure:"ttl"`
}

type RateLimitConfig struct {
    RequestsPerMinute int `mapstructure:"requests_per_minute"`
    TokensPerDay      int `mapstructure:"tokens_per_day"`
}
```

### 三层覆盖逻辑

```go
// internal/apiserver/biz/chat/resolver.go

// ResolveModel 解析最终使用的 model
// 优先级：请求指定 > 用户偏好 > 系统默认
func (b *chatBiz) ResolveModel(ctx context.Context, reqModel string, userID string) string {
    // 1. 请求指定
    if reqModel != "" {
        return reqModel
    }

    // 2. 用户偏好（从 DB 或缓存取）
    if pref, _ := b.userStore.GetAIPreference(ctx, userID); pref.Model != "" {
        return pref.Model
    }

    // 3. 系统默认
    return b.cfg.AI.DefaultModel
}
```

## 数据库设计

### Provider 业务配置表

```sql
-- Provider 业务配置（动态管理启用/禁用等）
CREATE TABLE ai_provider (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(32) NOT NULL,              -- 'openai', 'deepseek'
    display_name VARCHAR(64),               -- '通义千问'
    status VARCHAR(16) DEFAULT 'active',    -- active/disabled
    models JSON,                            -- ["gpt-4o", "gpt-4o-mini"]
    is_default TINYINT DEFAULT 0,           -- 默认 provider
    sort INT DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI Provider 配置';

-- Model 业务配置
CREATE TABLE ai_model (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    provider_name VARCHAR(32) NOT NULL,     -- 关联 provider
    model VARCHAR(64) NOT NULL,             -- 'gpt-4o'
    display_name VARCHAR(64),               -- 'GPT-4o'
    max_tokens INT,                         -- 模型最大 token
    input_price DECIMAL(10,6),              -- 输入价格 $/1K tokens
    output_price DECIMAL(10,6),             -- 输出价格
    status VARCHAR(16) DEFAULT 'active',
    is_default TINYINT DEFAULT 0,
    sort INT DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_model (model),
    KEY idx_provider (provider_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI Model 配置';
```

### 会话管理表

```sql
-- 会话表
CREATE TABLE ai_session (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    session_id VARCHAR(32) NOT NULL,        -- 外部 ID
    uid VARCHAR(64) NOT NULL,               -- 用户 ID
    title VARCHAR(128),                     -- 会话标题
    model VARCHAR(64),                      -- 用户为该会话选择的模型
    message_count INT DEFAULT 0,
    total_tokens INT DEFAULT 0,
    status VARCHAR(16) DEFAULT 'active',    -- active/archived/deleted
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY uk_session_id (session_id),
    KEY idx_uid_status (uid, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI 会话';

-- 消息表
CREATE TABLE ai_message (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    session_id VARCHAR(32) NOT NULL,
    role VARCHAR(16) NOT NULL,              -- system/user/assistant
    content TEXT NOT NULL,
    tokens INT DEFAULT 0,
    model VARCHAR(64),                      -- 生成该消息的模型（assistant）
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    KEY idx_session_created (session_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI 消息';
```

## 会话管理流程

```
┌─────────────────────────────────────────────────────────────────┐
│  请求: POST /v1/chat/completions                                │
│  { "session_id": "xxx", "messages": [{"role":"user",...}] }    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. 无 session_id → 创建新会话                                   │
│  2. 有 session_id → 从 DB 加载历史消息                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. 拼装 messages: 历史消息 + 本次请求消息                       │
│  4. 滑动窗口裁剪（超过 max_tokens 时截断早期消息）               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  5. 调用 LLM                                                    │
│  6. 存储用户消息 + AI 回复到 ai_message                          │
│  7. 更新 session 统计（message_count, total_tokens）            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  8. 返回响应（含 session_id，前端下次带上）                      │
└─────────────────────────────────────────────────────────────────┘
```

### 滑动窗口裁剪

```go
// internal/apiserver/biz/chat/history.go

func (b *chatBiz) BuildMessages(ctx context.Context, sessionID string, newMsg Message) ([]Message, error) {
    // 1. 取历史消息
    history, _ := b.msgStore.ListBySession(ctx, sessionID)

    // 2. 拼装：历史 + 新消息
    messages := append(history, newMsg)

    // 3. 滑动窗口裁剪
    messages = b.truncateByTokens(messages, b.cfg.AI.Session.MaxTokens)

    return messages, nil
}

func (b *chatBiz) truncateByTokens(msgs []Message, maxTokens int) []Message {
    // 从后往前累加，保留最近的消息
    total := 0
    start := len(msgs)

    for i := len(msgs) - 1; i >= 0; i-- {
        total += msgs[i].Tokens
        if total > maxTokens {
            break
        }
        start = i
    }

    // 保留 system 消息（如果有）
    if start > 0 && msgs[0].Role == "system" {
        return append([]Message{msgs[0]}, msgs[start:]...)
    }

    return msgs[start:]
}
```

## API 设计

### 端点列表

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/chat/completions` | 对话（OpenAI 兼容） |
| GET | `/v1/models` | 获取可用模型列表（OpenAI 兼容） |
| GET | `/v1/ai/sessions` | 获取用户会话列表 |
| GET | `/v1/ai/sessions/:id` | 获取会话详情 |
| PUT | `/v1/ai/sessions/:id` | 更新会话 |
| DELETE | `/v1/ai/sessions/:id` | 删除会话 |

### 对话接口

**POST /v1/chat/completions**

```json
// Request
{
  "model": "gpt-4o",
  "messages": [
    {"role": "user", "content": "你好"}
  ],
  "stream": true,
  "max_tokens": 2000,
  "temperature": 0.7,

  // 扩展字段
  "session_id": "sess_xxx"
}

// Response（非流式）
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "model": "gpt-4o",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "你好！有什么可以帮你的？"},
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 15,
    "total_tokens": 25
  },

  // 扩展字段
  "session_id": "sess_xxx"
}

// Response（流式 SSE）
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"你"},"index":0}]}
data: {"id":"chatcmpl-xxx","choices":[{"delta":{"content":"好"},"index":0}]}
data: {"id":"chatcmpl-xxx","choices":[{"delta":{},"finish_reason":"stop","index":0}],"usage":{...},"session_id":"sess_xxx"}
data: [DONE]
```

### 模型列表

**GET /v1/models**

```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "owned_by": "openai",
      "is_default": true
    },
    {
      "id": "deepseek-chat",
      "object": "model",
      "owned_by": "deepseek",
      "is_default": false
    }
  ]
}
```

### 会话列表

**GET /v1/ai/sessions**

```json
{
  "data": [
    {
      "id": "sess_xxx",
      "title": "关于 Go 并发的讨论",
      "model": "gpt-4o",
      "message_count": 12,
      "created_at": "2025-12-29T10:00:00Z",
      "updated_at": "2025-12-29T11:30:00Z"
    }
  ],
  "total": 25
}
```

### 会话详情

**GET /v1/ai/sessions/:id**

```json
{
  "id": "sess_xxx",
  "title": "关于 Go 并发的讨论",
  "model": "gpt-4o",
  "message_count": 12,
  "messages": [
    {"role": "user", "content": "Go 的并发模型是什么？", "created_at": "..."},
    {"role": "assistant", "content": "Go 使用 goroutine 和 channel...", "created_at": "..."}
  ],
  "created_at": "2025-12-29T10:00:00Z",
  "updated_at": "2025-12-29T11:30:00Z"
}
```

## 文件改动清单

### 新增文件

| 文件 | 说明 |
|------|------|
| **pkg/ai/** | |
| `pkg/ai/client.go` | 统一客户端，封装 Eino 调用 |
| `pkg/ai/config.go` | 配置结构定义 |
| `pkg/ai/provider.go` | Provider 接口定义 |
| `pkg/ai/registry.go` | Provider 注册表 |
| `pkg/ai/message.go` | Message、ChatRequest/Response 等结构 |
| `pkg/ai/errors.go` | AI 相关错误定义 |
| `pkg/ai/providers/openai/provider.go` | OpenAI 及兼容服务实现 |
| `pkg/ai/providers/anthropic/provider.go` | Claude 实现（后续添加） |
| `pkg/ai/providers/ollama/provider.go` | 本地模型实现（后续添加） |
| **internal/apiserver/** | |
| `internal/apiserver/biz/chat/chat.go` | 对话业务逻辑 |
| `internal/apiserver/biz/chat/session.go` | 会话管理逻辑 |
| `internal/apiserver/biz/chat/resolver.go` | Model/Credential 解析 |
| `internal/apiserver/biz/chat/history.go` | 历史消息处理、滑动窗口 |
| `internal/apiserver/store/ai_session.go` | 会话 Store |
| `internal/apiserver/store/ai_message.go` | 消息 Store |
| `internal/apiserver/model/ai_session.go` | Session Model |
| `internal/apiserver/model/ai_message.go` | Message Model |
| `internal/apiserver/handler/chat/chat.go` | 对话 Handler |
| `internal/apiserver/handler/chat/session.go` | 会话管理 Handler |
| `internal/apiserver/handler/chat/stream.go` | SSE 流式响应处理 |
| `internal/apiserver/router/chat.go` | 路由注册 |
| **pkg/api/** | |
| `pkg/api/apiserver/v1/chat.go` | Request/Response DTO |
| **internal/pkg/errno/** | |
| `internal/pkg/errno/ai.go` | AI 相关错误码 |
| **数据库迁移** | |
| `internal/pkg/database/migration/xxx_create_ai_tables.go` | ai_session、ai_message、ai_provider、ai_model 表 |

### 修改文件

| 文件 | 改动 |
|------|------|
| `configs/bingo-apiserver.example.yaml` | 新增 `ai` 配置块 |
| `internal/apiserver/config/config.go` | 新增 AI 配置结构 |
| `internal/apiserver/server.go` | 初始化 AI Registry、注册路由 |
| `internal/apiserver/biz/biz.go` | 注入 ChatBiz |
| `internal/apiserver/store/store.go` | 注入 SessionStore、MessageStore |
| `go.mod` | 新增 eino 依赖 |

### 依赖新增

```bash
go get github.com/cloudwego/eino@latest
go get github.com/cloudwego/eino-ext/components/model/openai@latest
```

## 后续扩展

| 功能 | 说明 | 优先级 |
|------|------|--------|
| Claude Provider | 实现 Anthropic 适配 | P1 |
| Ollama Provider | 本地模型支持 | P2 |
| BYOK | 用户自带 API Key | P2 |
| 用量统计 | Token 消耗、费用统计 | P2 |
| RAG | 知识库检索增强 | P3 |
| Function Calling | 工具调用 | P3 |
