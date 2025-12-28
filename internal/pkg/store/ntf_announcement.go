// ABOUTME: Store layer for announcements.
// ABOUTME: Provides CRUD and read status operations for system announcements.

package store

import (
	"context"
	"time"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfAnnouncementStore interface {
	Create(ctx context.Context, obj *model.NtfAnnouncementM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfAnnouncementM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error)
	Update(ctx context.Context, obj *model.NtfAnnouncementM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error

	NtfAnnouncementExpansion
}

type NtfAnnouncementExpansion interface {
	GetByUUID(ctx context.Context, uuid string) (*model.NtfAnnouncementM, error)
	ListPublished(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error)
	IsRead(ctx context.Context, userID string, announcementID uint64) (bool, error)
	MarkAsRead(ctx context.Context, userID string, announcementID uint64) error
	CountUnreadForUser(ctx context.Context, userID string) (int64, error)
}

type ntfAnnouncementStore struct {
	*genericstore.Store[model.NtfAnnouncementM]
}

var _ NtfAnnouncementStore = (*ntfAnnouncementStore)(nil)

func NewNtfAnnouncementStore(ds *datastore) *ntfAnnouncementStore {
	return &ntfAnnouncementStore{
		Store: genericstore.NewStore[model.NtfAnnouncementM](ds, NewLogger()),
	}
}

func (s *ntfAnnouncementStore) GetByUUID(ctx context.Context, uuid string) (*model.NtfAnnouncementM, error) {
	return s.Get(ctx, where.F("uuid", uuid))
}

func (s *ntfAnnouncementStore) ListPublished(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error) {
	db := s.DB(ctx, opts).
		Where("status = ?", string(model.AnnouncementStatusPublished)).
		Where("expires_at IS NULL OR expires_at > ?", time.Now())

	var ret []*model.NtfAnnouncementM
	var count int64
	err := db.Order("created_at desc").Find(&ret).Offset(-1).Limit(-1).Count(&count).Error

	return count, ret, err
}

func (s *ntfAnnouncementStore) IsRead(ctx context.Context, userID string, announcementID uint64) (bool, error) {
	var count int64
	err := s.DB(ctx).Model(&model.NtfAnnouncementReadM{}).
		Where("user_id = ? AND announcement_id = ?", userID, announcementID).
		Count(&count).Error

	return count > 0, err
}

func (s *ntfAnnouncementStore) MarkAsRead(ctx context.Context, userID string, announcementID uint64) error {
	read := &model.NtfAnnouncementReadM{
		UserID:         userID,
		AnnouncementID: announcementID,
		ReadAt:         time.Now(),
	}

	return s.DB(ctx).Create(read).Error
}

func (s *ntfAnnouncementStore) CountUnreadForUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.DB(ctx).Model(&model.NtfAnnouncementM{}).
		Where("status = ?", string(model.AnnouncementStatusPublished)).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Where("id NOT IN (?)",
			s.DB(ctx).Model(&model.NtfAnnouncementReadM{}).
				Select("announcement_id").
				Where("user_id = ?", userID),
		).
		Count(&count).Error

	return count, err
}
