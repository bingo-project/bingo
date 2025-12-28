// ABOUTME: Store layer for notification preferences.
// ABOUTME: Provides CRUD operations for user notification settings.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfPreferenceStore interface {
	Create(ctx context.Context, obj *model.NtfPreferenceM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfPreferenceM, error)
	Update(ctx context.Context, obj *model.NtfPreferenceM, fields ...string) error

	NtfPreferenceExpansion
}

type NtfPreferenceExpansion interface {
	GetByUserID(ctx context.Context, userID string) (*model.NtfPreferenceM, error)
	Upsert(ctx context.Context, userID string, prefs model.NotificationPreferences) error
}

type ntfPreferenceStore struct {
	*genericstore.Store[model.NtfPreferenceM]
}

var _ NtfPreferenceStore = (*ntfPreferenceStore)(nil)

func NewNtfPreferenceStore(ds *datastore) *ntfPreferenceStore {
	return &ntfPreferenceStore{
		Store: genericstore.NewStore[model.NtfPreferenceM](ds, NewLogger()),
	}
}

func (s *ntfPreferenceStore) GetByUserID(ctx context.Context, userID string) (*model.NtfPreferenceM, error) {
	return s.Get(ctx, where.F("user_id", userID))
}

func (s *ntfPreferenceStore) Upsert(ctx context.Context, userID string, prefs model.NotificationPreferences) error {
	pref := &model.NtfPreferenceM{UserID: userID}
	if err := pref.SetPreferences(prefs); err != nil {
		return err
	}

	return s.DB(ctx).
		Where("user_id = ?", userID).
		Assign(model.NtfPreferenceM{Preferences: pref.Preferences}).
		FirstOrCreate(pref).Error
}
