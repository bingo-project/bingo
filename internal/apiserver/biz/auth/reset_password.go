// ABOUTME: Password reset business logic.
// ABOUTME: Handles password reset via verification code.

package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type ResetPasswordBiz interface {
	ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) error
}

type resetPasswordBiz struct {
	ds      store.IStore
	codeBiz CodeBiz
}

func NewResetPasswordBiz(ds store.IStore, codeBiz CodeBiz) ResetPasswordBiz {
	return &resetPasswordBiz{ds: ds, codeBiz: codeBiz}
}

func (b *resetPasswordBiz) ResetPassword(ctx context.Context, req *v1.ResetPasswordRequest) error {
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return err
	}

	// 查找用户
	var user *model.UserM
	switch accountType {
	case AccountTypeEmail:
		user, err = b.ds.User().FindByEmail(ctx, req.Account)
	case AccountTypePhone:
		user, err = b.ds.User().FindByPhone(ctx, req.Account)
	}
	if err != nil {
		return errno.ErrUserNotFound
	}

	// 验证码检查
	if err := b.codeBiz.Verify(ctx, req.Account, CodeSceneResetPassword, req.Code); err != nil {
		return err
	}

	// 更新密码
	user.Password, _ = auth.Encrypt(req.Password)

	return b.ds.User().Update(ctx, user, "password")
}
