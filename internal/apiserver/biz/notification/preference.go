// ABOUTME: Business logic for notification preferences.
// ABOUTME: Handles getting and updating user notification settings.

package notification

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type PreferenceBiz interface {
	Get(ctx context.Context, userID string) (*v1.NotificationPreferences, error)
	Update(ctx context.Context, userID string, req *v1.NotificationPreferences) error
}

type preferenceBiz struct {
	ds store.IStore
}

var _ PreferenceBiz = (*preferenceBiz)(nil)

func NewPreference(ds store.IStore) PreferenceBiz {
	return &preferenceBiz{ds: ds}
}

func (b *preferenceBiz) Get(ctx context.Context, userID string) (*v1.NotificationPreferences, error) {
	pref, err := b.ds.NtfPreference().GetByUserID(ctx, userID)
	if err != nil {
		// Return default preferences if not set
		defaults := model.DefaultPreferences()

		return &v1.NotificationPreferences{
			System:      v1.ChannelPreference{InApp: defaults.System.InApp, Email: defaults.System.Email},
			Security:    v1.ChannelPreference{InApp: defaults.Security.InApp, Email: defaults.Security.Email},
			Transaction: v1.ChannelPreference{InApp: defaults.Transaction.InApp, Email: defaults.Transaction.Email},
			Social:      v1.ChannelPreference{InApp: defaults.Social.InApp, Email: defaults.Social.Email},
		}, nil
	}

	prefs := pref.GetPreferences()

	return &v1.NotificationPreferences{
		System:      v1.ChannelPreference{InApp: prefs.System.InApp, Email: prefs.System.Email},
		Security:    v1.ChannelPreference{InApp: prefs.Security.InApp, Email: prefs.Security.Email},
		Transaction: v1.ChannelPreference{InApp: prefs.Transaction.InApp, Email: prefs.Transaction.Email},
		Social:      v1.ChannelPreference{InApp: prefs.Social.InApp, Email: prefs.Social.Email},
	}, nil
}

func (b *preferenceBiz) Update(ctx context.Context, userID string, req *v1.NotificationPreferences) error {
	prefs := model.NotificationPreferences{
		System:      model.ChannelPreference{InApp: req.System.InApp, Email: req.System.Email},
		Security:    model.ChannelPreference{InApp: req.Security.InApp, Email: req.Security.Email},
		Transaction: model.ChannelPreference{InApp: req.Transaction.InApp, Email: req.Transaction.Email},
		Social:      model.ChannelPreference{InApp: req.Social.InApp, Email: req.Social.Email},
	}

	if err := b.ds.NtfPreference().Upsert(ctx, userID, prefs); err != nil {
		return errno.ErrDBWrite.WithMessage("update preferences: %v", err)
	}

	return nil
}
