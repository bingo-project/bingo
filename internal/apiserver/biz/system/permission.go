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

type PermissionBiz interface {
	List(ctx context.Context, offset, limit int) (*v1.ListPermissionResponse, error)
	Create(ctx context.Context, r *v1.CreatePermissionRequest) (*v1.GetPermissionResponse, error)
	Get(ctx context.Context, ID uint) (*v1.GetPermissionResponse, error)
	Update(ctx context.Context, ID uint, r *v1.UpdatePermissionRequest) (*v1.GetPermissionResponse, error)
	Delete(ctx context.Context, ID uint) error
}

type permissionBiz struct {
	ds store.IStore
}

// 确保 permissionBiz 实现了 PermissionBiz 接口.
var _ PermissionBiz = (*permissionBiz)(nil)

func NewPermission(ds store.IStore) *permissionBiz {
	return &permissionBiz{ds: ds}
}

func (b *permissionBiz) List(ctx context.Context, offset, limit int) (*v1.ListPermissionResponse, error) {
	count, list, err := b.ds.Permissions().List(ctx, offset, limit)
	if err != nil {
		log.C(ctx).Errorw("Failed to list permissions from storage", "err", err)

		return nil, err
	}

	permissions := make([]*v1.PermissionInfo, 0, len(list))
	for _, item := range list {
		var permission v1.PermissionInfo
		_ = copier.Copy(&permission, item)

		permissions = append(permissions, &permission)
	}

	log.C(ctx).Debugw("Get permissions from backend storage", "count", len(permissions))

	return &v1.ListPermissionResponse{TotalCount: count, Data: permissions}, nil
}

func (b *permissionBiz) Create(ctx context.Context, request *v1.CreatePermissionRequest) (*v1.GetPermissionResponse, error) {
	var permissionM system.PermissionM
	_ = copier.Copy(&permissionM, request)

	err := b.ds.Permissions().Create(ctx, &permissionM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrPermissionAlreadyExist
		}

		return nil, err
	}

	var resp v1.GetPermissionResponse
	_ = copier.Copy(&resp, permissionM)

	return &resp, nil
}

func (b *permissionBiz) Get(ctx context.Context, ID uint) (*v1.GetPermissionResponse, error) {
	permission, err := b.ds.Permissions().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrPermissionNotFound
	}

	var resp v1.GetPermissionResponse
	_ = copier.Copy(&resp, permission)

	return &resp, nil
}

func (b *permissionBiz) Update(ctx context.Context, ID uint, request *v1.UpdatePermissionRequest) (*v1.GetPermissionResponse, error) {
	permissionM, err := b.ds.Permissions().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrPermissionNotFound
	}

	if request.Method != nil {
		permissionM.Method = *request.Method
	}

	if request.Path != nil {
		permissionM.Path = *request.Path
	}

	if request.Group != nil {
		permissionM.Group = *request.Group
	}

	if request.Description != nil {
		permissionM.Description = *request.Description
	}

	if err := b.ds.Permissions().Update(ctx, permissionM); err != nil {
		return nil, err
	}

	var resp v1.GetPermissionResponse
	_ = copier.Copy(&resp, request)

	return &resp, nil
}

func (b *permissionBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Permissions().Delete(ctx, ID)
}
