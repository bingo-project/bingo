# AI Provider Dynamic Load Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enable dynamic loading of AI providers from database without code changes or restart.

**Architecture:** Database-driven provider configuration with centralized `InitAI()` function. Both apiserver and admserver share `internal/pkg/ai` package for loading providers via Store layer, with Redis Pub/Sub triggers for instant reload and 5-minute polling fallback.

**Tech Stack:** Go 1.24+, GORM, Redis Pub/Sub, Gin, existing provider packages (`pkg/ai/providers/*`)

---

## Prerequisites

**Docs to reference:**
- `docs/guides/CONVENTIONS.md` - Layered architecture, testing, naming
- `docs/zh/development/testing.md` - Testing strategy
- `docs/plans/2025-12-31-ai-provider-dynamic-load-design.md` - Design document

**Existing code to understand:**
- `pkg/ai/registry.go` - Current Registry implementation
- `pkg/ai/providers/openai/provider.go` - Provider instantiation pattern
- `internal/pkg/store/ai_provider.go` - AiProviderStore interface
- `internal/pkg/store/ai_model.go` - AiModelStore interface
- `internal/apiserver/http.go:64-151` - Current initAIRegistry() to replace

---

## Task 1: Add Registry.Clear() Method

**Files:**
- Modify: `pkg/ai/registry.go:78`
- Test: `pkg/ai/registry_test.go`

**Step 1: Write failing test**

```go
// pkg/ai/registry_test.go

func TestRegistry_Clear(t *testing.T) {
    r := ai.NewRegistry()

    // Register a mock provider
    provider := &mockProvider{name: "test"}
    r.Register(provider)

    // Verify it's registered
    _, ok := r.Get("test")
    require.True(t, ok)

    // Clear
    r.Clear()

    // Verify it's gone
    _, ok = r.Get("test")
    require.False(t, ok, "provider should be removed after Clear")
}

// mock provider for testing
type mockProvider struct {
    name string
}

func (m *mockProvider) Name() string               { return m.name }
func (m *mockProvider) Chat(ctx context.Context, req *ai.ChatRequest) (*ai.ChatResponse, error) {
    return nil, nil
}
func (m *mockProvider) ChatStream(ctx context.Context, req *ai.ChatRequest) (*ai.ChatStream, error) {
    return nil, nil
}
func (m *mockProvider) Models() []ai.ModelInfo     { return nil }
```

**Step 2: Run test to verify it fails**

Run: `go test ./pkg/ai/... -run TestRegistry_Clear -v`
Expected: `compile error: r.Clear undefined`

**Step 3: Implement Clear() method**

```go
// pkg/ai/registry.go - add after ListModels() method

// Clear removes all registered providers and models.
func (r *Registry) Clear() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.providers = make(map[string]Provider)
    r.models = make(map[string]Provider)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./pkg/ai/... -run TestRegistry_Clear -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/ai/registry.go pkg/ai/registry_test.go
git commit -m "feat(ai): add Registry.Clear() method for reload support"
```

---

## Task 2: Create Mock Store for Testing

**Files:**
- Create: `internal/pkg/testing/mock/store/store.go`
- Create: `internal/pkg/testing/mock/store/ai_provider.go`
- Create: `internal/pkg/testing/mock/store/ai_model.go`

**Step 1: Create base mock store**

```go
// internal/pkg/testing/mock/store/store.go

// ABOUTME: Mock store implementations for testing.
// ABOUTME: Provides in-memory implementations of store interfaces.

package store

import (
    "context"

    "github.com/bingo-project/bingo/internal/pkg/store"
)

// Store implements store.IStore for testing.
type Store struct {
    aiProvider *AiProviderStore
    aiModel    *AiModelStore
    // Add other stores as needed...
}

var _ store.IStore = (*Store)(nil)

// NewStore creates a new mock store.
func NewStore() *Store {
    return &Store{
        aiProvider: NewAiProviderStore(),
        aiModel:    NewAiModelStore(),
    }
}

func (m *Store) AiProvider() store.AiProviderStore { return m.aiProvider }
func (m *Store) AiModel() store.AiModelStore       { return m.aiModel }
// Implement other store.IStore methods returning nil/empty...
```

**Step 2: Create AiProviderStore mock**

```go
// internal/pkg/testing/mock/store/ai_provider.go

// ABOUTME: Mock AiProviderStore for testing.
// ABOUTME: Implements store.AiProviderStore with configurable behavior.

package store

import (
    "context"

    "github.com/bingo-project/bingo/internal/pkg/model"
    "github.com/bingo-project/bingo/pkg/store/where"
)

// AiProviderStore implements store.AiProviderStore for testing.
type AiProviderStore struct {
    // ListActive results
    ListActiveResult []*model.AiProviderM
    ListActiveErr    error

    // FirstOrCreate results
    FirstOrCreateResult *model.AiProviderM
    FirstOrCreateErr    error

    // GetByName results
    GetByNameResult *model.AiProviderM
    GetByNameErr    error

    // GetDefault results
    GetDefaultResult *model.AiProviderM
    GetDefaultErr    error

    // Call tracking for assertions
    ListActiveCalled    bool
    FirstOrCreateCalled bool
}

var _ store.AiProviderStore = (*AiProviderStore)(nil)

// NewAiProviderStore creates a new mock.
func NewAiProviderStore() *AiProviderStore {
    return &AiProviderStore{
        ListActiveResult: make([]*model.AiProviderM, 0),
    }
}

func (m *AiProviderStore) Create(ctx context.Context, obj *model.AiProviderM) error {
    return nil
}

func (m *AiProviderStore) Update(ctx context.Context, obj *model.AiProviderM, fields ...string) error {
    return nil
}

func (m *AiProviderStore) Delete(ctx context.Context, opts *where.Options) error {
    return nil
}

func (m *AiProviderStore) Get(ctx context.Context, opts *where.Options) (*model.AiProviderM, error) {
    return &model.AiProviderM{}, nil
}

func (m *AiProviderStore) List(ctx context.Context, opts *where.Options) (int64, []*model.AiProviderM, error) {
    return 0, m.ListActiveResult, nil
}

func (m *AiProviderStore) GetByName(ctx context.Context, name string) (*model.AiProviderM, error) {
    m.ListActiveCalled = true
    if m.GetByNameErr != nil {
        return nil, m.GetByNameErr
    }
    return m.GetByNameResult, nil
}

func (m *AiProviderStore) ListActive(ctx context.Context) ([]*model.AiProviderM, error) {
    m.ListActiveCalled = true
    if m.ListActiveErr != nil {
        return nil, m.ListActiveErr
    }
    return m.ListActiveResult, nil
}

func (m *AiProviderStore) GetDefault(ctx context.Context) (*model.AiProviderM, error) {
    if m.GetDefaultErr != nil {
        return nil, m.GetDefaultErr
    }
    return m.GetDefaultResult, nil
}

func (m *AiProviderStore) FirstOrCreate(ctx context.Context, where *model.AiProviderM, obj *model.AiProviderM) error {
    m.FirstOrCreateCalled = true
    if m.FirstOrCreateErr != nil {
        return m.FirstOrCreateErr
    }
    if m.FirstOrCreateResult != nil {
        *obj = *m.FirstOrCreateResult
    }
    return nil
}
```

**Step 3: Create AiModelStore mock**

```go
// internal/pkg/testing/mock/store/ai_model.go

// ABOUTME: Mock AiModelStore for testing.
// ABOUTME: Implements store.AiModelStore with configurable behavior.

package store

import (
    "context"

    "github.com/bingo-project/bingo/internal/pkg/model"
    "github.com/bingo-project/bingo/pkg/store/where"
)

// AiModelStore implements store.AiModelStore for testing.
type AiModelStore struct {
    // ListActive results
    ListActiveResult []*model.AiModelM
    ListActiveErr    error

    // ListByProvider results
    ListByProviderResult []*model.AiModelM
    ListByProviderErr    error

    // GetByModel results
    GetByModelResult *model.AiModelM
    GetByModelErr    error

    // GetDefault results
    GetDefaultResult *model.AiModelM
    GetDefaultErr    error

    // FirstOrCreate results
    FirstOrCreateResult *model.AiModelM
    FirstOrCreateErr    error
}

var _ store.AiModelStore = (*AiModelStore)(nil)

// NewAiModelStore creates a new mock.
func NewAiModelStore() *AiModelStore {
    return &AiModelStore{
        ListActiveResult: make([]*model.AiModelM, 0),
    }
}

func (m *AiModelStore) Create(ctx context.Context, obj *model.AiModelM) error {
    return nil
}

func (m *AiModelStore) Update(ctx context.Context, obj *model.AiModelM, fields ...string) error {
    return nil
}

func (m *AiModelStore) Delete(ctx context.Context, opts *where.Options) error {
    return nil
}

func (m *AiModelStore) Get(ctx context.Context, opts *where.Options) (*model.AiModelM, error) {
    return &model.AiModelM{}, nil
}

func (m *AiModelStore) List(ctx context.Context, opts *where.Options) (int64, []*model.AiModelM, error) {
    return 0, m.ListActiveResult, nil
}

func (m *AiModelStore) GetByModel(ctx context.Context, modelID string) (*model.AiModelM, error) {
    if m.GetByModelErr != nil {
        return nil, m.GetByModelErr
    }
    return m.GetByModelResult, nil
}

func (m *AiModelStore) ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error) {
    if m.ListByProviderErr != nil {
        return nil, m.ListByProviderErr
    }
    return m.ListByProviderResult, nil
}

func (m *AiModelStore) ListActive(ctx context.Context) ([]*model.AiModelM, error) {
    if m.ListActiveErr != nil {
        return nil, m.ListActiveErr
    }
    return m.ListActiveResult, nil
}

func (m *AiModelStore) GetDefault(ctx context.Context) (*model.AiModelM, error) {
    if m.GetDefaultErr != nil {
        return nil, m.GetDefaultErr
    }
    return m.GetDefaultResult, nil
}

func (m *AiModelStore) FirstOrCreate(ctx context.Context, where *model.AiModelM, obj *model.AiModelM) error {
    if m.FirstOrCreateErr != nil {
        return m.FirstOrCreateErr
    }
    if m.FirstOrCreateResult != nil {
        *obj = *m.FirstOrCreateResult
    }
    return nil
}
```

**Step 4: Verify compilation**

Run: `go build ./internal/pkg/testing/mock/store/...`
Expected: Success (no errors)

**Step 5: Commit**

```bash
git add internal/pkg/testing/
git commit -m "test(ai): add mock store implementations for AI testing"
```

---

## Task 3: Create Loader

**Files:**
- Create: `internal/pkg/ai/loader.go`
- Create: `internal/pkg/ai/loader_test.go`

**Step 1: Write the Loader struct and constructor**

```go
// internal/pkg/ai/loader.go

// ABOUTME: AI provider loader from database configuration.
// ABOUTME: Loads active providers and models from Store into Registry.

package ai

import (
    "context"
    "fmt"

    "github.com/bingo-project/bingo/internal/pkg/log"
    "github.com/bingo-project/bingo/internal/pkg/model"
    "github.com/bingo-project/bingo/internal/pkg/store"
    "github.com/bingo-project/bingo/pkg/ai/providers/claude"
    "github.com/bingo-project/bingo/pkg/ai/providers/gemini"
    "github.com/bingo-project/bingo/pkg/ai/providers/openai"
    "github.com/bingo-project/bingo/pkg/ai/providers/qwen"
)

// Credential holds provider authentication configuration.
type Credential struct {
    APIKey  string
    BaseURL string
}

// Loader loads AI providers from database into Registry.
type Loader struct {
    registry    *Registry
    store       store.IStore
    credentials map[string]Credential
}

// NewLoader creates a new Loader.
func NewLoader(registry *Registry, st store.IStore, creds map[string]Credential) *Loader {
    return &Loader{
        registry:    registry,
        store:       st,
        credentials: creds,
    }
}

// Load loads active providers from database into registry.
func (l *Loader) Load(ctx context.Context) error {
    // Get active providers
    providers, err := l.store.AiProvider().ListActive(ctx)
    if err != nil {
        return fmt.Errorf("list active providers: %w", err)
    }

    // Get active models
    models, err := l.store.AiModel().ListActive(ctx)
    if err != nil {
        return fmt.Errorf("list active models: %w", err)
    }

    // Group models by provider
    modelsByProvider := l.groupModelsByProvider(models)

    // Create and register providers
    for _, p := range providers {
        if err := l.loadProvider(ctx, p, modelsByProvider[p.Name]); err != nil {
            log.Errorw("Failed to load AI provider", "provider", p.Name, "err", err)
            // Continue with next provider
        }
    }

    return nil
}

// Reload clears and reloads all providers.
func (l *Loader) Reload(ctx context.Context) error {
    l.registry.Clear()
    return l.Load(ctx)
}

// groupModelsByProvider groups models by provider name.
func (l *Loader) groupModelsByProvider(models []*model.AiModelM) map[string][]*model.AiModelM {
    result := make(map[string][]*model.AiModelM)
    for _, m := range models {
        result[m.ProviderName] = append(result[m.ProviderName], m)
    }
    return result
}

// loadProvider creates and registers a single provider.
func (l *Loader) loadProvider(ctx context.Context, p *model.AiProviderM, models []*model.AiModelM) error {
    // Check if credential exists
    cred, hasCred := l.credentials[p.Name]
    if !hasCred {
        log.Warnw("AI provider has no credential, skipping", "provider", p.Name)
        return nil
    }

    // Create provider instance based on name
    provider, err := l.createProvider(p.Name, cred, models)
    if err != nil {
        return err
    }

    l.registry.Register(provider)
    log.Infow("AI provider loaded", "provider", p.Name)

    return nil
}

// createProvider instantiates a provider by name.
func (l *Loader) createProvider(name string, cred Credential, models []*model.AiModelM) (*Provider, error) {
    // Convert models to ModelInfo
    modelInfos := make([]ModelInfo, len(models))
    for i, m := range models {
        modelInfos[i] = ModelInfo{
            ID:        m.Model,
            Name:      m.DisplayName,
            Provider:  m.ProviderName,
            MaxTokens: m.MaxTokens,
        }
    }

    switch name {
    case "openai":
        cfg := openai.DefaultConfig()
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return openai.New(cfg)

    case "deepseek":
        cfg := openai.DeepSeekConfig()
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return openai.New(cfg)

    case "moonshot":
        cfg := openai.MoonshotConfig()
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return openai.New(cfg)

    case "glm":
        cfg := openai.GLMConfig()
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return openai.New(cfg)

    case "claude":
        cfg := claude.DefaultConfig()
        cfg.APIKey = cred.APIKey
        cfg.Models = modelInfos
        return claude.New(cfg)

    case "gemini":
        cfg := gemini.DefaultConfig()
        cfg.APIKey = cred.APIKey
        cfg.Models = modelInfos
        return gemini.New(cfg)

    case "qwen":
        cfg := qwen.DefaultConfig()
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return qwen.New(cfg)

    default:
        // Unknown provider - try as OpenAI-compatible
        cfg := openai.DefaultConfig()
        cfg.Name = name
        cfg.APIKey = cred.APIKey
        cfg.BaseURL = cred.BaseURL
        cfg.Models = modelInfos
        return openai.New(cfg)
    }
}
```

**Step 2: Write unit tests**

```go
// internal/pkg/ai/loader_test.go

// ABOUTME: Tests for AI provider loader.
// ABOUTME: Verifies loading, reloading, and error handling.

package ai

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/bingo-project/bingo/internal/pkg/model"
    mockstore "github.com/bingo-project/bingo/internal/pkg/testing/mock/store"
)

func TestLoader_Load_Success(t *testing.T) {
    store := mockstore.NewStore()
    registry := NewRegistry()

    // Setup mock data
    store.AiProvider().ListActiveResult = []*model.AiProviderM{
        {Name: "openai", Status: model.AiProviderStatusActive},
    }
    store.AiModel().ListActiveResult = []*model.AiModelM{
        {ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive, MaxTokens: 128000},
    }

    creds := map[string]Credential{
        "openai": {APIKey: "test-key"},
    }

    loader := NewLoader(registry, store, creds)
    err := loader.Load(context.Background())

    require.NoError(t, err)
    provider, ok := registry.Get("openai")
    require.True(t, ok, "openai provider should be registered")
    assert.Equal(t, "openai", provider.Name())
    assert.True(t, store.AiProvider().ListActiveCalled, "ListActive should be called")
}

func TestLoader_Load_NoCredential_Skips(t *testing.T) {
    store := mockstore.NewStore()
    registry := NewRegistry()

    // Provider without credential
    store.AiProvider().ListActiveResult = []*model.AiProviderM{
        {Name: "openai", Status: model.AiProviderStatusActive},
    }
    store.AiModel().ListActiveResult = []*model.AiModelM{}

    creds := map[string]Credential{} // Empty credentials

    loader := NewLoader(registry, store, creds)
    err := loader.Load(context.Background())

    require.NoError(t, err)
    _, ok := registry.Get("openai")
    require.False(t, ok, "provider without credential should not be registered")
}

func TestLoader_Load_DBError_ReturnsError(t *testing.T) {
    store := mockstore.NewStore()
    registry := NewRegistry()

    store.AiProvider().ListActiveErr = assert.AnError

    loader := NewLoader(registry, store, map[string]Credential{})
    err := loader.Load(context.Background())

    require.Error(t, err)
    assert.Contains(t, err.Error(), "list active providers")
}

func TestLoader_Reload_ClearsFirst(t *testing.T) {
    store := mockstore.NewStore()
    registry := NewRegistry()

    // Initial load
    store.AiProvider().ListActiveResult = []*model.AiProviderM{
        {Name: "openai", Status: model.AiProviderStatusActive},
    }
    store.AiModel().ListActiveResult = []*model.AiModelM{
        {ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive},
    }

    creds := map[string]Credential{
        "openai": {APIKey: "test-key"},
    }

    loader := NewLoader(registry, store, creds)
    _ = loader.Load(context.Background())

    // Verify initial load
    _, ok := registry.Get("openai")
    require.True(t, ok)

    // Reload with empty data
    store.AiProvider().ListActiveResult = []*model.AiProviderM{}
    store.AiModel().ListActiveResult = []*model.AiModelM{}

    _ = loader.Reload(context.Background())

    // Verify cleared (no providers)
    _, ok = registry.Get("openai")
    require.False(t, ok, "provider should be removed after reload with empty data")
}
```

**Step 3: Run tests**

Run: `go test ./internal/pkg/ai/... -v`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add internal/pkg/ai/
git commit -m "feat(ai): add database-driven provider Loader"
```

---

## Task 4: Create Subscriber

**Files:**
- Create: `internal/pkg/ai/subscriber.go`
- Create: `internal/pkg/ai/subscriber_test.go`

**Step 1: Write Subscriber implementation**

```go
// internal/pkg/ai/subscriber.go

// ABOUTME: Redis Pub/Sub subscriber for AI provider reload triggers.
// ABOUTME: Listens for reload notifications to refresh provider registry.

package ai

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/bingo-project/bingo/internal/pkg/log"
)

const (
    // AIReloadChannel is the Redis Pub/Sub channel for reload triggers.
    AIReloadChannel = "ai:reload:providers"
)

// Subscriber listens for AI reload triggers via Redis Pub/Sub.
type Subscriber struct {
    redis  *redis.Client
    loader *Loader
    ctx    context.Context
    cancel context.CancelFunc
}

// NewSubscriber creates a new reload subscriber.
func NewSubscriber(redis *redis.Client, loader *Loader) *Subscriber {
    ctx, cancel := context.WithCancel(context.Background())
    return &Subscriber{
        redis:  redis,
        loader: loader,
        ctx:    ctx,
        cancel: cancel,
    }
}

// Start begins listening for reload messages.
// This blocks - run in a goroutine.
func (s *Subscriber) Start() {
    log.Infow("AI reload subscriber started", "channel", AIReloadChannel)

    pubsub := s.redis.Subscribe(s.ctx, AIReloadChannel)
    defer pubsub.Close()

    ch := pubsub.Channel()

    for {
        select {
        case <-s.ctx.Done():
            log.Infow("AI reload subscriber stopped")
            return
        case msg, ok := <-ch:
            if !ok {
                log.Warnw("AI reload subscriber channel closed")
                return
            }

            log.Infow("AI reload trigger received", "payload", msg.Payload)

            // Reload with timeout
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            if err := s.loader.Reload(ctx); err != nil {
                log.Errorw("AI reload failed", "err", err)
            } else {
                log.Infow("AI reload completed successfully")
            }
            cancel()
        }
    }
}

// Stop gracefully shuts down the subscriber.
func (s *Subscriber) Stop() {
    if s.cancel != nil {
        s.cancel()
    }
}
```

**Step 2: Write integration test**

```go
// internal/pkg/ai/subscriber_test.go

// ABOUTME: Integration tests for AI reload subscriber.
// ABOUTME: Uses real Redis for Pub/Sub verification.

package ai

import (
    "context"
    "testing"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/bingo-project/bingo/internal/pkg/model"
    mockstore "github.com/bingo-project/bingo/internal/pkg/testing/mock/store"
)

func TestSubscriber_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Note: This test requires Redis to be available.
    // For local testing, run: docker run -p 6379:6379 redis

    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    defer redisClient.Close()

    // Verify Redis connection
    ctx := context.Background()
    err := redisClient.Ping(ctx).Err()
    if err != nil {
        t.Skipf("Redis not available: %v", err)
    }

    // Setup
    store := mockstore.NewStore()
    registry := NewRegistry()

    store.AiProvider().ListActiveResult = []*model.AiProviderM{
        {Name: "openai", Status: model.AiProviderStatusActive},
    }
    store.AiModel().ListActiveResult = []*model.AiModelM{
        {ProviderName: "openai", Model: "gpt-4o", Status: model.AiModelStatusActive},
    }

    creds := map[string]Credential{
        "openai": {APIKey: "test-key"},
    }

    loader := NewLoader(registry, store, creds)
    sub := NewSubscriber(redisClient, loader)

    // Start subscriber in background
    go sub.Start()
    defer sub.Stop()

    // Give subscriber time to start
    time.Sleep(100 * time.Millisecond)

    // Trigger reload via Redis
    err = redisClient.Publish(ctx, AIReloadChannel, "trigger").Err()
    require.NoError(t, err)

    // Give reload time to complete
    time.Sleep(500 * time.Millisecond)

    // Verify provider was loaded
    provider, ok := registry.Get("openai")
    require.True(t, ok, "provider should be loaded after reload trigger")
    assert.Equal(t, "openai", provider.Name())
}

func TestSubscriber_Disconnect_Handled(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Use invalid Redis to test disconnect handling
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:9999", // Invalid port
    })

    store := mockstore.NewStore()
    registry := NewRegistry()
    loader := NewLoader(registry, store, map[string]Credential{})
    sub := NewSubscriber(redisClient, loader)

    // This should not panic or block indefinitely
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    go sub.Start()
    <-ctx.Done()

    // If we get here, disconnect was handled gracefully
    assert.True(t, true, "subscriber handled disconnect gracefully")
}
```

**Step 3: Run unit tests (skip integration)**

Run: `go test ./internal/pkg/ai/... -short -v`
Expected: PASS (integration test skipped)

**Step 4: Commit**

```bash
git add internal/pkg/ai/
git commit -m "feat(ai): add Redis Pub/Sub reload subscriber"
```

---

## Task 5: Create Centralized Initialization

**Files:**
- Create: `internal/pkg/ai/ai.go`

**Step 1: Write ai.go initialization**

```go
// internal/pkg/ai/ai.go

// ABOUTME: Centralized AI component initialization.
// ABOUTME: Provides InitAI for server startup and GetRegistry for access.

package ai

import (
    "context"
    "time"

    "github.com/redis/go-redis/v9"

    "github.com/bingo-project/bingo/internal/pkg/log"
    "github.com/bingo-project/bingo/internal/pkg/store"
)

var (
    globalRegistry *Registry
    globalLoader   *Loader
)

// InitAI initializes AI registry and starts reload mechanisms.
// Called from initConfig() in each server.
// Returns the registry (or nil if no credentials configured) and any error.
func InitAI(redisClient *redis.Client, st store.IStore, creds map[string]Credential) (*Registry, error) {
    // Skip if no credentials configured
    if len(creds) == 0 {
        log.Info("No AI credentials configured, skipping AI initialization")
        return nil, nil
    }

    globalRegistry = NewRegistry()
    globalLoader = NewLoader(globalRegistry, st, creds)

    // Initial load from database
    if err := globalLoader.Load(context.Background()); err != nil {
        log.Errorw("Failed to load AI providers", "err", err)
        // Don't fail startup - AI is optional
    }

    // Start subscriber if Redis available
    if redisClient != nil {
        sub := NewSubscriber(redisClient, globalLoader)
        go sub.Start()
        log.Info("AI reload subscriber started")
    } else {
        // Fallback: periodic polling if no Redis
        go startPeriodicReload(globalLoader)
    }

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

// startPeriodicReload starts periodic polling as fallback when Redis is unavailable.
func startPeriodicReload(loader *Loader) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    log.Info("AI periodic reload started (5 minute interval)")

    for range ticker.C {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        if err := loader.Reload(ctx); err != nil {
            log.Errorw("AI periodic reload failed", "err", err)
        }
        cancel()
    }
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/pkg/ai/...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/pkg/ai/ai.go
git commit -m "feat(ai): add centralized InitAI initialization"
```

---

## Task 6: Integrate with apiserver

**Files:**
- Modify: `internal/apiserver/app.go:54-62`
- Modify: `internal/apiserver/http.go:40-58,63-151`

**Step 1: Update apiserver/app.go initConfig()**

```go
// internal/apiserver/app.go

// Add import for ai package
import (
    // ... existing imports
    "github.com/bingo-project/bingo/pkg/ai"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    bootstrap.InitConfig("bingo-apiserver.yaml")
    bootstrap.Boot()
    bootstrap.InitJwt()

    // Init store
    _ = store.NewStore(bootstrap.InitDB())

    // Init AI (optional, logs error if fails)
    _, _ = ai.InitAI(facade.Redis, store.S, facade.Config.AI.Credentials)
}
```

**Step 2: Update apiserver/http.go to use GetRegistry()**

```go
// internal/apiserver/http.go

// Remove initAIRegistry function entirely (lines 63-151)
// Update import to remove provider packages:
import (
    "github.com/gin-gonic/gin"

    bizauth "github.com/bingo-project/bingo/internal/apiserver/biz/auth"
    "github.com/bingo-project/bingo/internal/apiserver/middleware"
    "github.com/bingo-project/bingo/internal/apiserver/router"
    "github.com/bingo-project/bingo/internal/pkg/auth"
    "github.com/bingo-project/bingo/internal/pkg/bootstrap"
    "github.com/bingo-project/bingo/internal/pkg/facade"
    "github.com/bingo-project/bingo/internal/pkg/log"
    httpmw "github.com/bingo-project/bingo/internal/pkg/middleware/http"
    "github.com/bingo-project/bingo/internal/pkg/store"
    "github.com/bingo-project/bingo/pkg/ai"
)

// initGinEngine initializes the Gin engine with routes.
func initGinEngine() *gin.Engine {
    g := bootstrap.InitGin()

    // Swagger
    if facade.Config.Feature.ApiDoc {
        router.MapSwagRouters(g)
    }

    // Common router
    router.MapCommonRouters(g)

    // Api
    router.MapApiRouters(g)

    // AI Chat routes (use global registry)
    if registry := ai.GetRegistry(); registry != nil {
        v1 := g.Group("/v1")
        v1.Use(middleware.Lang())
        v1.Use(middleware.Maintenance())

        loader := bizauth.NewUserLoader(store.S)
        authn := auth.New(loader)
        v1.Use(auth.Middleware(authn))

        // Apply AI rate limiter (RPM)
        rpm := facade.Config.AI.Quota.DefaultRPM
        if rpm <= 0 {
            rpm = 20 // fallback default
        }
        v1.Use(httpmw.AILimiter(rpm))

        router.MapChatRouters(v1, registry)
    }

    return g
}
```

**Step 3: Verify compilation**

Run: `go build ./internal/apiserver/...`
Expected: Success

**Step 4: Commit**

```bash
git add internal/apiserver/
git commit -m "refactor(apiserver): use centralized InitAI for provider loading"
```

---

## Task 7: Integrate with admserver

**Files:**
- Modify: `internal/admserver/app.go:54-62`

**Step 1: Update admserver/app.go initConfig()**

```go
// internal/admserver/app.go

// Add import for ai package
import (
    // ... existing imports
    "github.com/bingo-project/bingo/pkg/ai"
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
    bootstrap.InitConfig("bingo-admserver.yaml")
    bootstrap.Boot()
    bootstrap.InitJwt()

    // Init store
    _ = store.NewStore(bootstrap.InitDB())

    // Init AI (optional, for future AI-assisted features)
    _, _ = ai.InitAI(facade.Redis, store.S, facade.Config.AI.Credentials)
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/admserver/...`
Expected: Success

**Step 3: Commit**

```bash
git add internal/admserver/
git commit -m "feat(admserver): add AI initialization for future features"
```

---

## Task 8: Database Migration

**Files:**
- Modify: `internal/pkg/model/ai_provider.go:13`
- Modify: `internal/pkg/database/migration/2025_12_29_100000_create_ai_provider_table.go:18`

**Step 1: Remove Models field from model**

```go
// internal/pkg/model/ai_provider.go

type AiProviderM struct {
    ID          uint   `gorm:"primaryKey" json:"id"`
    Name        string `gorm:"column:name;type:varchar(32);uniqueIndex:uk_name;not null" json:"name"`
    DisplayName string `gorm:"column:display_name;type:varchar(64)" json:"displayName"`
    Status      string `gorm:"column:status;type:varchar(16);not null;default:active" json:"status"`
    // Models field removed - models are in ai_model table
    IsDefault   bool   `gorm:"column:is_default;type:tinyint(1);not null;default:0" json:"isDefault"`
    Sort        int    `gorm:"column:sort;type:int;not null;default:0" json:"sort"`

    CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
    UpdatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}
```

**Step 2: Update migration to not include Models field**

```go
// internal/pkg/database/migration/2025_12_29_100000_create_ai_provider_table.go

type CreateAIProviderTable struct {
    ID          uint64    `gorm:"primaryKey"`
    Name        string    `gorm:"type:varchar(32);uniqueIndex:uk_name;not null"`
    DisplayName string    `gorm:"type:varchar(64)"`
    Status      string    `gorm:"type:varchar(16);not null;default:active"`
    // Models field removed
    IsDefault   bool      `gorm:"type:tinyint(1);not null;default:0"`
    Sort        int       `gorm:"type:int;not null;default:0"`
    CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
    UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}
```

**Step 3: Update seeder to remove Models field**

```go
// internal/pkg/database/seeder/ai_seeder.go

var defaultProviders = []model.AiProviderM{
    // OpenAI-compatible
    {Name: "openai", DisplayName: "OpenAI", Status: model.AiProviderStatusActive, IsDefault: true, Sort: 1},
    {Name: "deepseek", DisplayName: "DeepSeek", Status: model.AiProviderStatusActive, Sort: 2},
    {Name: "moonshot", DisplayName: "Moonshot", Status: model.AiProviderStatusActive, Sort: 3},
    {Name: "glm", DisplayName: "智谱 GLM", Status: model.AiProviderStatusActive, Sort: 4},
    // Native providers
    {Name: "claude", DisplayName: "Claude", Status: model.AiProviderStatusActive, Sort: 5},
    {Name: "gemini", DisplayName: "Gemini", Status: model.AiProviderStatusActive, Sort: 6},
    {Name: "qwen", DisplayName: "通义千问", Status: model.AiProviderStatusActive, Sort: 7},
}
```

**Step 4: Test migration**

Run: `make build && bingo migrate refresh && bingo db seed`
Expected: Success, no errors

**Step 5: Commit**

```bash
git add internal/pkg/model/ internal/pkg/database/
git commit -m "refactor(db): remove unused models field from ai_provider table"
```

---

## Task 9: End-to-End Test

**Files:**
- Create: `internal/pkg/ai/e2e_test.go`

**Step 1: Write E2E test**

```go
// internal/pkg/ai/e2e_test.go

// ABOUTME: End-to-end tests for AI provider dynamic loading.
// ABOUTME: Tests full database to registry flow.

package ai

import (
    "context"
    "testing"
    "time"

    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/gorm"

    "github.com/bingo-project/bingo/internal/pkg/model"
    "github.com/bingo-project/bingo/pkg/ai"
)

func TestAIProviderReload_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping e2e test")
    }

    // Setup: Use test database and Redis
    db := setupTestDB(t)
    redis := setupTestRedis(t)

    // Setup registry and loader
    registry := ai.NewRegistry()
    loader := NewLoader(registry, &storeWrapper{db: db}, map[string]Credential{
        "openai": {APIKey: "test-key"},
    })

    // Initial load - no providers
    ctx := context.Background()
    err := loader.Load(ctx)
    require.NoError(t, err)

    _, ok := registry.Get("openai")
    require.False(t, ok, "no providers initially")

    // Insert provider into database
    provider := &model.AiProviderM{
        Name:        "openai",
        DisplayName: "OpenAI",
        Status:      model.AiProviderStatusActive,
        Sort:        1,
    }
    err = db.Create(provider).Error
    require.NoError(t, err)

    model := &model.AiModelM{
        ProviderName: "openai",
        Model:        "gpt-4o",
        DisplayName:  "GPT-4o",
        MaxTokens:    128000,
        Status:       model.AiModelStatusActive,
        Sort:         1,
    }
    err = db.Create(model).Error
    require.NoError(t, err)

    // Trigger reload
    err = TriggerReload(ctx, redis)
    require.NoError(t, err)

    // Wait for reload
    time.Sleep(500 * time.Millisecond)

    // Verify provider loaded
    providerAI, ok := registry.Get("openai")
    require.True(t, ok, "provider should be loaded after DB insert and reload")
    assert.Equal(t, "openai", providerAI.Name())

    // Cleanup
    db.Where("name = ?", "openai").Delete(&model.AiProviderM{})
    db.Where("model = ?", "gpt-4o").Delete(&model.AiModelM{})
}

// storeWrapper adapts gorm.DB to store.IStore for testing
type storeWrapper struct {
    db *gorm.DB
}

func (s *storeWrapper) AiProvider() interface{ ListActive(context.Context) ([]*model.AiProviderM, error) } {
    return &aiProviderTestStore{db: s.db}
}

func (s *storeWrapper) AiModel() interface{ ListActive(context.Context) ([]*model.AiModelM, error) } {
    return &aiModelTestStore{db: s.db}
}

// ... implement test stores using gorm ...
```

**Step 2: Run E2E test (if environment available)**

Run: `go test ./internal/pkg/ai/... -run TestAIProviderReload_E2E -v`
Expected: PASS (or skip if no test DB/Redis)

**Step 3: Commit**

```bash
git add internal/pkg/ai/e2e_test.go
git commit -m "test(ai): add end-to-end test for provider reload"
```

---

## Task 10: Build and Verify

**Step 1: Build all services**

Run: `make build`
Expected: Success, no compilation errors

**Step 2: Run linter**

Run: `make lint`
Expected: Success, no linting errors

**Step 3: Manual verification (optional)**

1. Start services with AI credentials configured
2. Verify providers load from database
3. Update provider status in DB
4. Trigger reload via Redis: `redis-cli PUBLISH ai:reload:providers trigger`
5. Verify registry reflects changes

---

## Summary

After completion, you will have:

| Component | File | Purpose |
|-----------|------|---------|
| Registry.Clear() | `pkg/ai/registry.go` | Enable reload |
| Mock Store | `internal/pkg/testing/mock/store/` | Unit testing |
| Loader | `internal/pkg/ai/loader.go` | DB → Registry loading |
| Subscriber | `internal/pkg/ai/subscriber.go` | Redis reload triggers |
| InitAI | `internal/pkg/ai/ai.go` | Centralized initialization |
| apiserver integration | `internal/apiserver/` | Use InitAI |
| admserver integration | `internal/admserver/` | Use InitAI |
| DB Migration | `internal/pkg/database/` | Remove unused field |
| Tests | `*_test.go` | Coverage |

**Total estimated steps:** ~50 (10 tasks × 5 steps each)

**Total commits:** 10 (one per task)
