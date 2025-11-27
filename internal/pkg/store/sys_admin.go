package store

import (
	"context"

	"bingo/internal/pkg/model"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

type AdminStore interface {
	Create(ctx context.Context, obj *model.AdminM) error
	Update(ctx context.Context, obj *model.AdminM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AdminM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AdminM, error)

	AdminExpansion
}

// AdminExpansion 定义了用户操作的附加方法.
// nolint: iface
type AdminExpansion interface {
}

type adminStore struct {
	*genericstore.Store[model.AdminM]
}

var _ AdminStore = (*adminStore)(nil)

func NewAdminStore(store *datastore) *adminStore {
	return &adminStore{
		Store: genericstore.NewStore[model.AdminM](store, NewLogger()),
	}
}
