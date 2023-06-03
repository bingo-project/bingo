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
)

type RoleBiz interface {
	List(ctx context.Context, offset, limit int) (*v1.ListRoleResponse, error)
	Create(ctx context.Context, r *v1.CreateRoleRequest) (*v1.GetRoleResponse, error)
	Get(ctx context.Context, ID uint) (*v1.GetRoleResponse, error)
	Update(ctx context.Context, ID uint, r *v1.UpdateRoleRequest) (*v1.GetRoleResponse, error)
	Delete(ctx context.Context, ID uint) error
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

func (b *roleBiz) Get(ctx context.Context, ID uint) (*v1.GetRoleResponse, error) {
	role, err := b.ds.Roles().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrRoleNotFound
	}

	var resp v1.GetRoleResponse
	_ = copier.Copy(&resp, role)

	return &resp, nil
}

func (b *roleBiz) Update(ctx context.Context, ID uint, request *v1.UpdateRoleRequest) (*v1.GetRoleResponse, error) {
	roleM, err := b.ds.Roles().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrRoleNotFound
	}

	if request.Name != nil {
		roleM.Name = *request.Name
	}

	if err := b.ds.Roles().Update(ctx, roleM); err != nil {
		return nil, err
	}

	var resp v1.GetRoleResponse
	_ = copier.Copy(&resp, request)

	return &resp, nil
}

func (b *roleBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Roles().Delete(ctx, ID)
}
