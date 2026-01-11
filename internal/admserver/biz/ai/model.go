// ABOUTME: AI Model business logic for admin management.
// ABOUTME: Provides list/get/update operations for AI Model resources.
package ai

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AiModelBiz defines AI model management interface for admin.
type AiModelBiz interface {
	Create(ctx context.Context, req *v1.CreateAiModelRequest) (*v1.AiModelInfo, error)
	Get(ctx context.Context, id uint) (*v1.AiModelInfo, error)
	List(ctx context.Context, req *v1.ListAiModelRequest) (*v1.ListAiModelResponse, error)
	Update(ctx context.Context, id uint, req *v1.UpdateAiModelRequest) (*v1.AiModelInfo, error)
	Delete(ctx context.Context, id uint) error
}

type aiModelBiz struct {
	ds store.IStore
}

var _ AiModelBiz = (*aiModelBiz)(nil)

func NewAiModel(ds store.IStore) AiModelBiz {
	return &aiModelBiz{ds: ds}
}

// toModelInfo converts model.AiModelM to v1.AiModelInfo.
func toModelInfo(m *model.AiModelM) *v1.AiModelInfo {
	return &v1.AiModelInfo{
		ID:            m.ID,
		ProviderName:  m.ProviderName,
		Model:         m.Model,
		DisplayName:   m.DisplayName,
		MaxTokens:     m.MaxTokens,
		InputPrice:    m.InputPrice,
		OutputPrice:   m.OutputPrice,
		Status:        string(m.Status),
		IsDefault:     m.IsDefault,
		Sort:          m.Sort,
		AllowFallback: m.AllowFallback,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func (b *aiModelBiz) Create(ctx context.Context, req *v1.CreateAiModelRequest) (*v1.AiModelInfo, error) {
	// Check if model already exists
	existing, err := b.ds.AiModel().GetByProviderAndModel(ctx, req.ProviderName, req.Model)
	if err == nil && existing != nil {
		return nil, errno.ErrResourceAlreadyExists.WithMessage(
			"ai model already exists: provider=%s, model=%s", req.ProviderName, req.Model)
	}

	// Set default values
	status := model.AiModelStatusActive
	if req.Status != "" {
		status = model.AiModelStatus(req.Status)
	}

	maxTokens := 4096
	if req.MaxTokens > 0 {
		maxTokens = req.MaxTokens
	}

	aiModel := &model.AiModelM{
		ProviderName:  req.ProviderName,
		Model:         req.Model,
		DisplayName:   req.DisplayName,
		MaxTokens:     maxTokens,
		InputPrice:    req.InputPrice,
		OutputPrice:   req.OutputPrice,
		Status:        status,
		IsDefault:     req.IsDefault,
		Sort:          req.Sort,
		AllowFallback: req.AllowFallback,
	}

	if err := b.ds.AiModel().Create(ctx, aiModel); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create ai model: %v", err)
	}

	log.C(ctx).Infow("ai model created", "id", aiModel.ID, "provider", aiModel.ProviderName, "model", aiModel.Model)

	return toModelInfo(aiModel), nil
}

func (b *aiModelBiz) List(ctx context.Context, req *v1.ListAiModelRequest) (*v1.ListAiModelResponse, error) {
	var models []*model.AiModelM
	var total int64
	var err error

	if req.ProviderName != "" {
		models, err = b.ds.AiModel().ListByProvider(ctx, req.ProviderName)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai models: %v", err)
		}
		total = int64(len(models))
	} else if req.Status == string(model.AiModelStatusActive) {
		models, err = b.ds.AiModel().ListActive(ctx)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai models: %v", err)
		}
		total = int64(len(models))
	} else if req.Status == string(model.AiModelStatusDisabled) {
		// Query all and filter for disabled
		var all []*model.AiModelM
		total, all, err = b.ds.AiModel().List(ctx, nil)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai models: %v", err)
		}
		models = make([]*model.AiModelM, 0)
		for _, m := range all {
			if m.Status == model.AiModelStatusDisabled {
				models = append(models, m)
			}
		}
		total = int64(len(models))
	} else {
		// List all
		total, models, err = b.ds.AiModel().List(ctx, nil)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai models: %v", err)
		}
	}

	data := make([]v1.AiModelInfo, len(models))
	for i, m := range models {
		data[i] = *toModelInfo(m)
	}

	return &v1.ListAiModelResponse{
		Total: total,
		Data:  data,
	}, nil
}

func (b *aiModelBiz) Get(ctx context.Context, id uint) (*v1.AiModelInfo, error) {
	aiModel, err := b.ds.AiModel().Get(ctx, where.F("id", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIModelNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai model: %v", err)
	}

	return toModelInfo(aiModel), nil
}

func (b *aiModelBiz) Update(ctx context.Context, id uint, req *v1.UpdateAiModelRequest) (*v1.AiModelInfo, error) {
	aiModel, err := b.ds.AiModel().Get(ctx, where.F("id", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIModelNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai model: %v", err)
	}

	// Update fields
	if req.DisplayName != "" {
		aiModel.DisplayName = req.DisplayName
	}
	if req.MaxTokens != nil {
		aiModel.MaxTokens = *req.MaxTokens
	}
	if req.InputPrice != nil {
		aiModel.InputPrice = *req.InputPrice
	}
	if req.OutputPrice != nil {
		aiModel.OutputPrice = *req.OutputPrice
	}
	if req.Status != "" {
		aiModel.Status = model.AiModelStatus(req.Status)
	}
	if req.IsDefault != nil {
		aiModel.IsDefault = *req.IsDefault
	}
	if req.Sort != nil {
		aiModel.Sort = *req.Sort
	}
	if req.AllowFallback != nil {
		aiModel.AllowFallback = *req.AllowFallback
	}

	if err := b.ds.AiModel().Update(ctx, aiModel); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai model: %v", err)
	}

	log.C(ctx).Infow("ai model updated", "id", aiModel.ID, "model", aiModel.Model)

	return toModelInfo(aiModel), nil
}

func (b *aiModelBiz) Delete(ctx context.Context, id uint) error {
	aiModel, err := b.ds.AiModel().Get(ctx, where.F("id", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAIModelNotFound
		}

		return errno.ErrDBRead.WithMessage("get ai model: %v", err)
	}

	// Soft delete: set status to disabled
	aiModel.Status = model.AiModelStatusDisabled
	if err := b.ds.AiModel().Update(ctx, aiModel, "status"); err != nil {
		return errno.ErrDBWrite.WithMessage("delete ai model: %v", err)
	}

	log.C(ctx).Infow("ai model deleted", "id", aiModel.ID, "provider", aiModel.ProviderName, "model", aiModel.Model)

	return nil
}
