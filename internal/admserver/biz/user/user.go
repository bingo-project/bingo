package user

import (
	"context"
	"errors"
	"regexp"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// UserBiz 定义了 user 模块在 biz 层所实现的方法.
type UserBiz interface {
	List(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserResponse, error)
	Create(ctx context.Context, req *v1.CreateUserRequest) error
	Get(ctx context.Context, username string) (*v1.UserInfo, error)
	Update(ctx context.Context, username string, req *v1.UpdateUserRequest) error
	Delete(ctx context.Context, username string) error
}

type userBiz struct {
	ds store.IStore
}

var _ UserBiz = (*userBiz)(nil)

func New(ds store.IStore) *userBiz {
	return &userBiz{ds: ds}
}

func (b *userBiz) List(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserResponse, error) {
	count, list, err := b.ds.User().ListWithRequest(ctx, req)
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

	err = b.ds.User().Create(ctx, &userM)
	if err == nil {
		// User exists
		if match, _ := regexp.MatchString("Duplicate entry '.*'", err.Error()); match {
			return errno.ErrUserAlreadyExist
		}

		return
	}

	return
}

func (b *userBiz) Get(ctx context.Context, username string) (*v1.UserInfo, error) {
	user, err := b.ds.User().GetByUsername(ctx, username)
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
	userM, err := b.ds.User().GetByUsername(ctx, username)
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

	if err := b.ds.User().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}

func (b *userBiz) Delete(ctx context.Context, username string) error {
	return b.ds.User().DeleteByUsername(ctx, username)
}
