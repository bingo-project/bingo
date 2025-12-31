// ABOUTME: Mock AI provider store for testing.
// ABOUTME: Provides in-memory implementation of AiProviderStore interface.

package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AiProviderStore implements store.AiProviderStore for testing.
type AiProviderStore struct {
	// Configurable results for testing
	ListActiveResult []*model.AiProviderM
	ListActiveErr    error
	ListActiveCalled bool

	GetByNameResult  *model.AiProviderM
	GetByNameErr     error
	GetByNameCalled  bool
	GetByNameArgName string

	GetDefaultResult *model.AiProviderM
	GetDefaultErr    error
	GetDefaultCalled bool

	FirstOrCreateResult *model.AiProviderM
	FirstOrCreateErr    error
	FirstOrCreateCalled bool
	FirstOrCreateArg    *model.AiProviderM

	// In-memory storage
	providers map[uint]*model.AiProviderM
	nextID    uint
}

var _ store.AiProviderStore = (*AiProviderStore)(nil)

// NewAiProviderStore creates a new mock AI provider store.
func NewAiProviderStore() *AiProviderStore {
	return &AiProviderStore{
		providers: make(map[uint]*model.AiProviderM),
		nextID:    1,
	}
}

// Create creates a new AI provider.
func (m *AiProviderStore) Create(ctx context.Context, obj *model.AiProviderM) error {
	if obj.ID == 0 {
		obj.ID = m.nextID
		m.nextID++
	}
	m.providers[obj.ID] = obj

	return nil
}

// Update updates an AI provider.
func (m *AiProviderStore) Update(ctx context.Context, obj *model.AiProviderM, fields ...string) error {
	if _, exists := m.providers[obj.ID]; !exists {
		return gorm.ErrRecordNotFound
	}
	m.providers[obj.ID] = obj

	return nil
}

// Delete deletes AI providers by options.
func (m *AiProviderStore) Delete(ctx context.Context, opts *where.Options) error {
	for id := range m.providers {
		delete(m.providers, id)
	}

	return nil
}

// Get gets an AI provider by options.
func (m *AiProviderStore) Get(ctx context.Context, opts *where.Options) (*model.AiProviderM, error) {
	if len(m.providers) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	for _, p := range m.providers {
		return p, nil
	}

	return nil, gorm.ErrRecordNotFound
}

// List lists AI providers by options.
func (m *AiProviderStore) List(ctx context.Context, opts *where.Options) (int64, []*model.AiProviderM, error) {
	var providers []*model.AiProviderM
	for _, p := range m.providers {
		providers = append(providers, p)
	}

	return int64(len(providers)), providers, nil
}

// GetByName gets an AI provider by name.
func (m *AiProviderStore) GetByName(ctx context.Context, name string) (*model.AiProviderM, error) {
	m.GetByNameCalled = true
	m.GetByNameArgName = name

	if m.GetByNameResult != nil || m.GetByNameErr != nil {
		return m.GetByNameResult, m.GetByNameErr
	}

	for _, p := range m.providers {
		if p.Name == name {
			return p, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// ListActive lists all active AI providers.
func (m *AiProviderStore) ListActive(ctx context.Context) ([]*model.AiProviderM, error) {
	m.ListActiveCalled = true

	if m.ListActiveResult != nil || m.ListActiveErr != nil {
		return m.ListActiveResult, m.ListActiveErr
	}

	var providers []*model.AiProviderM
	for _, p := range m.providers {
		if p.Status == model.AiProviderStatusActive {
			providers = append(providers, p)
		}
	}

	return providers, nil
}

// GetDefault gets the default AI provider.
func (m *AiProviderStore) GetDefault(ctx context.Context) (*model.AiProviderM, error) {
	m.GetDefaultCalled = true

	if m.GetDefaultResult != nil || m.GetDefaultErr != nil {
		return m.GetDefaultResult, m.GetDefaultErr
	}

	for _, p := range m.providers {
		if p.Status == model.AiProviderStatusActive && p.IsDefault {
			return p, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// FirstOrCreate gets the first AI provider matching where or creates a new one.
func (m *AiProviderStore) FirstOrCreate(ctx context.Context, whereClause *model.AiProviderM, obj *model.AiProviderM) error {
	m.FirstOrCreateCalled = true
	m.FirstOrCreateArg = whereClause

	if m.FirstOrCreateErr != nil {
		return m.FirstOrCreateErr
	}

	if m.FirstOrCreateResult != nil {
		*obj = *m.FirstOrCreateResult

		return nil
	}

	// Simple implementation: check if provider with same name exists
	for _, p := range m.providers {
		if p.Name == whereClause.Name {
			*obj = *p

			return nil
		}
	}

	// Create new
	if obj.ID == 0 {
		obj.ID = m.nextID
		m.nextID++
	}
	m.providers[obj.ID] = obj

	return nil
}
