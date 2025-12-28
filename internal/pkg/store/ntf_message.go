// ABOUTME: Store layer for notification messages.
// ABOUTME: Provides CRUD operations for personal notifications.

package store

import (
	"context"
	"time"

	genericstore "github.com/bingo-project/bingo/pkg/store"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfMessageStore interface {
	Create(ctx context.Context, obj *model.NtfMessageM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfMessageM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.NtfMessageM, error)
	Update(ctx context.Context, obj *model.NtfMessageM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error

	NtfMessageExpansion
}

type NtfMessageExpansion interface {
	GetByUUID(ctx context.Context, uuid string) (*model.NtfMessageM, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkAsRead(ctx context.Context, userID string, uuid string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type ntfMessageStore struct {
	*genericstore.Store[model.NtfMessageM]
}

var _ NtfMessageStore = (*ntfMessageStore)(nil)

func NewNtfMessageStore(ds *datastore) *ntfMessageStore {
	return &ntfMessageStore{
		Store: genericstore.NewStore[model.NtfMessageM](ds, NewLogger()),
	}
}

func (s *ntfMessageStore) GetByUUID(ctx context.Context, uuid string) (*model.NtfMessageM, error) {
	return s.Get(ctx, where.F("uuid", uuid))
}

func (s *ntfMessageStore) CountUnread(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.DB(ctx).Model(&model.NtfMessageM{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (s *ntfMessageStore) MarkAsRead(ctx context.Context, userID string, uuid string) error {
	return s.DB(ctx).Model(&model.NtfMessageM{}).
		Where("user_id = ? AND uuid = ?", userID, uuid).
		Updates(map[string]any{"is_read": true, "read_at": time.Now()}).Error
}

func (s *ntfMessageStore) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.DB(ctx).Model(&model.NtfMessageM{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{"is_read": true, "read_at": time.Now()}).Error
}
