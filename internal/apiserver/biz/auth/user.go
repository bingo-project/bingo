// ABOUTME: User profile update business logic.
// ABOUTME: Handles email/phone binding with verification.

package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type UserBiz interface {
	UpdateProfile(ctx context.Context, uid string, req *v1.UpdateProfileRequest) error
}

type userBiz struct {
	ds      store.IStore
	codeBiz CodeBiz
}

func NewUserBiz(ds store.IStore, codeBiz CodeBiz) UserBiz {
	return &userBiz{ds: ds, codeBiz: codeBiz}
}

func (b *userBiz) UpdateProfile(ctx context.Context, uid string, req *v1.UpdateProfileRequest) error {
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return errno.ErrUserNotFound
	}

	// 更新 email
	if req.Email != nil && *req.Email != user.Email {
		// 检查是否被占用
		if existing, _ := b.ds.User().FindByEmail(ctx, *req.Email); existing != nil && existing.UID != uid {
			return errno.ErrAccountOccupied
		}
		// 验证码检查
		if err := b.codeBiz.Verify(ctx, *req.Email, CodeSceneBind, req.Code); err != nil {
			return err
		}
		user.Email = *req.Email
	}

	// 更新 phone
	if req.Phone != nil && *req.Phone != user.Phone {
		// 检查是否被占用
		if existing, _ := b.ds.User().FindByPhone(ctx, *req.Phone); existing != nil && existing.UID != uid {
			return errno.ErrAccountOccupied
		}
		// 验证码检查
		if err := b.codeBiz.Verify(ctx, *req.Phone, CodeSceneBind, req.Code); err != nil {
			return err
		}
		user.Phone = *req.Phone
	}

	// 更新 nickname
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}

	return b.ds.User().Update(ctx, user)
}
