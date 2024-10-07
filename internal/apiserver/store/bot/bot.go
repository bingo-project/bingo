package bot

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	v1 "bingo/internal/apiserver/http/request/v1/bot"
	"bingo/internal/pkg/global"
	model "bingo/internal/pkg/model/bot"
)

type BotStore interface {
	List(ctx context.Context, req *v1.ListBotRequest) (int64, []*model.Bot, error)
	Create(ctx context.Context, bot *model.Bot) error
	Get(ctx context.Context, ID uint) (*model.Bot, error)
	Update(ctx context.Context, bot *model.Bot, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, bots []*model.Bot) error
	CreateIfNotExist(ctx context.Context, bot *model.Bot) error
	FirstOrCreate(ctx context.Context, where any, bot *model.Bot) error
	UpdateOrCreate(ctx context.Context, where any, bot *model.Bot) error
	Upsert(ctx context.Context, bot *model.Bot, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error
}

type bots struct {
	db *gorm.DB
}

var _ BotStore = (*bots)(nil)

func NewBots(db *gorm.DB) *bots {
	return &bots{db: db}
}

func SearchBot(req *v1.ListBotRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Name != nil {
			db.Where("name = ?", req.Name)
		}
		if req.Source != nil {
			db.Where("source = ?", req.Source)
		}
		if req.Description != nil {
			db.Where("description = ?", req.Description)
		}
		if req.Token != nil {
			db.Where("token = ?", req.Token)
		}
		if req.Enabled != nil {
			db.Where("enabled = ?", req.Enabled)
		}

		return db
	}
}

func (s *bots) List(ctx context.Context, req *v1.ListBotRequest) (count int64, ret []*model.Bot, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchBot(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *bots) Create(ctx context.Context, bot *model.Bot) error {
	return s.db.WithContext(ctx).Create(&bot).Error
}

func (s *bots) Get(ctx context.Context, ID uint) (bot *model.Bot, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&bot).Error

	return
}

func (s *bots) Update(ctx context.Context, bot *model.Bot, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&bot).Error
}

func (s *bots) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.Bot{}).Error
}

func (s *bots) CreateInBatch(ctx context.Context, bots []*model.Bot) error {
	return s.db.WithContext(ctx).CreateInBatches(&bots, global.CreateBatchSize).Error
}

func (s *bots) CreateIfNotExist(ctx context.Context, bot *model.Bot) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&bot).
		Error
}

func (s *bots) FirstOrCreate(ctx context.Context, where any, bot *model.Bot) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&bot).
		FirstOrCreate(&bot).
		Error
}

func (s *bots) UpdateOrCreate(ctx context.Context, where any, bot *model.Bot) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var exist model.Bot
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		bot.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&bot).Error
	})
}

func (s *bots) Upsert(ctx context.Context, bot *model.Bot, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.Clauses(do).
		Create(&bot).
		Error
}

func (s *bots) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Bot{}).
		Error
}
