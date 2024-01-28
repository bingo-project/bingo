package system

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/model"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

type RoleBiz interface {
	List(ctx context.Context, req *v1.ListRoleRequest) (*v1.ListRoleResponse, error)
	Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error)
	Get(ctx context.Context, roleName string) (*v1.RoleInfo, error)
	Update(ctx context.Context, roleName string, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error)
	Delete(ctx context.Context, roleName string) error

	SetApis(ctx context.Context, a *auth.Authz, roleName string, apiIDs []uint) error
	GetApiIDs(ctx context.Context, a *auth.Authz, roleName string) (v1.GetApiIDsResponse, error)
	SetMenus(ctx context.Context, roleName string, menuIDs []uint) error
	GetMenuIDs(ctx context.Context, roleName string) (v1.GetMenuIDsResponse, error)
	GetMenuTree(ctx context.Context, roleName string) ([]*v1.MenuInfo, error)

	All(ctx context.Context) ([]*v1.RoleInfo, error)
}

type roleBiz struct {
	ds store.IStore
}

var _ RoleBiz = (*roleBiz)(nil)

func NewRole(ds store.IStore) *roleBiz {
	return &roleBiz{ds: ds}
}

func (b *roleBiz) List(ctx context.Context, req *v1.ListRoleRequest) (*v1.ListRoleResponse, error) {
	count, list, err := b.ds.Roles().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list roles", "err", err)

		return nil, err
	}

	data := make([]v1.RoleInfo, 0)
	for _, item := range list {
		var role v1.RoleInfo
		_ = copier.Copy(&role, item)

		data = append(data, role)
	}

	return &v1.ListRoleResponse{Total: count, Data: data}, nil
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
	if req.Remark != nil {
		roleM.Remark = *req.Remark
	}

	if err := b.ds.Roles().Update(ctx, roleM); err != nil {
		return nil, err
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, req)

	return &resp, nil
}

func (b *roleBiz) Delete(ctx context.Context, roleName string) error {
	if roleName == global.RoleRoot {
		return errno.ErrForbidden
	}

	return b.ds.Roles().Delete(ctx, roleName)
}

func (b *roleBiz) SetApis(ctx context.Context, a *auth.Authz, roleName string, apiIDs []uint) error {
	if roleName == global.RoleRoot {
		return errno.ErrForbidden
	}

	// 1. Get apis by ids
	apis, err := b.ds.Apis().GetByIDs(ctx, apiIDs)
	if err != nil {
		return err
	}

	// Get role
	role, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return err
	}

	// Remove policy
	_, err = a.RemoveFilteredPolicy(0, global.RolePrefix+role.Name)
	if err != nil {
		return err
	}

	// Empty
	if len(apis) == 0 {
		return nil
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

func (b *roleBiz) GetApiIDs(ctx context.Context, a *auth.Authz, roleName string) (v1.GetApiIDsResponse, error) {
	// Get role
	role, err := b.ds.Roles().Get(ctx, roleName)
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

func (b *roleBiz) SetMenus(ctx context.Context, roleName string, menuIDs []uint) error {
	if roleName == global.RoleRoot {
		return errno.ErrForbidden
	}

	roleM, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return errno.ErrResourceNotFound
	}

	// Update menus
	roleM.Menus, _ = b.ds.Menus().GetByIDs(ctx, menuIDs)

	err = b.ds.Roles().Update(ctx, roleM)
	if err != nil {
		return err
	}

	return nil
}

func (b *roleBiz) GetMenuIDs(ctx context.Context, roleName string) (v1.GetMenuIDsResponse, error) {
	return b.ds.RoleMenus().GetMenuIDsByRoleName(ctx, roleName)
}

func (b *roleBiz) GetMenuTree(ctx context.Context, roleName string) (ret []*v1.MenuInfo, err error) {
	roleM, err := b.ds.Roles().Get(ctx, roleName)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if roleName == global.RoleRoot {
		roleM.Menus, _ = b.ds.Menus().All(ctx)
	} else {
		// Auto-fill parent menu
		menuIDs, _ := b.ds.RoleMenus().GetMenuIDsByRoleNameWithParent(ctx, roleName)
		roleM.Menus, _ = b.ds.Menus().GetByIDs(ctx, menuIDs)
	}

	// Get menus
	tree, _ := b.ds.Menus().Tree(ctx, roleM.Menus)
	if err != nil {
		return nil, err
	}

	data := make([]*v1.MenuInfo, 0, len(tree))
	for _, item := range tree {
		var menu v1.MenuInfo
		_ = copier.Copy(&menu, item)

		data = append(data, &menu)
	}

	return data, nil
}

func (b *roleBiz) All(ctx context.Context) ([]*v1.RoleInfo, error) {
	list, err := b.ds.Roles().All(ctx)
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

	return data, err
}
