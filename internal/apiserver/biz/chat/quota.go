// ABOUTME: Token quota management for AI chat.
// ABOUTME: Provides TPD (Tokens Per Day) quota checking and tracking with Redis atomic operations.

package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

const (
	// defaultEstimatedTokens is the default token estimate for quota reservation
	// when MaxTokens is not specified in the request.
	defaultEstimatedTokens = 4096

	// quotaKeyTTL is the TTL for Redis quota keys (25 hours to cover full day + buffer)
	quotaKeyTTL = 25 * time.Hour
)

// quotaChecker handles token quota validation and tracking.
type quotaChecker struct {
	ds store.IStore
}

func newQuotaChecker(ds store.IStore) *quotaChecker {
	return &quotaChecker{ds: ds}
}

// ensureQuotaState ensures the user's daily quota is reset in DB if needed,
// and returns the current used tokens from DB for Redis initialization.
func (q *quotaChecker) ensureQuotaState(ctx context.Context, uid string) (int, int, error) {
	quota, tpd, err := q.getUserQuota(ctx, uid)
	if err != nil {
		return 0, 0, err
	}

	// Check reset
	if q.shouldResetDaily(quota) {
		if err := q.ds.AiUserQuota().ResetDailyTokens(ctx, uid); err != nil {
			return 0, 0, errno.ErrOperationFailed.WithMessage("failed to reset daily quota: %v", err)
		}
		quota.UsedTokensToday = 0
	}

	return quota.UsedTokensToday, tpd, nil
}

// ReserveTPD atomically reserves tokens before an API call.
// Uses Redis INCRBY for atomic check-and-reserve to prevent race conditions.
// Returns the number of tokens reserved.
func (q *quotaChecker) ReserveTPD(ctx context.Context, uid string, estimatedTokens int) (int, error) {
	if !facade.Config.AI.Quota.Enabled {
		return 0, nil
	}

	// Use default estimate if not provided
	if estimatedTokens <= 0 {
		estimatedTokens = defaultEstimatedTokens
	}

	key := q.buildQuotaKey(uid)

	var tpd int
	var err error

	// 1. Check/Init Redis
	exists, err := facade.Redis.Exists(ctx, key).Result()
	if err != nil {
		return 0, errno.ErrOperationFailed.WithMessage("redis error: %v", err)
	}

	if exists == 0 {
		// Key missing: Initialize from DB
		var usedInDB int
		// Reuse tpd from this call
		usedInDB, tpd, err = q.ensureQuotaState(ctx, uid)
		if err != nil {
			return 0, err
		}

		// Atomically set initial value if not exists (NX)
		_, err := facade.Redis.SetNX(ctx, key, usedInDB, quotaKeyTTL).Result()
		if err != nil {
			return 0, errno.ErrOperationFailed.WithMessage("redis setnx error: %v", err)
		}

		// Optimistic check: if we just initialized, check limit immediately before incrementing
		if usedInDB+estimatedTokens > tpd {
			return 0, errno.ErrAIQuotaExceeded
		}
	}

	// 2. Increment
	newTotal, err := facade.Redis.IncrBy(ctx, key, int64(estimatedTokens)).Result()
	if err != nil {
		return 0, errno.ErrOperationFailed.WithMessage("failed to reserve quota: %v", err)
	}

	// Refresh TTL
	facade.Redis.Expire(ctx, key, quotaKeyTTL)

	// 3. Check Limit
	// Only fetch TPD if we haven't already (i.e., Redis key existed)
	if tpd == 0 {
		_, tpd, err = q.getUserQuota(ctx, uid)
		if err != nil {
			// Rollback
			if err := facade.Redis.DecrBy(ctx, key, int64(estimatedTokens)).Err(); err != nil {
				log.C(ctx).Errorw("Failed to rollback quota reservation", "uid", uid, "err", err)
			}

			return 0, err
		}
	}

	if int(newTotal) > tpd {
		// Rollback the reservation
		if err := facade.Redis.DecrBy(ctx, key, int64(estimatedTokens)).Err(); err != nil {
			log.C(ctx).Errorw("Failed to rollback quota reservation", "uid", uid, "err", err)
		}

		return 0, errno.ErrAIQuotaExceeded.WithMessage("daily token quota exceeded (%d/%d)", int(newTotal)-estimatedTokens, tpd)
	}

	return estimatedTokens, nil
}

// AdjustTPD adjusts the quota after API call completes with actual token usage.
// It adjusts the difference between actual and reserved tokens in Redis,
// and persists the actual usage to database.
func (q *quotaChecker) AdjustTPD(ctx context.Context, uid string, actualTokens, reservedTokens int) error {
	if !facade.Config.AI.Quota.Enabled {
		return nil
	}

	// Adjust Redis count
	diff := actualTokens - reservedTokens
	if diff != 0 {
		key := q.buildQuotaKey(uid)
		var err error
		if diff > 0 {
			err = facade.Redis.IncrBy(ctx, key, int64(diff)).Err()
		} else {
			err = facade.Redis.DecrBy(ctx, key, int64(-diff)).Err()
		}
		if err != nil {
			log.C(ctx).Errorw("Failed to adjust Redis quota", "uid", uid, "diff", diff, "err", err)
			// Continue to persist DB usage even if Redis fails, to ensure billing accuracy
		}
	}

	// Persist to database for statistics (only actual tokens)
	if actualTokens > 0 {
		// Ensure user quota record exists
		if _, _, err := q.getUserQuota(ctx, uid); err != nil {
			return err
		}

		return q.ds.AiUserQuota().IncrementTokens(ctx, uid, actualTokens)
	}

	return nil
}

// buildQuotaKey builds the Redis key for daily quota tracking.
func (q *quotaChecker) buildQuotaKey(uid string) string {
	date := time.Now().Format("2006-01-02")

	return fmt.Sprintf("%s:ai:tpd:%s:%s", facade.Config.App.Name, uid, date)
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
