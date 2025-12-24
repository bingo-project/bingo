// ABOUTME: AdminLoader implementation for admserver authentication.
// ABOUTME: Loads admin information from database into context.

package auth

import (
	"context"

	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
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

	// Root user gets virtual root role + all real roles
	if userID == known.UserRoot {
		adminInfo.Roles = l.getAllRolesForRoot(ctx)
	}

	ctx = contextx.WithUserInfo(ctx, &adminInfo)
	ctx = contextx.WithUsername(ctx, adminInfo.Username)

	return ctx, nil
}

// getAllRolesForRoot returns virtual root role + all real roles for root user.
func (l *AdminLoader) getAllRolesForRoot(ctx context.Context) []v1.RoleInfo {
	rootRole := v1.RoleInfo{
		Name:        known.UserRoot,
		Description: "Root",
		Status:      string(model.AdminStatusEnabled),
	}

	roles := []v1.RoleInfo{rootRole}

	allRoles, err := l.store.SysRole().All(ctx)
	if err != nil {
		return roles
	}

	for _, r := range allRoles {
		roles = append(roles, v1.RoleInfo{
			Name:        r.Name,
			Description: r.Description,
			Status:      string(r.Status),
		})
	}

	return roles
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
