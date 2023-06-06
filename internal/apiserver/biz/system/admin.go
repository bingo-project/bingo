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

type AdminBiz interface {
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)

	List(ctx context.Context, offset, limit int) (*v1.ListAdminResponse, error)
	Create(ctx context.Context, r *v1.CreateAdminRequest) (*v1.GetAdminResponse, error)
	Get(ctx context.Context, username string) (*v1.GetAdminResponse, error)
	Update(ctx context.Context, username string, r *v1.UpdateAdminRequest) (*v1.GetAdminResponse, error)
	Delete(ctx context.Context, username string) error
}

type adminBiz struct {
	ds store.IStore
}

// 确保 adminBiz 实现了 AdminBiz 接口.
var _ AdminBiz = (*adminBiz)(nil)

func NewAdmin(ds store.IStore) *adminBiz {
	return &adminBiz{ds: ds}
}

func (b *adminBiz) List(ctx context.Context, offset, limit int) (*v1.ListAdminResponse, error) {
	count, list, err := b.ds.Admins().List(ctx, offset, limit)
	if err != nil {
		log.C(ctx).Errorw("Failed to list admins from storage", "err", err)

		return nil, err
	}

	admins := make([]*v1.AdminInfo, 0, len(list))
	for _, item := range list {
		var admin v1.AdminInfo
		_ = copier.Copy(&admin, item)

		admins = append(admins, &admin)
	}

	log.C(ctx).Debugw("Get admins from backend storage", "count", len(admins))

	return &v1.ListAdminResponse{TotalCount: count, Data: admins}, nil
}

func (b *adminBiz) Create(ctx context.Context, request *v1.CreateAdminRequest) (*v1.GetAdminResponse, error) {
	var adminM system.AdminM
	_ = copier.Copy(&adminM, request)

	err := b.ds.Admins().Create(ctx, &adminM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrAdminAlreadyExist
		}

		return nil, err
	}

	var resp v1.GetAdminResponse
	_ = copier.Copy(&resp, adminM)

	return &resp, nil
}

func (b *adminBiz) Get(ctx context.Context, username string) (*v1.GetAdminResponse, error) {
	admin, err := b.ds.Admins().Get(ctx, username)
	if err != nil {
		return nil, errno.ErrAdminNotFound
	}

	var resp v1.GetAdminResponse
	_ = copier.Copy(&resp, admin)

	return &resp, nil
}

func (b *adminBiz) Update(ctx context.Context, username string, request *v1.UpdateAdminRequest) (*v1.GetAdminResponse, error) {
	adminM, err := b.ds.Admins().Get(ctx, username)
	if err != nil {
		return nil, errno.ErrAdminNotFound
	}

	if request.Nickname != nil {
		adminM.Nickname = *request.Nickname
	}

	if request.Email != nil {
		adminM.Email = *request.Email
	}

	if request.Phone != nil {
		adminM.Phone = *request.Phone
	}

	if request.Avatar != nil {
		adminM.Avatar = *request.Avatar
	}

	// Update roles & current role
	if len(request.RoleNames) > 0 {
		adminM.RoleName = request.RoleNames[0]
	}

	if err := b.ds.Admins().Update(ctx, adminM); err != nil {
		return nil, err
	}

	var resp v1.GetAdminResponse
	_ = copier.Copy(&resp, request)

	return &resp, nil
}

func (b *adminBiz) Delete(ctx context.Context, username string) error {
	return b.ds.Admins().Delete(ctx, username)
}
