// ABOUTME: AI Provider business logic for admin management.
// ABOUTME: Provides list/get/update operations for AI Provider resources.
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

// AiProviderBiz defines AI provider management interface for admin.
type AiProviderBiz interface {
	List(ctx context.Context, req *v1.ListAiProviderRequest) (*v1.ListAiProviderResponse, error)
	Get(ctx context.Context, id uint) (*v1.AiProviderInfo, error)
	Update(ctx context.Context, id uint, req *v1.UpdateAiProviderRequest) (*v1.AiProviderInfo, error)
}

type aiProviderBiz struct {
	ds store.IStore
}

var _ AiProviderBiz = (*aiProviderBiz)(nil)

func NewAiProvider(ds store.IStore) AiProviderBiz {
	return &aiProviderBiz{ds: ds}
}

// toProviderInfo converts model.AiProviderM to v1.AiProviderInfo.
func toProviderInfo(m *model.AiProviderM) *v1.AiProviderInfo {
	return &v1.AiProviderInfo{
		ID:          m.ID,
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Status:      string(m.Status),
		IsDefault:   m.IsDefault,
		Sort:        m.Sort,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func (b *aiProviderBiz) List(ctx context.Context, req *v1.ListAiProviderRequest) (*v1.ListAiProviderResponse, error) {
	var providers []*model.AiProviderM
	var err error

	if req.Status == string(model.AiProviderStatusActive) {
		providers, err = b.ds.AiProvider().ListActive(ctx)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai providers: %v", err)
		}
	} else if req.Status == string(model.AiProviderStatusDisabled) {
		// Query all and filter for disabled
		var all []*model.AiProviderM
		_, all, err = b.ds.AiProvider().List(ctx, nil)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai providers: %v", err)
		}
		providers = make([]*model.AiProviderM, 0)
		for _, p := range all {
			if p.Status == model.AiProviderStatusDisabled {
				providers = append(providers, p)
			}
		}
	} else {
		// List all
		_, providers, err = b.ds.AiProvider().List(ctx, nil)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list ai providers: %v", err)
		}
	}

	data := make([]v1.AiProviderInfo, len(providers))
	for i, p := range providers {
		data[i] = *toProviderInfo(p)
	}

	return &v1.ListAiProviderResponse{
		Total: int64(len(providers)),
		Data:  data,
	}, nil
}

func (b *aiProviderBiz) Get(ctx context.Context, id uint) (*v1.AiProviderInfo, error) {
	provider, err := b.ds.AiProvider().Get(ctx, where.F("id", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIProviderNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai provider: %v", err)
	}

	return toProviderInfo(provider), nil
}

func (b *aiProviderBiz) Update(ctx context.Context, id uint, req *v1.UpdateAiProviderRequest) (*v1.AiProviderInfo, error) {
	provider, err := b.ds.AiProvider().Get(ctx, where.F("id", id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIProviderNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai provider: %v", err)
	}

	// Update fields
	if req.DisplayName != "" {
		provider.DisplayName = req.DisplayName
	}
	if req.Status != "" {
		provider.Status = model.AiProviderStatus(req.Status)
	}
	if req.IsDefault != nil {
		provider.IsDefault = *req.IsDefault
	}
	if req.Sort != nil {
		provider.Sort = *req.Sort
	}

	if err := b.ds.AiProvider().Update(ctx, provider); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai provider: %v", err)
	}

	log.C(ctx).Infow("ai provider updated", "id", provider.ID, "name", provider.Name)

	return toProviderInfo(provider), nil
}
