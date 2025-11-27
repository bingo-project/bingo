package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
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
	CreateInBatch(ctx context.Context, channels []*model.Channel) error
	CreateIfNotExist(ctx context.Context, channel *model.Channel) error
	FirstOrCreate(ctx context.Context, where any, channel *model.Channel) error
	UpdateOrCreate(ctx context.Context, where any, channel *model.Channel) error
	Upsert(ctx context.Context, channel *model.Channel, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error
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

// CreateInBatch 批量创建.
func (s *botChannelStore) CreateInBatch(ctx context.Context, channels []*model.Channel) error {
	return s.DB(ctx).CreateInBatches(channels, global.CreateBatchSize).Error
}

// CreateIfNotExist 如果不存在则创建.
func (s *botChannelStore) CreateIfNotExist(ctx context.Context, channel *model.Channel) error {
	return s.DB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(channel).
		Error
}

// FirstOrCreate 首先查找，不存在则创建.
func (s *botChannelStore) FirstOrCreate(ctx context.Context, where any, channel *model.Channel) error {
	return s.DB(ctx).
		Where(where).
		Attrs(channel).
		FirstOrCreate(channel).
		Error
}

// UpdateOrCreate 更新或创建.
func (s *botChannelStore) UpdateOrCreate(ctx context.Context, where any, channel *model.Channel) error {
	return s.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Channel
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		channel.ID = exist.ID
		return tx.Omit("CreatedAt").Save(channel).Error
	})
}

// Upsert 创建或更新.
func (s *botChannelStore) Upsert(ctx context.Context, channel *model.Channel, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.DB(ctx).Clauses(do).Create(channel).Error
}

// DeleteInBatch 批量删除.
func (s *botChannelStore) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.DB(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Channel{}).
		Error
}

// DeleteChannel 根据 channel_id 删除.
func (s *botChannelStore) DeleteChannel(ctx context.Context, channelID string) error {
	return s.DB(ctx).Where("channel_id = ?", channelID).Delete(&model.Channel{}).Error
}
