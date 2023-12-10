package system

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

type RoleBiz interface {
	List(ctx context.Context, req *v1.ListRoleRequest) (*v1.ListResponse, error)
	Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error)
	Get(ctx context.Context, roleName string) (*v1.RoleInfo, error)
	Update(ctx context.Context, roleName string, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error)
	Delete(ctx context.Context, roleName string) error

	SetApis(ctx context.Context, a *auth.Authz, name string, apiIDs []uint) error
	GetApiIDs(ctx context.Context, a *auth.Authz, name string) (v1.GetApiIDsResponse, error)
	GetMenuIDs(ctx context.Context, roleName string) (v1.GetMenuIDsResponse, error)
}

type roleBiz struct {
	ds store.IStore
}

var _ RoleBiz = (*roleBiz)(nil)

func NewRole(ds store.IStore) *roleBiz {
	return &roleBiz{ds: ds}
}

func (b *roleBiz) List(ctx context.Context, req *v1.ListRoleRequest) (*v1.ListResponse, error) {
	count, list, err := b.ds.Roles().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list roles", "err", err)

		return nil, err
	}

	data := make([]*v1.RoleInfo, 0, len(list))
	for _, item := range list {
		var role v1.RoleInfo
		_ = copier.Copy(&role, item)

		data = append(data, &role)
	}

	return &v1.ListResponse{Total: count, Data: data}, nil
}

func (b *roleBiz) Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error) {
	var roleM model.RoleM
	_ = copier.Copy(&roleM, req)

	err := b.ds.Roles().Create(ctx, &roleM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, roleM)

	return &resp, nil
}

func (b *roleBiz) Get(ctx context.Context, roleName string) (*v1.RoleInfo, error) {
	role, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, role)

	return &resp, nil
}

func (b *roleBiz) Update(ctx context.Context, roleName string, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error) {
	roleM, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Description != nil {
		roleM.Description = *req.Description
	}

	if err := b.ds.Roles().Update(ctx, roleM); err != nil {
		return nil, err
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, req)

	return &resp, nil
}

func (b *roleBiz) Delete(ctx context.Context, roleName string) error {
	return b.ds.Roles().Delete(ctx, roleName)
}

func (b *roleBiz) SetApis(ctx context.Context, a *auth.Authz, name string, apiIDs []uint) error {
	// 1. Get apis by ids
	apis, err := b.ds.Apis().GetByIDs(ctx, apiIDs)
	if err != nil {
		return err
	}

	// Get role
	role, err := b.ds.Roles().Get(ctx, name)
	if err != nil {
		return err
	}

	// Remove policy
	_, err = a.RemoveFilteredPolicy(0, global.RolePrefix+role.Name)
	if err != nil {
		return err
	}

	// Add casbin rule
	var rules [][]string
	for _, api := range apis {
		rules = append(rules, []string{global.RolePrefix + role.Name, api.Path, api.Method})
	}

	_, err = a.AddPolicies(rules)
	if err != nil {
		return err
	}

	return nil
}

func (b *roleBiz) GetApiIDs(ctx context.Context, a *auth.Authz, name string) (v1.GetApiIDsResponse, error) {
	// Get role
	role, err := b.ds.Roles().Get(ctx, name)
	if err != nil {
		return nil, err
	}

	list := a.GetFilteredPolicy(0, global.RolePrefix+role.Name)

	var pathAndMethod [][]string
	for _, v := range list {
		pathAndMethod = append(pathAndMethod, []string{v[1], v[2]})
	}

	resp, err := b.ds.Apis().GetIDsByPathAndMethod(ctx, pathAndMethod)

	return resp, nil
}

func (b *roleBiz) GetMenuIDs(ctx context.Context, roleName string) (v1.GetMenuIDsResponse, error) {
	return b.ds.RoleMenus().GetMenuIDsByRoleName(ctx, roleName)
}
