# AI Provider Dynamic Load Design

**Goal:** Enable dynamic loading of AI providers from database without code changes or restart.

**Problem:** Current `initAIRegistry()` hardcodes provider initialization in `internal/apiserver/http.go`. Adding or disabling providers requires code modification and redeployment. Database tables `ai_provider` and `ai_model` exist but are unused. Both `apiserver` and `admserver` need AI provider access.

**Solution:** Database-driven provider configuration with Redis Pub/Sub trigger and periodic polling fallback.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Config (YAML)                            │
│  ai.credentials: { openai.api-key, claude.api-key, ... }        │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      Database (MySQL)                            │
│  ┌─────────────────┐  ┌─────────────────────────────────────┐  │
│  │ ai_provider     │  │ ai_model                             │  │
│  │ - name          │  │ - provider_name                      │  │
│  │ - status        │  │ - model                              │  │
│  │ - is_default    │  │ - status                             │  │
│  │ - sort          │  │ - is_default                         │  │
│  └─────────────────┘  │ - sort                               │  │
│                       └─────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│              internal/pkg/ai (SHARED)                            │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ Loader                                                 │    │
│  │ - Load(ctx): load from DB + register to Registry      │    │
│  │ - Reload(ctx): clear and reload                       │    │
│  └────────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ Subscriber (Redis Pub/Sub)                             │    │
│  │ - Start(): listen for reload triggers                  │    │
│  │ - Stop(): graceful shutdown                            │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      pkg/ai.Registry                             │
│  - providers: map[name]Provider                                 │
│  - models: map[modelID]Provider                                 │
│  - Register/Clear/Get/GetByModel                                │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                             ▼
┌─────────────────────────┐   ┌─────────────────────────────────┐
│   apiserver             │   │   admserver                     │
│   - ChatBiz             │   │   - RoleBiz (AI-assisted)       │
│   - router.MapChat...   │   │   - future AI features          │
└─────────────────────────┘   └─────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     Reload Triggers                              │
│  ┌─────────────────────┐  ┌─────────────────────────────────┐  │
│  │ Redis Pub/Sub       │  │ Periodic Polling (fallback)     │  │
│  │ "ai:reload:providers"│  │ Every 5 minutes                 │  │
│  └─────────────────────┘  └─────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Components

### 1. Registry (pkg/ai/registry.go)

**Responsibility:** Thread-safe provider storage and lookup. No external dependencies.

**Changes:** Add `Clear()` method for reload support.

```go
type Registry struct {
    providers map[string]Provider
    models    map[string]Provider
    mu        sync.RWMutex
}

func NewRegistry() *Registry
func (r *Registry) Register(p Provider)
func (r *Registry) Unregister(name string)
func (r *Registry) Clear()  // NEW: removes all providers, for reload
func (r *Registry) Get(name string) (Provider, bool)
func (r *Registry) GetByModel(model string) (Provider, bool)
func (r *Registry) ListProviders() []string
func (r *Registry) ListModels() []ModelInfo

// NEW: Clear implementation
func (r *Registry) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers = make(map[string]Provider)
    r.models = make(map[string]Provider)
}
```

### 2. Loader (internal/pkg/ai/loader.go)

**Responsibility:** Bridge between database and Registry.

```go
type Credential struct {
    APIKey  string
    BaseURL string
}

type Loader struct {
    registry    *ai.Registry
    store       store.IStore
    credentials map[string]Credential  // from config
}

func NewLoader(registry *ai.Registry, store store.IStore, credentials map[string]Credential) *Loader
func (l *Loader) Load(ctx context.Context) error
func (l *Loader) Reload(ctx context.Context) error
```

**Load Flow:**
1. `store.AiProvider().ListActive(ctx)` - get active providers
2. `store.AiModel().ListActive(ctx)` - get active models
3. Group models by provider
4. For each provider with credential:
   - Create Provider instance
   - `registry.Register(provider)`

### 3. Subscriber (internal/pkg/ai/subscriber.go)

**Responsibility:** Listen to Redis Pub/Sub for instant reload. Shared by both apiserver and admserver.

```go
const AIReloadChannel = "ai:reload:providers"

type AIReloadSubscriber struct {
    redis  *redis.Client
    loader *ai.Loader
    ctx    context.Context
    cancel context.CancelFunc
}

func NewAIReloadSubscriber(loader *ai.Loader) *AIReloadSubscriber
func (s *AIReloadSubscriber) Start()
func (s *AIReloadSubscriber) Stop()
```

**Trigger:** Any service can publish:
```go
redis.Publish(ctx, "ai:reload:providers", "trigger")
```

### 4. Initialization (internal/pkg/ai/ai.go)

**Responsibility:** Centralized AI component initialization. Called from each server's `initConfig()`.

**Shared Pattern:** Both apiserver and admserver call `InitAI()` during startup.

```go
// internal/pkg/ai/ai.go
package ai

var (
    globalRegistry *Registry
    globalLoader   *Loader
)

// InitAI initializes AI registry and starts reload mechanisms.
// Called from initConfig() in each server.
func InitAI(redisClient *redis.Client, store store.IStore, creds map[string]Credential) (*Registry, error) {
    globalRegistry = NewRegistry()
    globalLoader = NewLoader(globalRegistry, store, creds)

    // Initial load from database
    if err := globalLoader.Load(context.Background()); err != nil {
        log.Errorw("Failed to load AI providers", "err", err)
        // Don't fail startup - AI is optional
    }

    // Start subscriber if Redis available
    if redisClient != nil {
        sub := NewSubscriber(redisClient, globalLoader)
        go sub.Start()
    }

    // Start periodic polling fallback (5 min)
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()
        for range ticker.C {
            _ = globalLoader.Reload(context.Background())
        }
    }()

    return globalRegistry, nil
}

// GetRegistry returns the global AI registry.
func GetRegistry() *Registry {
    return globalRegistry
}

// TriggerReload sends a Redis pub/sub message to trigger reload across all services.
func TriggerReload(ctx context.Context, redis *redis.Client) error {
    return redis.Publish(ctx, AIReloadChannel, "trigger").Err()
}
```

### 5. Server Integration

**apiserver/http.go:**
```go
func initGinEngine() *gin.Engine {
    g := bootstrap.InitGin()
    // ... other routes

    // AI Chat routes (if registry available)
    if registry := ai.GetRegistry(); registry != nil {
        v1 := g.Group("/v1")
        v1.Use(middleware.Lang())
        v1.Use(middleware.Maintenance())
        // ... auth, rate limit ...
        router.MapChatRouters(v1, registry)
    }

    return g
}
```

**admserver/http.go (future):**
```go
func initGinEngine() *gin.Engine {
    g := bootstrap.InitGin()
    // ... other routes

    // AI-assisted features (if registry available)
    if registry := ai.GetRegistry(); registry != nil {
        router.MapAIRoutes(g, registry)  // Role prompt generation, etc.
    }

    return g
}
```

---

## Data Model

### ai_provider table

**Change:** Remove `models` field (no longer used).

```sql
ALTER TABLE `ai_provider` DROP COLUMN `models`;
```

**Fields:**
- `name` (varchar(32)) - Provider ID: openai, claude, gemini, qwen...
- `display_name` - Display label
- `status` - active/disabled
- `is_default` - Default provider flag
- `sort` - Display order

### ai_model table

**No changes needed.**

**Fields:**
- `provider_name` - Links to ai_provider.name
- `model` - Model ID: gpt-4o, claude-3-5-sonnet...
- `display_name` - Display label
- `max_tokens` - Context window size
- `status` - active/disabled
- `is_default` - Default model flag
- `sort` - Display order

### Credential Storage

**Location:** Configuration file with environment variable injection.

```yaml
# configs/apiserver.yaml
ai:
  credentials:
    openai:
      api-key: ${OPENAI_API_KEY}
    claude:
      api-key: ${CLAUDE_API_KEY}
    gemini:
      api-key: ${GEMINI_API_KEY}
    qwen:
      api-key: ${QWEN_API_KEY}
      base-url: https://dashscope.aliyuncs.com/compatible-mode/v1
```

**Rationale:**
- Secrets never enter database or version control
- 12-Factor App compliant
- Simple to audit and rotate

---

## Files Changed

| File | Operation | Description |
|------|-----------|-------------|
| `pkg/ai/registry.go` | Modify | Add `Clear()` method |
| `internal/pkg/ai/loader.go` | Create | Database-driven provider loading |
| `internal/pkg/ai/subscriber.go` | Create | Redis Pub/Sub reload listener (shared) |
| `internal/pkg/ai/ai.go` | Create | Centralized AI initialization (shared) |
| `internal/apiserver/app.go` | Modify | Call `ai.InitAI()` in `initConfig()` |
| `internal/admserver/app.go` | Modify | Call `ai.InitAI()` in `initConfig()` |
| `internal/apiserver/http.go` | Modify | Use `ai.GetRegistry()` instead of `initAIRegistry()` |
| `internal/pkg/model/ai_provider.go` | Modify | Remove `Models` field |
| `internal/pkg/database/migration/...` | Create | Drop ai_provider.models column |
| `internal/pkg/testing/mock/store/store.go` | Modify | Add AiProvider/AiModel mock stores |

---

## Mock Store Design

For unit testing, extend the mock store with AI-related methods:

```go
// internal/pkg/testing/mock/store/store.go

// ABOUTME: Mock store implementations for testing.
// ABOUTME: Provides in-memory implementations of store interfaces.

type AiProviderStore struct {
    ListActiveResult []*model.AiProviderM
    ListActiveErr    error

    // For FirstOrCreate
    FirstOrCreateResult *model.AiProviderM
    FirstOrCreateErr    error
}

func (m *AiProviderStore) ListActive(ctx context.Context) ([]*model.AiProviderM, error) {
    if m.ListActiveErr != nil {
        return nil, m.ListActiveErr
    }
    return m.ListActiveResult, nil
}

func (m *AiProviderStore) FirstOrCreate(ctx context.Context, where any, data any) error {
    if m.FirstOrCreateErr != nil {
        return m.FirstOrCreateErr
    }
    // Update result pointer
    if result, ok := data.(*model.AiProviderM); ok && m.FirstOrCreateResult != nil {
        *result = *m.FirstOrCreateResult
    }
    return nil
}

type AiModelStore struct {
    ListActiveResult []*model.AiModelM
    ListActiveErr    error

    // For FirstOrCreate
    FirstOrCreateResult *model.AiModelM
    FirstOrCreateErr    error
}

func (m *AiModelStore) ListActive(ctx context.Context) ([]*model.AiModelM, error) {
    if m.ListActiveErr != nil {
        return nil, m.ListActiveErr
    }
    return m.ListActiveResult, nil
}

func (m *AiModelStore) FirstOrCreate(ctx context.Context, where any, data any) error {
    if m.FirstOrCreateErr != nil {
        return m.FirstOrCreateErr
    }
    if result, ok := data.(*model.AiModelM); ok && m.FirstOrCreateResult != nil {
        *result = *m.FirstOrCreateResult
    }
    return nil
}

// Store interface extension
type Store struct {
    // ... existing fields ...
    aiProvider *AiProviderStore
    aiModel    *AiModelStore
}

func (m *Store) AiProvider() store.AiProviderStore { return m.aiProvider }
func (m *Store) AiModel() store.AiModelStore { return m.aiModel }
```

---

## Scenarios

### Enable a new provider
```sql
INSERT INTO ai_provider (name, display_name, status, sort)
VALUES ('deepseek', 'DeepSeek', 'active', 10);
```
Next reload cycle (5 min) or Redis trigger → provider available.

### Disable a provider
```sql
UPDATE ai_provider SET status = 'disabled' WHERE name = 'openai';
```
Next reload → OpenAI unregistered.

### Disable a specific model
```sql
UPDATE ai_model SET status = 'disabled' WHERE model = 'gpt-3.5-turbo';
```
Next reload → gpt-3.5-turbo removed, gpt-4o still available.

---

## Security Considerations

1. **Credentials never in database** - Config file only
2. **No SQL injection** - Store layer uses parameterized queries
3. **Redis auth** - Pub/Sub channel requires authentication
4. **Graceful degradation** - Missing credential → provider skipped, logged

---

## Testing Plan

Per `docs/zh/development/testing.md`, use layered testing strategy:

### Unit Tests (Loader Layer)

**File:** `internal/pkg/ai/loader_test.go`

| Test Case | Description | Mock |
|-----------|-------------|------|
| `TestLoader_Load_Success` | Load providers from DB successfully | Mock Store returns valid data |
| `TestLoader_Load_NoCredential` | Skip provider without credential | Mock Store returns provider, no matching cred |
| `TestLoader_Load_DBError` | Handle database read error | Mock Store returns error |
| `TestLoader_Reload_ClearsFirst` | Verify Clear() called before reload | Spy on Registry.Clear() |
| `TestLoader_Reload_EmptyDB` | Handle empty database gracefully | Mock Store returns empty lists |

```go
func TestLoader_Load_Success(t *testing.T) {
    store := mockstore.NewStore()
    registry := ai.NewRegistry()

    // Setup mock data
    store.AiProvider().ListActiveResult = []*model.AiProviderM{
        {Name: "openai", Status: model.AiProviderStatusActive},
    }
    store.AiModel().ListActiveResult = []*model.AiModelM{
        {ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive},
    }

    creds := map[string]ai.Credential{
        "openai": {APIKey: "test-key"},
    }

    loader := ai.NewLoader(registry, store, creds)
    err := loader.Load(context.Background())

    require.NoError(t, err)
    provider, ok := registry.Get("openai")
    require.True(t, ok)
    assert.Equal(t, "openai", provider.Name())
}
```

### Integration Tests (Subscriber)

**File:** `internal/pkg/ai/subscriber_test.go`

| Test Case | Description | Dependencies |
|-----------|-------------|--------------|
| `TestSubscriber_Integration` | Redis publish triggers reload | Real Redis (testcontainers) |
| `TestSubscriber_Disconnect` | Handle Redis disconnection gracefully | Real Redis |

```go
func TestSubscriber_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup testcontainers Redis
    redis := setupRedis(t)
    defer redis.Close()

    registry := ai.NewRegistry()
    loader := ai.NewLoader(registry, mockStore, creds)
    sub := ai.NewSubscriber(redis, loader)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go sub.Start()

    // Trigger reload via Redis
    redis.Publish(ctx, ai.AIReloadChannel, "trigger")

    // Verify reload was called (via spy or timeout)
    // ...
}
```

### E2E Tests (Full Flow)

**File:** `internal/pkg/ai/e2e_test.go`

| Test Case | Description |
|-----------|-------------|
| `TestAIProviderReload_E2E` | Update DB → trigger reload → verify Registry |
| `TestAIProviderDisable_E2E` | Disable provider in DB → verify removed from Registry |

```go
func TestAIProviderReload_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping e2e test")
    }

    // Use real test database
    db := setupTestDB(t)
    redis := setupRedis(t)

    // Insert new provider
    db.Exec("INSERT INTO ai_provider ...")

    // Trigger reload
    ai.TriggerReload(context.Background(), redis)

    // Verify provider available
    // ...
}
```

### Test Execution

```bash
# Unit tests only (fast)
go test -short ./internal/pkg/ai/...

# All tests including integration
go test ./internal/pkg/ai/...

# With race detection
go test -race ./internal/pkg/ai/...
```
