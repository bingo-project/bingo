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
	CreateInBatch(ctx context.Context, bots []*model.Bot) error
	CreateIfNotExist(ctx context.Context, bot *model.Bot) error
	FirstOrCreate(ctx context.Context, where any, bot *model.Bot) error
	UpdateOrCreate(ctx context.Context, where any, bot *model.Bot) error
	Upsert(ctx context.Context, bot *model.Bot, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error
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

// CreateInBatch 批量创建.
func (s *botStore) CreateInBatch(ctx context.Context, bots []*model.Bot) error {
	return s.DB(ctx).CreateInBatches(bots, global.CreateBatchSize).Error
}

// CreateIfNotExist 如果不存在则创建.
func (s *botStore) CreateIfNotExist(ctx context.Context, bot *model.Bot) error {
	return s.DB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(bot).
		Error
}

// FirstOrCreate 首先查找，不存在则创建.
func (s *botStore) FirstOrCreate(ctx context.Context, where any, bot *model.Bot) error {
	return s.DB(ctx).
		Where(where).
		Attrs(bot).
		FirstOrCreate(bot).
		Error
}

// UpdateOrCreate 更新或创建.
func (s *botStore) UpdateOrCreate(ctx context.Context, where any, bot *model.Bot) error {
	return s.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Bot
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		bot.ID = exist.ID
		return tx.Omit("CreatedAt").Save(bot).Error
	})
}

// Upsert 创建或更新.
func (s *botStore) Upsert(ctx context.Context, bot *model.Bot, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.DB(ctx).Clauses(do).Create(bot).Error
}

// DeleteInBatch 批量删除.
func (s *botStore) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.DB(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Bot{}).
		Error
}
