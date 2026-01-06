// ABOUTME: Tests for AI model fallback selection.
// ABOUTME: Verifies fallback model selection logic.

package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bingo-project/bingo/internal/pkg/model"
	aipkg "github.com/bingo-project/bingo/pkg/ai"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// mockModelStore is a minimal mock for testing
type mockModelStore struct {
	models []*model.AiModelM
	err    error
}

func (m *mockModelStore) ListActive(context.Context) ([]*model.AiModelM, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.models, nil
}

// Implement other required methods as no-ops
func (m *mockModelStore) Create(ctx context.Context, obj *model.AiModelM) error { return nil }
func (m *mockModelStore) Update(ctx context.Context, obj *model.AiModelM, fields ...string) error {
	return nil
}
func (m *mockModelStore) Delete(ctx context.Context, opts *where.Options) error { return nil }
func (m *mockModelStore) Get(ctx context.Context, opts *where.Options) (*model.AiModelM, error) {
	return nil, nil
}
func (m *mockModelStore) List(ctx context.Context, opts *where.Options) (int64, []*model.AiModelM, error) {
	return 0, nil, nil
}
func (m *mockModelStore) GetByModel(ctx context.Context, modelID string) (*model.AiModelM, error) {
	return nil, nil
}
func (m *mockModelStore) ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error) {
	return nil, nil
}
func (m *mockModelStore) GetDefault(ctx context.Context) (*model.AiModelM, error) { return nil, nil }
func (m *mockModelStore) FirstOrCreate(ctx context.Context, where *model.AiModelM, obj *model.AiModelM) error {
	return nil
}

func TestFallbackSelector_SelectFallback_Success(t *testing.T) {
	registry := aipkg.NewRegistry()

	// Register test providers
	registry.Register(&mockProvider{name: "openai", models: []aipkg.ModelInfo{
		{ID: "gpt-4o", Provider: "openai"},
		{ID: "gpt-3.5-turbo", Provider: "openai"},
	}})

	store := &mockModelStore{
		models: []*model.AiModelM{
			{Model: "gpt-4o", AllowFallback: true, Sort: 1, Status: model.AiModelStatusActive},
			{Model: "gpt-3.5-turbo", AllowFallback: true, Sort: 2, Status: model.AiModelStatusActive},
			{Model: "claude-3-5-sonnet", AllowFallback: false, Sort: 3, Status: model.AiModelStatusActive},
		},
	}

	selector := NewFallbackSelector(store, registry)

	fallback := selector.SelectFallback(context.Background(), "gpt-4o")

	assert.Equal(t, "gpt-3.5-turbo", fallback, "should return next model by sort order")
}

func TestFallbackSelector_SelectFallback_NoFallbackAllowed(t *testing.T) {
	registry := aipkg.NewRegistry()
	registry.Register(&mockProvider{name: "openai", models: []aipkg.ModelInfo{
		{ID: "gpt-4o", Provider: "openai"},
	}})

	store := &mockModelStore{
		models: []*model.AiModelM{
			{Model: "gpt-4o", AllowFallback: false, Sort: 1, Status: model.AiModelStatusActive},
		},
	}

	selector := NewFallbackSelector(store, registry)

	fallback := selector.SelectFallback(context.Background(), "gpt-4o")

	assert.Equal(t, "", fallback, "should return empty when no fallback allowed")
}

func TestFallbackSelector_SelectFallback_NoModelsRegistered(t *testing.T) {
	registry := aipkg.NewRegistry()

	store := &mockModelStore{
		models: []*model.AiModelM{
			{Model: "gpt-4o", AllowFallback: true, Sort: 1, Status: model.AiModelStatusActive},
		},
	}

	selector := NewFallbackSelector(store, registry)

	fallback := selector.SelectFallback(context.Background(), "unknown-model")

	assert.Equal(t, "", fallback, "should return empty when model not found")
}

func TestFallbackSelector_SelectFallback_StoreError(t *testing.T) {
	registry := aipkg.NewRegistry()

	store := &mockModelStore{
		err: assert.AnError,
	}

	selector := NewFallbackSelector(store, registry)

	fallback := selector.SelectFallback(context.Background(), "gpt-4o")

	assert.Equal(t, "", fallback, "should return empty on store error")
}

// mockProvider is a minimal mock Provider
type mockProvider struct {
	name   string
	models []aipkg.ModelInfo
}

func (m *mockProvider) Name() string              { return m.name }
func (m *mockProvider) Models() []aipkg.ModelInfo { return m.models }
func (m *mockProvider) Chat(ctx context.Context, req *aipkg.ChatRequest) (*aipkg.ChatResponse, error) {
	return nil, nil
}
func (m *mockProvider) ChatStream(ctx context.Context, req *aipkg.ChatRequest) (*aipkg.ChatStream, error) {
	return nil, nil
}
