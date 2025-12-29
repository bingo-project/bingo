// ABOUTME: Registry unit tests.
// ABOUTME: Tests provider registration and model lookup.

package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider implements Provider for testing
type mockProvider struct {
	name   string
	models []ModelInfo
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{}, nil
}

func (m *mockProvider) ChatStream(ctx context.Context, req *ChatRequest) (*ChatStream, error) {
	return NewChatStream(10), nil
}

func (m *mockProvider) Models() []ModelInfo {
	return m.models
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{
		name: "test",
		models: []ModelInfo{
			{ID: "test-model", Name: "Test Model", Provider: "test"},
		},
	}

	r.Register(p)

	got, ok := r.Get("test")
	require.True(t, ok)
	assert.Equal(t, "test", got.Name())
}

func TestRegistry_GetByModel(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{
		name: "test",
		models: []ModelInfo{
			{ID: "test-model", Name: "Test Model", Provider: "test"},
		},
	}

	r.Register(p)

	got, ok := r.GetByModel("test-model")
	require.True(t, ok)
	assert.Equal(t, "test", got.Name())
}

func TestRegistry_ListModels(t *testing.T) {
	r := NewRegistry()
	p1 := &mockProvider{
		name: "provider1",
		models: []ModelInfo{
			{ID: "model1", Name: "Model 1", Provider: "provider1"},
		},
	}
	p2 := &mockProvider{
		name: "provider2",
		models: []ModelInfo{
			{ID: "model2", Name: "Model 2", Provider: "provider2"},
		},
	}

	r.Register(p1)
	r.Register(p2)

	models := r.ListModels()
	assert.Len(t, models, 2)
}
