package system

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
)

type PermissionBiz interface {
	List(ctx context.Context, req *v1.ListPermissionRequest) (*v1.ListResponse, error)
	Create(ctx context.Context, req *v1.CreatePermissionRequest) (*v1.PermissionInfo, error)
	Get(ctx context.Context, ID uint) (*v1.PermissionInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdatePermissionRequest) (*v1.PermissionInfo, error)
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*v1.PermissionInfo, error)
}

type permissionBiz struct {
	ds store.IStore
}

var _ PermissionBiz = (*permissionBiz)(nil)

func NewPermission(ds store.IStore) *permissionBiz {
	return &permissionBiz{ds: ds}
}

func (b *permissionBiz) List(ctx context.Context, req *v1.ListPermissionRequest) (*v1.ListResponse, error) {
	count, list, err := b.ds.Permissions().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list permissions", "err", err)

		return nil, err
	}

	data := make([]*v1.PermissionInfo, 0, len(list))
	for _, item := range list {
		var permission v1.PermissionInfo
		_ = copier.Copy(&permission, item)

		data = append(data, &permission)
	}

	return &v1.ListResponse{Total: count, Data: data}, nil
}

func (b *permissionBiz) Create(ctx context.Context, req *v1.CreatePermissionRequest) (*v1.PermissionInfo, error) {
	var permissionM model.PermissionM
	_ = copier.Copy(&permissionM, req)

	err := b.ds.Permissions().Create(ctx, &permissionM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.PermissionInfo
	_ = copier.Copy(&resp, permissionM)

	return &resp, nil
}

func (b *permissionBiz) Get(ctx context.Context, ID uint) (*v1.PermissionInfo, error) {
	permission, err := b.ds.Permissions().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.PermissionInfo
	_ = copier.Copy(&resp, permission)

	return &resp, nil
}

func (b *permissionBiz) Update(ctx context.Context, ID uint, req *v1.UpdatePermissionRequest) (*v1.PermissionInfo, error) {
	permissionM, err := b.ds.Permissions().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Method != nil {
		permissionM.Method = *req.Method
	}

	if req.Path != nil {
		permissionM.Path = *req.Path
	}

	if req.Group != nil {
		permissionM.Group = *req.Group
	}

	if req.Description != nil {
		permissionM.Description = *req.Description
	}

	if err := b.ds.Permissions().Update(ctx, permissionM); err != nil {
		return nil, err
	}

	var resp v1.PermissionInfo
	_ = copier.Copy(&resp, req)

	return &resp, nil
}

func (b *permissionBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Permissions().Delete(ctx, ID)
}

func (b *permissionBiz) All(ctx context.Context) ([]*v1.PermissionInfo, error) {
	list, err := b.ds.Permissions().All(ctx)
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

	return permissions, nil
}
