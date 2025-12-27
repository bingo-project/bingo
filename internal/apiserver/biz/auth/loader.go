// ABOUTME: UserLoader implementation for apiserver authentication.
// ABOUTME: Loads user information from database into context.

package auth

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
)

// UserLoader loads user information for apiserver.
type UserLoader struct {
	store store.IStore
}

// NewUserLoader creates a new UserLoader.
func NewUserLoader(store store.IStore) *UserLoader {
	return &UserLoader{store: store}
}

// LoadUser loads user information into context.
func (l *UserLoader) LoadUser(ctx context.Context, userID string) (context.Context, error) {
	user, err := l.store.User().GetByUID(ctx, userID)
	if err != nil || user.ID == 0 {
		return ctx, errorsx.New(401, "Unauthenticated", "user not found")
	}

	var userInfo v1.UserInfo
	_ = copier.Copy(&userInfo, user)
	userInfo.PayPassword = user.PayPassword != ""
	userInfo.Avatar = facade.Config.App.AssetURL(user.Avatar)

	ctx = contextx.WithUserInfo(ctx, &userInfo)
	ctx = contextx.WithUsername(ctx, userInfo.Username)

	return ctx, nil
}
