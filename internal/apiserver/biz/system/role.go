package system

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model/system"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

type RoleBiz interface {
	List(ctx context.Context, offset, limit int) (*v1.ListRoleResponse, error)
	Create(ctx context.Context, r *v1.CreateRoleRequest) (*v1.GetRoleResponse, error)
	Get(ctx context.Context, roleName string) (*v1.GetRoleResponse, error)
	Update(ctx context.Context, roleName string, r *v1.UpdateRoleRequest) (*v1.GetRoleResponse, error)
	Delete(ctx context.Context, roleName string) error

	SetPermissions(ctx context.Context, a *auth.Authz, name string, permissionIDs []uint) error
	GetPermissionIDs(ctx context.Context, a *auth.Authz, name string) (v1.GetPermissionIDsResponse, error)
}

type roleBiz struct {
	ds store.IStore
}

// 确保 roleBiz 实现了 RoleBiz 接口.
var _ RoleBiz = (*roleBiz)(nil)

func NewRole(ds store.IStore) *roleBiz {
	return &roleBiz{ds: ds}
}

func (b *roleBiz) List(ctx context.Context, offset, limit int) (*v1.ListRoleResponse, error) {
	count, list, err := b.ds.Roles().List(ctx, offset, limit)
	if err != nil {
		log.C(ctx).Errorw("Failed to list roles from storage", "err", err)

		return nil, err
	}

	roles := make([]*v1.RoleInfo, 0, len(list))
	for _, item := range list {
		var role v1.RoleInfo
		_ = copier.Copy(&role, item)

		roles = append(roles, &role)
	}

	log.C(ctx).Debugw("Get roles from backend storage", "count", len(roles))

	return &v1.ListRoleResponse{TotalCount: count, Data: roles}, nil
}

func (b *roleBiz) Create(ctx context.Context, request *v1.CreateRoleRequest) (*v1.GetRoleResponse, error) {
	var roleM system.RoleM
	_ = copier.Copy(&roleM, request)

	err := b.ds.Roles().Create(ctx, &roleM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrRoleAlreadyExist
		}

		return nil, err
	}

	var resp v1.GetRoleResponse
	_ = copier.Copy(&resp, roleM)

	return &resp, nil
}

func (b *roleBiz) Get(ctx context.Context, roleName string) (*v1.GetRoleResponse, error) {
	role, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return nil, errno.ErrRoleNotFound
	}

	var resp v1.GetRoleResponse
	_ = copier.Copy(&resp, role)

	return &resp, nil
}

func (b *roleBiz) Update(ctx context.Context, roleName string, request *v1.UpdateRoleRequest) (*v1.GetRoleResponse, error) {
	roleM, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return nil, errno.ErrRoleNotFound
	}

	if request.Description != nil {
		roleM.Description = *request.Description
	}

	if err := b.ds.Roles().Update(ctx, roleM); err != nil {
		return nil, err
	}

	var resp v1.GetRoleResponse
	_ = copier.Copy(&resp, request)

	return &resp, nil
}

func (b *roleBiz) Delete(ctx context.Context, roleName string) error {
	return b.ds.Roles().Delete(ctx, roleName)
}

func (b *roleBiz) SetPermissions(ctx context.Context, a *auth.Authz, name string, permissionIDs []uint) error {
	// 1. Get permissions by ids
	permissions, err := b.ds.Permissions().GetByIDs(ctx, permissionIDs)
	if err != nil {
		return err
	}

	// Get role
	role, err := b.ds.Roles().Get(ctx, name)
	if err != nil {
		return err
	}

	// Remove policy
	_, err = a.RemoveFilteredPolicy(0, system.RolePrefix+role.Name)
	if err != nil {
		return err
	}

	// Add casbin rule
	var rules [][]string
	for _, permission := range permissions {
		rules = append(rules, []string{system.RolePrefix + role.Name, permission.Path, permission.Method})
	}

	_, err = a.AddPolicies(rules)
	if err != nil {
		return err
	}

	return nil
}

func (b *roleBiz) GetPermissionIDs(ctx context.Context, a *auth.Authz, name string) (v1.GetPermissionIDsResponse, error) {
	// Get role
	role, err := b.ds.Roles().Get(ctx, name)
	if err != nil {
		return nil, err
	}

	list := a.GetFilteredPolicy(0, system.RolePrefix+role.Name)

	var pathAndMethod [][]string
	for _, v := range list {
		pathAndMethod = append(pathAndMethod, []string{v[1], v[2]})
	}

	resp, err := b.ds.Permissions().GetIDsByPathAndMethod(ctx, pathAndMethod)

	return resp, nil
}
