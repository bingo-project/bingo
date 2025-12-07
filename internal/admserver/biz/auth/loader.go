// ABOUTME: AdminLoader implementation for admserver authentication.
// ABOUTME: Loads admin information from database into context.

package auth

import (
	"context"

	"github.com/jinzhu/copier"

	"bingo/internal/pkg/known"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
	"bingo/pkg/contextx"
	"bingo/pkg/errorsx"
)

// AdminLoader loads admin information for admserver.
type AdminLoader struct {
	store store.IStore
}

// NewAdminLoader creates a new AdminLoader.
func NewAdminLoader(store store.IStore) *AdminLoader {
	return &AdminLoader{store: store}
}

// LoadUser loads admin information into context.
func (l *AdminLoader) LoadUser(ctx context.Context, userID string) (context.Context, error) {
	admin, err := l.store.Admin().GetUserInfo(ctx, userID)
	if err != nil || admin.ID == 0 {
		return ctx, errorsx.New(401, "Unauthenticated", "admin not found")
	}

	var adminInfo v1.AdminInfo
	_ = copier.Copy(&adminInfo, admin)

	ctx = contextx.WithUserInfo(ctx, &adminInfo)
	ctx = contextx.WithUsername(ctx, adminInfo.Username)

	return ctx, nil
}

// AdminSubjectResolver resolves authorization subject from admin info.
type AdminSubjectResolver struct{}

// ResolveSubject returns the role-based subject for authorization.
func (r *AdminSubjectResolver) ResolveSubject(ctx context.Context) (string, error) {
	admin, ok := contextx.UserInfo[*v1.AdminInfo](ctx)
	if !ok {
		return "", errorsx.New(401, "Unauthenticated", "admin info not found")
	}

	return known.RolePrefix + admin.RoleName, nil
}
