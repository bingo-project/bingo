// ABOUTME: Mock AI model store for testing.
// ABOUTME: Provides in-memory implementation of AiModelStore interface.

package store

import (
	"context"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AiModelStore implements store.AiModelStore for testing.
type AiModelStore struct {
	// Configurable results for testing
	ListActiveResult []*model.AiModelM
	ListActiveErr    error
	ListActiveCalled bool

	ListByProviderResult  []*model.AiModelM
	ListByProviderErr     error
	ListByProviderCalled  bool
	ListByProviderArgName string

	GetByProviderAndModelResult   *model.AiModelM
	GetByProviderAndModelErr      error
	GetByProviderAndModelCalled   bool
	GetByProviderAndModelArgName  string
	GetByProviderAndModelArgModel string

	FindActiveByModelResult *model.AiModelM
	FindActiveByModelErr    error
	FindActiveByModelCalled bool
	FindActiveByModelArgID  string

	GetDefaultResult *model.AiModelM
	GetDefaultErr    error
	GetDefaultCalled bool

	FirstOrCreateResult *model.AiModelM
	FirstOrCreateErr    error
	FirstOrCreateCalled bool
	FirstOrCreateArg    *model.AiModelM

	// In-memory storage
	models map[uint]*model.AiModelM
	nextID uint
}

var _ store.AiModelStore = (*AiModelStore)(nil)

// NewAiModelStore creates a new mock AI model store.
func NewAiModelStore() *AiModelStore {
	return &AiModelStore{
		models: make(map[uint]*model.AiModelM),
		nextID: 1,
	}
}

// Create creates a new AI model.
func (m *AiModelStore) Create(ctx context.Context, obj *model.AiModelM) error {
	if obj.ID == 0 {
		obj.ID = m.nextID
		m.nextID++
	}
	m.models[obj.ID] = obj

	return nil
}

// Update updates an AI model.
func (m *AiModelStore) Update(ctx context.Context, obj *model.AiModelM, fields ...string) error {
	if _, exists := m.models[obj.ID]; !exists {
		return gorm.ErrRecordNotFound
	}
	m.models[obj.ID] = obj

	return nil
}

// Delete deletes AI models by options.
func (m *AiModelStore) Delete(ctx context.Context, opts *where.Options) error {
	for id := range m.models {
		delete(m.models, id)
	}

	return nil
}

// Get gets an AI model by options.
func (m *AiModelStore) Get(ctx context.Context, opts *where.Options) (*model.AiModelM, error) {
	if len(m.models) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	for _, aiModel := range m.models {
		return aiModel, nil
	}

	return nil, gorm.ErrRecordNotFound
}

// List lists AI models by options.
func (m *AiModelStore) List(ctx context.Context, opts *where.Options) (int64, []*model.AiModelM, error) {
	var models []*model.AiModelM
	for _, aiModel := range m.models {
		models = append(models, aiModel)
	}

	return int64(len(models)), models, nil
}

// GetByProviderAndModel gets an AI model by provider name and model ID.
func (m *AiModelStore) GetByProviderAndModel(ctx context.Context, providerName, modelID string) (*model.AiModelM, error) {
	m.GetByProviderAndModelCalled = true
	m.GetByProviderAndModelArgName = providerName
	m.GetByProviderAndModelArgModel = modelID

	if m.GetByProviderAndModelResult != nil || m.GetByProviderAndModelErr != nil {
		return m.GetByProviderAndModelResult, m.GetByProviderAndModelErr
	}

	for _, aiModel := range m.models {
		if aiModel.ProviderName == providerName && aiModel.Model == modelID {
			return aiModel, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// FindActiveByModel finds the first active AI model by model ID (sorted by priority).
func (m *AiModelStore) FindActiveByModel(ctx context.Context, modelID string) (*model.AiModelM, error) {
	m.FindActiveByModelCalled = true
	m.FindActiveByModelArgID = modelID

	if m.FindActiveByModelResult != nil || m.FindActiveByModelErr != nil {
		return m.FindActiveByModelResult, m.FindActiveByModelErr
	}

	for _, aiModel := range m.models {
		if aiModel.Model == modelID && aiModel.Status == model.AiModelStatusActive {
			return aiModel, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// ListByProvider lists all AI models for a provider.
func (m *AiModelStore) ListByProvider(ctx context.Context, providerName string) ([]*model.AiModelM, error) {
	m.ListByProviderCalled = true
	m.ListByProviderArgName = providerName

	if m.ListByProviderResult != nil || m.ListByProviderErr != nil {
		return m.ListByProviderResult, m.ListByProviderErr
	}

	var models []*model.AiModelM
	for _, aiModel := range m.models {
		if aiModel.ProviderName == providerName && aiModel.Status == model.AiModelStatusActive {
			models = append(models, aiModel)
		}
	}

	return models, nil
}

// ListActive lists all active AI models.
func (m *AiModelStore) ListActive(ctx context.Context) ([]*model.AiModelM, error) {
	m.ListActiveCalled = true

	if m.ListActiveResult != nil || m.ListActiveErr != nil {
		return m.ListActiveResult, m.ListActiveErr
	}

	var models []*model.AiModelM
	for _, aiModel := range m.models {
		if aiModel.Status == model.AiModelStatusActive {
			models = append(models, aiModel)
		}
	}

	return models, nil
}

// GetDefault gets the default AI model.
func (m *AiModelStore) GetDefault(ctx context.Context) (*model.AiModelM, error) {
	m.GetDefaultCalled = true

	if m.GetDefaultResult != nil || m.GetDefaultErr != nil {
		return m.GetDefaultResult, m.GetDefaultErr
	}

	for _, aiModel := range m.models {
		if aiModel.Status == model.AiModelStatusActive && aiModel.IsDefault {
			return aiModel, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// FirstOrCreate gets the first AI model matching where or creates a new one.
func (m *AiModelStore) FirstOrCreate(ctx context.Context, whereClause *model.AiModelM, obj *model.AiModelM) error {
	m.FirstOrCreateCalled = true
	m.FirstOrCreateArg = whereClause

	if m.FirstOrCreateErr != nil {
		return m.FirstOrCreateErr
	}

	if m.FirstOrCreateResult != nil {
		*obj = *m.FirstOrCreateResult

		return nil
	}

	// Simple implementation: check if model with same model ID exists
	for _, aiModel := range m.models {
		if aiModel.Model == whereClause.Model {
			*obj = *aiModel

			return nil
		}
	}

	// Create new
	if obj.ID == 0 {
		obj.ID = m.nextID
		m.nextID++
	}
	m.models[obj.ID] = obj

	return nil
}
