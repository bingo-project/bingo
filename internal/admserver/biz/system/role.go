package system

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type RoleBiz interface {
	List(ctx context.Context, req *v1.ListRoleRequest) (*v1.ListRoleResponse, error)
	Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error)
	Get(ctx context.Context, roleName string) (*v1.RoleInfo, error)
	Update(ctx context.Context, roleName string, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error)
	Delete(ctx context.Context, roleName string) error

	SetApis(ctx context.Context, a *auth.Authorizer, roleName string, apiIDs []uint) error
	GetApiIDs(ctx context.Context, a *auth.Authorizer, roleName string) (v1.GetApiIDsResponse, error)
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
	count, list, err := b.ds.SysRole().ListWithRequest(ctx, req)
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

	err := b.ds.SysRole().Create(ctx, &roleM)
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
	role, err := b.ds.SysRole().GetByName(ctx, roleName)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, role)

	return &resp, nil
}

func (b *roleBiz) Update(ctx context.Context, roleName string, req *v1.UpdateRoleRequest) (*v1.RoleInfo, error) {
	roleM, err := b.ds.SysRole().GetByName(ctx, roleName)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	if req.Description != nil {
		roleM.Description = *req.Description
	}
	if req.Status != nil {
		roleM.Status = *req.Status
	}
	if req.Remark != nil {
		roleM.Remark = *req.Remark
	}

	if err := b.ds.SysRole().Update(ctx, roleM); err != nil {
		return nil, err
	}

	var resp v1.RoleInfo
	_ = copier.Copy(&resp, roleM)

	return &resp, nil
}

func (b *roleBiz) Delete(ctx context.Context, roleName string) error {
	if roleName == known.RoleRoot {
		return errno.ErrPermissionDenied
	}

	return b.ds.SysRole().DeleteByName(ctx, roleName)
}

func (b *roleBiz) SetApis(ctx context.Context, a *auth.Authorizer, roleName string, apiIDs []uint) error {
	if roleName == known.RoleRoot {
		return errno.ErrPermissionDenied
	}

	// 1. Get apis by ids
	apis, err := b.ds.SysApi().GetByIDs(ctx, apiIDs)
	if err != nil {
		return err
	}

	// Get role
	role, err := b.ds.SysRole().GetByName(ctx, roleName)
	if err != nil {
		return err
	}

	// Remove policy
	_, err = a.Enforcer().RemoveFilteredPolicy(0, known.RolePrefix+role.Name)
	if err != nil {
		return err
	}

	// Empty
	if len(apis) == 0 {
		return nil
	}

	// Add casbin rule
	rules := make([][]string, 0)
	for _, api := range apis {
		rules = append(rules, []string{known.RolePrefix + role.Name, api.Path, api.Method})
	}

	_, err = a.Enforcer().AddPolicies(rules)
	if err != nil {
		return err
	}

	return nil
}

func (b *roleBiz) GetApiIDs(ctx context.Context, a *auth.Authorizer, roleName string) (ret v1.GetApiIDsResponse, err error) {
	// Get role
	role, err := b.ds.SysRole().GetByName(ctx, roleName)
	if err != nil {
		return
	}

	list, _ := a.Enforcer().GetFilteredPolicy(0, known.RolePrefix+role.Name)

	pathAndMethod := make([][]string, 0)
	for _, v := range list {
		pathAndMethod = append(pathAndMethod, []string{v[1], v[2]})
	}

	resp, err := b.ds.SysApi().GetIDsByPathAndMethod(ctx, pathAndMethod)
	if err != nil {
		return
	}

	return resp, nil
}

func (b *roleBiz) SetMenus(ctx context.Context, roleName string, menuIDs []uint) error {
	if roleName == known.RoleRoot {
		return errno.ErrPermissionDenied
	}

	roleM, err := b.ds.SysRole().GetByName(ctx, roleName)
	if err != nil {
		return errno.ErrNotFound
	}

	// Update menus
	roleM.Menus, _ = b.ds.SysMenu().GetByIDs(ctx, menuIDs)

	err = b.ds.SysRole().UpdateWithMenus(ctx, roleM)
	if err != nil {
		return err
	}

	return nil
}

func (b *roleBiz) GetMenuIDs(ctx context.Context, roleName string) (v1.GetMenuIDsResponse, error) {
	return b.ds.SysRoleMenu().GetMenuIDsByRoleName(ctx, roleName)
}

func (b *roleBiz) GetMenuTree(ctx context.Context, roleName string) (ret []*v1.MenuInfo, err error) {
	var menus []*model.MenuM
	if roleName == known.RoleRoot {
		menus, _ = b.ds.SysMenu().AllEnabled(ctx)
	} else {
		// Auto-fill parent menu
		menuIDs, _ := b.ds.SysRoleMenu().GetMenuIDsByRoleNameWithParent(ctx, roleName)
		menus, _ = b.ds.SysMenu().GetByIDs(ctx, menuIDs)
	}

	// Get menus
	tree, err := b.ds.SysMenu().Tree(ctx, menus)
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
	list, err := b.ds.SysRole().All(ctx)
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
