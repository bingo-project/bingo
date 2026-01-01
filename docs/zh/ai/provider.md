# 供应商与模型管理 (Provider & Models)

Bingo AI 模块支持动态管理 AI 供应商和模型，无需更改代码或重启服务即可调整模型可用性。

## 1. 混合配置模式

为了平衡安全性和灵活性，我们采用 **配置/环境变量 + 数据库** 的混合管理模式。

### 1.1 敏感凭证 (Config/Env)

API Key、Base URL 等敏感信息**绝不存储在数据库中**，而是通过 `configs/bingo-apiserver.yaml` 配置，支持环境变量注入，符合 [12-Factor App](https://12factor.net/zh_cn/config) 原则。

```yaml
# configs/bingo-apiserver.yaml
ai:
  credentials:
    openai:
      api_key: "${OPENAI_API_KEY}"
    claude:
      api_key: "${CLAUDE_API_KEY}"
    # ... 其他厂商
```

### 1.2 元数据 (Database)

供应商的状态、模型的启用情况、优先级等元数据存储在数据库中，支持动态调整。

**表结构简述:**

- `ai_provider`: 存储厂商基本信息 (name, display_name, status, is_default, sort)
- `ai_model`: 存储模型详细参数 (model, display_name, max_tokens, status, is_default, sort)

## 2. 动态加载机制

系统通过 `pkg/ai/loader` 和 `pkg/ai/registry` 实现动态加载。

### 2.1 加载流程

1. **读取数据库**: 系统启动或触发重载时，从 `ai_provider` 和 `ai_model` 表读取 `status = 'active'` 的记录。
2. **匹配凭证**: 将数据库记录与 `config.yaml` 中的凭证进行匹配。只有既在数据库中启用，又配置了有效凭证的 Provider 才会加载。
3. **注册实例**: 初始化的 Provider 实例被注册到全局 `Registry` 中，供业务层即时调用。

### 2.2 触发重载

当你修改了数据库状态后，可以通过以下方式让变更生效：

1. **Redis Pub/Sub (推荐)**:
   向 `ai:reload:providers` 频道发送任意消息，所有订阅该频道的服务实例（apiserver, admserver）将立即重载。
   ```bash
   redis-cli publish ai:reload:providers "trigger"
   ```

2. **自动轮询 (兜底)**:
   系统后台每 5 分钟会自动检查一次数据库变更并重载。

3. **重启服务**:
   这是最暴力的方式，当然也有效。

## 3. 常用操作指南

### 3.1 接入新 AI 厂商

1. **实现接口**: 在 `pkg/ai/providers/<name>/` 下实现 `ai.Provider` 接口。
2. **配置凭证**: 在配置文件中添加对应的 `api_key` 配置项。
3. **添加数据**:
   ```sql
   INSERT INTO ai_provider (name, display_name, status, sort) 
   VALUES ('deepseek', 'DeepSeek', 'active', 10);
   ```

### 3.2 紧急禁用某个厂商

如果某个厂商服务挂了，可以立即禁用，自动切换到其他 Default 厂商（如有配置 fallback 逻辑）。

```sql
UPDATE ai_provider SET status = 'disabled' WHERE name = 'openai';
-- 别忘了触发重载
```

### 3.3 调整模型显示顺序

可以调整 `sort` 字段（值越小越靠前）来改变前端展示顺序。

```sql
UPDATE ai_model SET sort = 1 WHERE model = 'gpt-4o';
```
