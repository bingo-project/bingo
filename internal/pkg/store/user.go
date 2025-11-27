package store

import (
	"context"

	"bingo/internal/pkg/model"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

type UserStore interface {
	Create(ctx context.Context, obj *model.UserM) error
	Update(ctx context.Context, obj *model.UserM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error)

	UserExpansion
}

// UserExpansion 定义了用户操作的附加方法.
// nolint: iface
type UserExpansion interface {
}

type userStore struct {
	*genericstore.Store[model.UserM]
}

var _ UserStore = (*userStore)(nil)

func NewUserStore(store *datastore) *userStore {
	return &userStore{
		Store: genericstore.NewStore[model.UserM](store, NewLogger()),
	}
}
