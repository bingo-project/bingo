package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	model "bingo/internal/pkg/model/bot"
	v1 "bingo/pkg/api/apiserver/v1/bot"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// BotChannelStore 定义了 Bot Channel 相关操作的接口.
type BotChannelStore interface {
	Create(ctx context.Context, obj *model.Channel) error
	Update(ctx context.Context, obj *model.Channel, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.Channel, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.Channel, error)

	BotChannelExpansion
}

// BotChannelExpansion 定义了 Bot Channel 操作的扩展方法.
type BotChannelExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListChannelRequest) (int64, []*model.Channel, error)
	DeleteChannel(ctx context.Context, channelID string) error
}

type botChannelStore struct {
	*genericstore.Store[model.Channel]
}

var _ BotChannelStore = (*botChannelStore)(nil)

func NewBotChannelStore(store *datastore) *botChannelStore {
	return &botChannelStore{
		Store: genericstore.NewStore[model.Channel](store, NewLogger()),
	}
}

// ListWithRequest 根据请求参数列表查询.
func (s *botChannelStore) ListWithRequest(ctx context.Context, req *v1.ListChannelRequest) (int64, []*model.Channel, error) {
	// 构建查询条件
	opts := where.NewWhere()

	if req.Source != nil {
		opts = opts.F("source", *req.Source)
	}
	if req.ChannelID != nil {
		opts = opts.F("channel_id", *req.ChannelID)
	}
	if req.Author != nil {
		opts = opts.F("author", *req.Author)
	}

	// 处理分页
	db := s.DB(ctx, opts)
	var ret []*model.Channel
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// DeleteChannel 根据 channel_id 删除.
func (s *botChannelStore) DeleteChannel(ctx context.Context, channelID string) error {
	return s.DB(ctx).Where("channel_id = ?", channelID).Delete(&model.Channel{}).Error
}
