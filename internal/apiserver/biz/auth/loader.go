// ABOUTME: UserLoader implementation for apiserver authentication.
// ABOUTME: Loads user information from database into context.

package auth

import (
	"context"

	"github.com/jinzhu/copier"

	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/contextx"
	"bingo/pkg/errorsx"
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

	ctx = contextx.WithUserInfo(ctx, &userInfo)
	ctx = contextx.WithUsername(ctx, userInfo.Username)

	return ctx, nil
}
