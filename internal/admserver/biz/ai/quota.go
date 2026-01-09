// ABOUTME: AI Quota business logic for admin management.
// ABOUTME: Provides list/get/update/reset operations for user quota resources.
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

// AiQuotaBiz defines AI quota management interface for admin.
type AiQuotaBiz interface {
	List(ctx context.Context, req *v1.ListAiUserQuotaRequest) (*v1.ListAiUserQuotaResponse, error)
	Get(ctx context.Context, uid string) (*v1.AiUserQuotaInfo, error)
	Update(ctx context.Context, uid string, req *v1.UpdateAiUserQuotaRequest) (*v1.AiUserQuotaInfo, error)
	ResetDailyTokens(ctx context.Context, uid string) error
}

type aiQuotaBiz struct {
	ds store.IStore
}

var _ AiQuotaBiz = (*aiQuotaBiz)(nil)

func NewAiQuota(ds store.IStore) AiQuotaBiz {
	return &aiQuotaBiz{ds: ds}
}

// toQuotaInfo converts model.AiUserQuotaM to v1.AiUserQuotaInfo.
func toQuotaInfo(m *model.AiUserQuotaM) *v1.AiUserQuotaInfo {
	return &v1.AiUserQuotaInfo{
		UID:             m.UID,
		Tier:            m.Tier,
		RPM:             m.RPM,
		TPD:             m.TPD,
		UsedTokensToday: m.UsedTokensToday,
		LastResetAt:     m.LastResetAt,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

func (b *aiQuotaBiz) List(ctx context.Context, req *v1.ListAiUserQuotaRequest) (*v1.ListAiUserQuotaResponse, error) {
	// Default pagination
	page := 1
	pageSize := 20
	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 {
		pageSize = req.PageSize
	}

	// Build where clause
	var opts *where.Options
	if req.Tier != "" {
		opts = where.P(page, pageSize).F("tier", req.Tier)
	} else if req.UID != "" {
		opts = where.P(page, pageSize).F("uid", req.UID+"%")
	} else {
		opts = where.P(page, pageSize)
	}

	total, quotas, err := b.ds.AiUserQuota().List(ctx, opts)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list ai user quotas: %v", err)
	}

	data := make([]v1.AiUserQuotaInfo, len(quotas))
	for i, q := range quotas {
		data[i] = *toQuotaInfo(q)
	}

	return &v1.ListAiUserQuotaResponse{
		Total: total,
		Data:  data,
	}, nil
}

func (b *aiQuotaBiz) Get(ctx context.Context, uid string) (*v1.AiUserQuotaInfo, error) {
	quota, err := b.ds.AiUserQuota().GetByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrNotFound.WithMessage("user quota not found")
		}

		return nil, errno.ErrDBRead.WithMessage("get ai user quota: %v", err)
	}

	return toQuotaInfo(quota), nil
}

func (b *aiQuotaBiz) Update(ctx context.Context, uid string, req *v1.UpdateAiUserQuotaRequest) (*v1.AiUserQuotaInfo, error) {
	quota, err := b.ds.AiUserQuota().GetByUID(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrNotFound.WithMessage("user quota not found")
		}

		return nil, errno.ErrDBRead.WithMessage("get ai user quota: %v", err)
	}

	// Update fields
	if req.Tier != "" {
		quota.Tier = req.Tier
	}
	if req.RPM != nil {
		quota.RPM = *req.RPM
	}
	if req.TPD != nil {
		quota.TPD = *req.TPD
	}

	if err := b.ds.AiUserQuota().Update(ctx, quota); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai user quota: %v", err)
	}

	log.C(ctx).Infow("ai user quota updated", "uid", quota.UID, "tier", quota.Tier)

	return toQuotaInfo(quota), nil
}

func (b *aiQuotaBiz) ResetDailyTokens(ctx context.Context, uid string) error {
	err := b.ds.AiUserQuota().ResetDailyTokens(ctx, uid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrNotFound.WithMessage("user quota not found")
		}

		return errno.ErrDBWrite.WithMessage("reset ai user quota: %v", err)
	}

	log.C(ctx).Infow("ai user quota daily tokens reset", "uid", uid)

	return nil
}
