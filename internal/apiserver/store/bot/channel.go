package bot

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1/bot"
	model "bingo/internal/apiserver/model/bot"
)

type ChannelStore interface {
	List(ctx context.Context, req *v1.ListChannelRequest) (int64, []*model.Channel, error)
	Create(ctx context.Context, channel *model.Channel) error
	Get(ctx context.Context, ID uint) (*model.Channel, error)
	Update(ctx context.Context, channel *model.Channel, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, channels []*model.Channel) error
	CreateIfNotExist(ctx context.Context, channel *model.Channel) error
	FirstOrCreate(ctx context.Context, where any, channel *model.Channel) error
	UpdateOrCreate(ctx context.Context, where any, channel *model.Channel) error
	Upsert(ctx context.Context, channel *model.Channel, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	DeleteChannel(ctx context.Context, channelID string) error
}

type channels struct {
	db *gorm.DB
}

var _ ChannelStore = (*channels)(nil)

func NewChannels(db *gorm.DB) *channels {
	return &channels{db: db}
}

func SearchChannel(req *v1.ListChannelRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Source != nil {
			db.Where("source = ?", req.Source)
		}
		if req.ChannelID != nil {
			db.Where("channel_id = ?", req.ChannelID)
		}
		if req.Author != nil {
			db.Where("author = ?", req.Author)
		}

		return db
	}
}

func (s *channels) List(ctx context.Context, req *v1.ListChannelRequest) (count int64, ret []*model.Channel, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchChannel(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *channels) Create(ctx context.Context, channel *model.Channel) error {
	return s.db.WithContext(ctx).Create(&channel).Error
}

func (s *channels) Get(ctx context.Context, ID uint) (channel *model.Channel, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&channel).Error

	return
}

func (s *channels) Update(ctx context.Context, channel *model.Channel, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&channel).Error
}

func (s *channels) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.Channel{}).Error
}

func (s *channels) CreateInBatch(ctx context.Context, channels []*model.Channel) error {
	return s.db.WithContext(ctx).CreateInBatches(&channels, global.CreateBatchSize).Error
}

func (s *channels) CreateIfNotExist(ctx context.Context, channel *model.Channel) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&channel).
		Error
}

func (s *channels) FirstOrCreate(ctx context.Context, where any, channel *model.Channel) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&channel).
		FirstOrCreate(&channel).
		Error
}

func (s *channels) UpdateOrCreate(ctx context.Context, where any, channel *model.Channel) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Channel
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		channel.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&channel).Error
	})
}

func (s *channels) Upsert(ctx context.Context, channel *model.Channel, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&channel).
		Error
}

func (s *channels) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Channel{}).
		Error
}

func (s *channels) DeleteChannel(ctx context.Context, channelID string) error {
	return s.db.WithContext(ctx).Where("channel_id = ?", channelID).Delete(&model.Channel{}).Error
}
