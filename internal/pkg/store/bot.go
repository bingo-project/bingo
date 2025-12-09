package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	model "github.com/bingo-project/bingo/internal/pkg/model/bot"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1/bot"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// BotStore 定义了 Bot 相关操作的接口.
type BotStore interface {
	Create(ctx context.Context, obj *model.Bot) error
	Update(ctx context.Context, obj *model.Bot, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.Bot, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.Bot, error)

	BotExpansion
}

// BotExpansion 定义了 Bot 操作的扩展方法.
type BotExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListBotRequest) (int64, []*model.Bot, error)
	GetByID(ctx context.Context, id uint) (*model.Bot, error)
	DeleteByID(ctx context.Context, id uint) error
}

type botStore struct {
	*genericstore.Store[model.Bot]
}

var _ BotStore = (*botStore)(nil)

func NewBotStore(store *datastore) *botStore {
	return &botStore{
		Store: genericstore.NewStore[model.Bot](store, NewLogger()),
	}
}

// ListWithRequest 根据请求参数列表查询.
func (s *botStore) ListWithRequest(ctx context.Context, req *v1.ListBotRequest) (int64, []*model.Bot, error) {
	// 构建查询条件
	opts := where.NewWhere()

	if req.Name != nil {
		opts = opts.F("name", *req.Name)
	}
	if req.Source != nil {
		opts = opts.F("source", *req.Source)
	}
	if req.Description != nil {
		opts = opts.F("description", *req.Description)
	}
	if req.Token != nil {
		opts = opts.F("token", *req.Token)
	}
	if req.Enabled != nil {
		opts = opts.F("enabled", *req.Enabled)
	}

	// 处理分页
	db := s.DB(ctx, opts)
	var ret []*model.Bot
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves a bot by ID.
func (s *botStore) GetByID(ctx context.Context, id uint) (*model.Bot, error) {
	return s.Get(ctx, where.F("id", id))
}

// DeleteByID deletes a bot by ID.
func (s *botStore) DeleteByID(ctx context.Context, id uint) error {
	return s.Delete(ctx, where.F("id", id))
}
