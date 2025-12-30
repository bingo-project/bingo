// ABOUTME: Token quota management for AI chat.
// ABOUTME: Provides TPD (Tokens Per Day) quota checking and tracking.

package chat

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// quotaChecker handles token quota validation and tracking.
type quotaChecker struct {
	ds store.IStore
}

func newQuotaChecker(ds store.IStore) *quotaChecker {
	return &quotaChecker{ds: ds}
}

// CheckTPD checks if user has remaining token quota for today.
// Returns nil if quota is available or quota checking is disabled.
func (q *quotaChecker) CheckTPD(ctx context.Context, uid string) error {
	if !facade.Config.AI.Quota.Enabled {
		return nil
	}

	quota, tpd, err := q.getUserQuota(ctx, uid)
	if err != nil {
		return err
	}

	// Check if need to reset daily tokens
	if q.shouldResetDaily(quota) {
		if err := q.ds.AiUserQuota().ResetDailyTokens(ctx, uid); err != nil {
			return errno.ErrOperationFailed.WithMessage("failed to reset daily quota: %v", err)
		}
		quota.UsedTokensToday = 0
	}

	if quota.UsedTokensToday >= tpd {
		return errno.ErrAIQuotaExceeded.WithMessage("daily token quota exceeded (%d/%d)", quota.UsedTokensToday, tpd)
	}

	return nil
}

// UpdateTPD updates the user's token usage for today.
func (q *quotaChecker) UpdateTPD(ctx context.Context, uid string, tokens int) error {
	if !facade.Config.AI.Quota.Enabled || tokens <= 0 {
		return nil
	}

	// Ensure user quota record exists
	if _, _, err := q.getUserQuota(ctx, uid); err != nil {
		return err
	}

	return q.ds.AiUserQuota().IncrementTokens(ctx, uid, tokens)
}

// getUserQuota retrieves user quota, creating default if not exists.
// Returns the user quota record and effective TPD limit.
func (q *quotaChecker) getUserQuota(ctx context.Context, uid string) (*model.AiUserQuotaM, int, error) {
	quota, err := q.ds.AiUserQuota().GetByUID(ctx, uid)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, 0, errno.ErrOperationFailed.WithMessage("failed to get quota: %v", err)
		}

		// Create default quota for new user
		quota = &model.AiUserQuotaM{
			UID:  uid,
			Tier: model.AiQuotaTierFree,
		}
		if err := q.ds.AiUserQuota().Create(ctx, quota); err != nil {
			return nil, 0, errno.ErrOperationFailed.WithMessage("failed to create quota: %v", err)
		}
	}

	// Determine effective TPD: user override > tier default > config default
	tpd := quota.TPD
	if tpd == 0 {
		// Get tier default
		tier, err := q.ds.AiQuotaTier().GetByTier(ctx, quota.Tier)
		if err == nil {
			tpd = tier.TPD
		}
	}
	if tpd == 0 {
		tpd = facade.Config.AI.Quota.DefaultTPD
	}
	if tpd == 0 {
		tpd = 100000 // fallback default
	}

	return quota, tpd, nil
}

// shouldResetDaily checks if daily tokens should be reset.
func (q *quotaChecker) shouldResetDaily(quota *model.AiUserQuotaM) bool {
	if quota.LastResetAt == nil {
		return true
	}

	now := time.Now()
	lastReset := *quota.LastResetAt

	// Reset if last reset was on a different day
	return now.Year() != lastReset.Year() ||
		now.YearDay() != lastReset.YearDay()
}
