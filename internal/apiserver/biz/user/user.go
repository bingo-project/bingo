package user

import (
	"context"
	"errors"
	"regexp"

	"github.com/bingo-project/component-base/web/token"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
	"bingo/pkg/auth"
)

// UserBiz 定义了 user 模块在 biz 层所实现的方法.
type UserBiz interface {
	ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)
	Create(ctx context.Context, r *v1.CreateUserRequest) error
	Get(ctx context.Context, username string) (*v1.GetUserResponse, error)
	List(ctx context.Context, offset, limit int) (*v1.ListUserResponse, error)
	Update(ctx context.Context, username string, r *v1.UpdateUserRequest) error
	Delete(ctx context.Context, username string) error
}

// UserBiz 接口的实现.
type userBiz struct {
	ds store.IStore
}

// 确保 userBiz 实现了 UserBiz 接口.
var _ UserBiz = (*userBiz)(nil)

// New 创建一个实现了 UserBiz 接口的实例.
func New(ds store.IStore) *userBiz {
	return &userBiz{ds: ds}
}

// ChangePassword 是 UserBiz 接口中 `ChangePassword` 方法的实现.
func (b *userBiz) ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		return err
	}

	// Check password
	if err := auth.Compare(userM.Password, r.OldPassword); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(r.NewPassword)
	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

// Login 是 UserBiz 接口中 `Login` 方法的实现.
func (b *userBiz) Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Users().Get(ctx, r.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// Check password
	err = auth.Compare(user.Password, r.Password)
	if err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// Generate token
	t, err := token.Sign(user.Username, nil)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{Token: t.AccessToken}, nil
}

// Create 是 UserBiz 接口中 `Create` 方法的实现.
func (b *userBiz) Create(ctx context.Context, r *v1.CreateUserRequest) (err error) {
	var userM model.UserM
	_ = copier.Copy(&userM, r)

	err = b.ds.Users().Create(ctx, &userM)
	if err == nil {
		return
	}

	// User exists
	if match, _ := regexp.MatchString("Duplicate entry '.*' for key 'username'", err.Error()); match {
		return errno.ErrUserAlreadyExist
	}

	return
}

// Get 是 UserBiz 接口中 `Get` 方法的实现.
func (b *userBiz) Get(ctx context.Context, username string) (*v1.GetUserResponse, error) {
	user, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}

		return nil, err
	}

	var resp v1.GetUserResponse
	_ = copier.Copy(&resp, user)

	return &resp, nil
}

// List 是 UserBiz 接口中 `List` 方法的实现.
func (b *userBiz) List(ctx context.Context, offset, limit int) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, offset, limit)
	if err != nil {
		log.C(ctx).Errorw("Failed to list users from storage", "err", err)

		return nil, err
	}

	users := make([]*v1.UserInfo, 0, len(list))
	for _, item := range list {
		var user v1.UserInfo
		_ = copier.Copy(&user, item)

		users = append(users, &user)
	}

	log.C(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &v1.ListUserResponse{TotalCount: count, Data: users}, nil
}

// Update 是 UserBiz 接口中 `Update` 方法的实现.
func (b *userBiz) Update(ctx context.Context, username string, user *v1.UpdateUserRequest) error {
	userM, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		return err
	}

	if user.Email != nil {
		userM.Email = *user.Email
	}

	if user.Nickname != nil {
		userM.Nickname = *user.Nickname
	}

	if user.Phone != nil {
		userM.Phone = *user.Phone
	}

	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

func (b *userBiz) Delete(ctx context.Context, username string) error {
	return b.ds.Users().Delete(ctx, username)
}
