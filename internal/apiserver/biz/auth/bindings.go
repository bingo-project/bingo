// ABOUTME: Social account bindings business logic.
// ABOUTME: Handles listing and unbinding social accounts.

package auth

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type BindingsBiz interface {
	ListBindings(ctx context.Context, uid string) ([]v1.BindingInfo, error)
	Unbind(ctx context.Context, uid string, provider string) error
}

type bindingsBiz struct {
	ds store.IStore
}

func NewBindingsBiz(ds store.IStore) BindingsBiz {
	return &bindingsBiz{ds: ds}
}

func (b *bindingsBiz) ListBindings(ctx context.Context, uid string) ([]v1.BindingInfo, error) {
	accounts, err := b.ds.UserAccount().FindByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	data := make([]v1.BindingInfo, 0, len(accounts))
	for _, acc := range accounts {
		data = append(data, v1.BindingInfo{
			Provider:  acc.Provider,
			AccountID: acc.AccountID,
			Username:  acc.Username,
			Avatar:    acc.Avatar,
			BindTime:  acc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return data, nil
}

func (b *bindingsBiz) Unbind(ctx context.Context, uid string, provider string) error {
	// 检查是否绑定了该 provider
	account, err := b.ds.UserAccount().FindByUIDAndProvider(ctx, uid, provider)
	if err != nil {
		return errno.ErrNotBound
	}

	// 检查是否为唯一登录方式
	user, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	hasPassword := user.Email != "" || user.Phone != ""
	accountCount, _ := b.ds.UserAccount().CountByUID(ctx, uid)

	if !hasPassword && accountCount <= 1 {
		return errno.ErrCannotUnbindLastLogin
	}

	// 删除绑定
	return b.ds.UserAccount().DeleteByID(ctx, account.ID)
}
