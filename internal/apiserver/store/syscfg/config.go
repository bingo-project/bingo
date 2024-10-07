package syscfg

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"github.com/duke-git/lancet/v2/convertor"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1/syscfg"
	model "bingo/internal/pkg/model/syscfg"
)

type ConfigStore interface {
	List(ctx context.Context, req *v1.ListConfigRequest) (int64, []*model.Config, error)
	Create(ctx context.Context, config *model.Config) error
	Get(ctx context.Context, ID uint) (*model.Config, error)
	Update(ctx context.Context, config *model.Config, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, configs []*model.Config) error
	CreateIfNotExist(ctx context.Context, config *model.Config) error
	FirstOrCreate(ctx context.Context, where any, config *model.Config) error
	UpdateOrCreate(ctx context.Context, where any, config *model.Config) error
	Upsert(ctx context.Context, config *model.Config, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	GetObject(ctx context.Context, key model.CfgKey, resp any) error
	GetServerConfig(ctx context.Context) (*model.ServerConfig, error)
	UpdateServerConfig(ctx context.Context, data *model.ServerConfig) error
}

type configs struct {
	db *gorm.DB
}

var _ ConfigStore = (*configs)(nil)

func NewConfigs(db *gorm.DB) *configs {
	return &configs{db: db}
}

func SearchConfig(req *v1.ListConfigRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Name != nil {
			db.Where("name = ?", req.Name)
		}
		if req.Description != nil {
			db.Where("description = ?", req.Description)
		}
		if req.Key != nil {
			db.Where("key = ?", req.Key)
		}

		return db
	}
}

func (s *configs) List(ctx context.Context, req *v1.ListConfigRequest) (count int64, ret []*model.Config, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchConfig(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *configs) Create(ctx context.Context, config *model.Config) error {
	return s.db.WithContext(ctx).Create(&config).Error
}

func (s *configs) Get(ctx context.Context, ID uint) (config *model.Config, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&config).Error

	return
}

func (s *configs) Update(ctx context.Context, config *model.Config, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&config).Error
}

func (s *configs) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.Config{}).Error
}

func (s *configs) CreateInBatch(ctx context.Context, configs []*model.Config) error {
	return s.db.WithContext(ctx).CreateInBatches(&configs, global.CreateBatchSize).Error
}

func (s *configs) CreateIfNotExist(ctx context.Context, config *model.Config) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&config).
		Error
}

func (s *configs) FirstOrCreate(ctx context.Context, where any, config *model.Config) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&config).
		FirstOrCreate(&config).
		Error
}

func (s *configs) UpdateOrCreate(ctx context.Context, where any, config *model.Config) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Config
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		config.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&config).Error
	})
}

func (s *configs) Upsert(ctx context.Context, config *model.Config, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&config).
		Error
}

func (s *configs) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Config{}).
		Error
}

func (s *configs) GetObject(ctx context.Context, key model.CfgKey, resp any) error {
	where := &model.Config{Key: key}
	cfg := &model.Config{Key: key, Value: convertor.ToString(resp)}

	err := s.FirstOrCreate(ctx, where, cfg)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(cfg.Value), &resp)
	if err != nil {
		return err
	}

	return nil
}

func (s *configs) GetServerConfig(ctx context.Context) (*model.ServerConfig, error) {
	data := model.ServerConfig{
		Status: model.ServerStatusOK,
	}

	err := s.GetObject(ctx, model.CfgKeyServer, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *configs) UpdateServerConfig(ctx context.Context, data *model.ServerConfig) error {
	return s.db.WithContext(ctx).
		Model(&model.Config{}).
		Where(&model.Config{Key: model.CfgKeyServer}).
		Update("value", convertor.ToString(data)).
		Error
}
