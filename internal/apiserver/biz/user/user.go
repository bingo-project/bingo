package user

import (
	"context"
	"errors"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/web/token"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/model"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/pkg/auth"
)

// UserBiz 定义了 user 模块在 biz 层所实现的方法.
type UserBiz interface {
	List(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserResponse, error)
	Create(ctx context.Context, req *v1.CreateUserRequest) error
	Get(ctx context.Context, username string) (*v1.UserInfo, error)
	Update(ctx context.Context, username string, req *v1.UpdateUserRequest) error
	Delete(ctx context.Context, username string) error

	ChangePassword(ctx context.Context, username string, r *v1.ChangePasswordRequest) error
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)
}

type userBiz struct {
	ds store.IStore
}

var _ UserBiz = (*userBiz)(nil)

func New(ds store.IStore) *userBiz {
	return &userBiz{ds: ds}
}

func (b *userBiz) List(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.Users().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list users", "err", err)

		return nil, err
	}

	data := make([]v1.UserInfo, 0, len(list))
	for _, item := range list {
		var user v1.UserInfo
		_ = copier.Copy(&user, item)

		data = append(data, user)
	}

	return &v1.ListUserResponse{Total: count, Data: data}, nil
}

func (b *userBiz) Create(ctx context.Context, req *v1.CreateUserRequest) (err error) {
	var userM model.UserM
	_ = copier.Copy(&userM, req)

	err = b.ds.Users().Create(ctx, &userM)
	if err == nil {
		return
	}

	// User exists
	if match, _ := regexp.MatchString("Duplicate entry '.*'", err.Error()); match {
		return errno.ErrUserAlreadyExist
	}

	return
}

func (b *userBiz) Get(ctx context.Context, username string) (*v1.UserInfo, error) {
	user, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrUserNotFound
		}

		return nil, err
	}

	var resp v1.UserInfo
	_ = copier.Copy(&resp, user)

	return &resp, nil
}

func (b *userBiz) Update(ctx context.Context, username string, req *v1.UpdateUserRequest) error {
	userM, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		return err
	}

	if req.Email != nil {
		userM.Email = *req.Email
	}

	if req.Nickname != nil {
		userM.Nickname = *req.Nickname
	}

	if req.Phone != nil {
		userM.Phone = *req.Phone
	}

	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

func (b *userBiz) Delete(ctx context.Context, username string) error {
	return b.ds.Users().Delete(ctx, username)
}

func (b *userBiz) ChangePassword(ctx context.Context, username string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.Users().Get(ctx, username)
	if err != nil {
		return err
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)
	if err := b.ds.Users().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

func (b *userBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.Users().Get(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
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
