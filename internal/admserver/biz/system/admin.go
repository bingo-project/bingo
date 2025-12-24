package system

import (
	"context"
	"regexp"
	"slices"

	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type AdminBiz interface {
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)
	ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error

	List(ctx context.Context, req *v1.ListAdminRequest) (*v1.ListAdminResponse, error)
	Create(ctx context.Context, req *v1.CreateAdminRequest) (*v1.AdminInfo, error)
	Get(ctx context.Context, username string) (*v1.AdminInfo, error)
	Update(ctx context.Context, username string, req *v1.UpdateAdminRequest) (*v1.AdminInfo, error)
	Delete(ctx context.Context, username string) error

	SetRoles(ctx context.Context, username string, req *v1.SetRolesRequest) (*v1.AdminInfo, error)
	SwitchRole(ctx context.Context, username string, admin *v1.SwitchRoleRequest) (*v1.AdminInfo, error)
}

type adminBiz struct {
	ds store.IStore
}

var _ AdminBiz = (*adminBiz)(nil)

func NewAdmin(ds store.IStore) *adminBiz {
	return &adminBiz{ds: ds}
}

func (b *adminBiz) List(ctx context.Context, req *v1.ListAdminRequest) (*v1.ListAdminResponse, error) {
	count, list, err := b.ds.Admin().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list admins", "err", err)

		return nil, err
	}

	data := make([]v1.AdminInfo, 0)
	for _, item := range list {
		var admin v1.AdminInfo
		_ = copier.Copy(&admin, item)

		data = append(data, admin)
	}

	return &v1.ListAdminResponse{Total: count, Data: data}, nil
}

func (b *adminBiz) Create(ctx context.Context, req *v1.CreateAdminRequest) (*v1.AdminInfo, error) {
	// Block creating root user
	if req.Username == known.UserRoot {
		return nil, errno.ErrInvalidArgument.WithMessage("该用户名不可用")
	}

	var adminM model.AdminM
	_ = copier.Copy(&adminM, req)

	// Create roles & current role
	if len(req.RoleNames) > 0 {
		adminM.Roles, _ = b.ds.SysRole().GetByNames(ctx, req.RoleNames)
		if len(adminM.Roles) > 0 {
			adminM.RoleName = adminM.Roles[0].Name
		}
	}

	err := b.ds.Admin().Create(ctx, &adminM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.AdminInfo
	_ = copier.Copy(&resp, adminM)

	return &resp, nil
}

func (b *adminBiz) Get(ctx context.Context, username string) (*v1.AdminInfo, error) {
	admin, err := b.ds.Admin().GetUserInfo(ctx, username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.AdminInfo
	_ = copier.Copy(&resp, admin)

	// Add virtual root role for root user
	if username == known.UserRoot {
		rootRole := v1.RoleInfo{
			Name:        known.UserRoot,
			Description: "Root",
			Status:      "enabled",
		}
		resp.Roles = append([]v1.RoleInfo{rootRole}, resp.Roles...)
	}

	return &resp, nil
}

func (b *adminBiz) Update(ctx context.Context, username string, req *v1.UpdateAdminRequest) (*v1.AdminInfo, error) {
	adminM, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	if req.Nickname != nil {
		adminM.Nickname = *req.Nickname
	}
	if req.Email != nil {
		adminM.Email = req.Email
	}
	if req.Phone != nil {
		adminM.Phone = req.Phone
	}
	if req.Avatar != nil {
		adminM.Avatar = *req.Avatar
	}
	if req.Status != "" {
		adminM.Status = model.AdminStatus(req.Status)
	}
	if req.Password != nil {
		adminM.Password, _ = auth.Encrypt(*req.Password)
	}

	// Update roles
	updateRoles := len(req.RoleNames) > 0
	if updateRoles {
		adminM.Roles, _ = b.ds.SysRole().GetByNames(ctx, req.RoleNames)
		if len(adminM.Roles) == 0 {
			return nil, errno.ErrInvalidArgument
		}
		if !slices.ContainsFunc(adminM.Roles, func(r model.RoleM) bool { return r.Name == adminM.RoleName }) {
			adminM.RoleName = adminM.Roles[0].Name
		}
	}

	if updateRoles {
		err = b.ds.Admin().UpdateWithRoles(ctx, adminM)
	} else {
		err = b.ds.Admin().Update(ctx, adminM)
	}
	if err != nil {
		return nil, err
	}

	var resp v1.AdminInfo
	_ = copier.Copy(&resp, adminM)

	return &resp, nil
}

func (b *adminBiz) Delete(ctx context.Context, username string) error {
	if username == known.UserRoot {
		return errno.ErrPermissionDenied
	}

	return b.ds.Admin().DeleteByUsername(ctx, username)
}

func (b *adminBiz) SetRoles(ctx context.Context, username string, req *v1.SetRolesRequest) (*v1.AdminInfo, error) {
	adminM, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Update roles & current role
	adminM.Roles, _ = b.ds.SysRole().GetByNames(ctx, req.RoleNames)
	if len(adminM.Roles) == 0 {
		return nil, errno.ErrInvalidArgument
	}
	if !slices.ContainsFunc(adminM.Roles, func(r model.RoleM) bool { return r.Name == adminM.RoleName }) {
		adminM.RoleName = adminM.Roles[0].Name
	}

	err = b.ds.Admin().UpdateWithRoles(ctx, adminM)
	if err != nil {
		return nil, err
	}

	var resp v1.AdminInfo
	_ = copier.Copy(&resp, adminM)

	return &resp, nil
}

func (b *adminBiz) SwitchRole(ctx context.Context, username string, req *v1.SwitchRoleRequest) (*v1.AdminInfo, error) {
	adminM, err := b.ds.Admin().GetByUsername(ctx, username)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Root user can switch back to root
	if username == known.UserRoot && req.RoleName == known.UserRoot {
		adminM.RoleName = known.UserRoot
		err = b.ds.Admin().Update(ctx, adminM, "role_name")
		if err != nil {
			return nil, err
		}

		var resp v1.AdminInfo
		_ = copier.Copy(&resp, adminM)
		return &resp, nil
	}

	// Check if the user has the role
	hasRole := b.ds.Admin().HasRole(ctx, adminM, req.RoleName)
	if !hasRole {
		return nil, errno.ErrNotFound
	}

	// Update roles & current role
	adminM.RoleName = req.RoleName
	err = b.ds.Admin().Update(ctx, adminM, "role_name")
	if err != nil {
		return nil, err
	}

	var resp v1.AdminInfo
	_ = copier.Copy(&resp, adminM)

	return &resp, err
}
