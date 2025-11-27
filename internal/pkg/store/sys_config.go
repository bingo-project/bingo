package store

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/component-base/util/gormutil"
	"github.com/duke-git/lancet/v2/convertor"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
	model "bingo/internal/pkg/model/syscfg"
	v1 "bingo/pkg/api/apiserver/v1/syscfg"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// ConfigStore 定义了 Config 相关操作的接口.
type ConfigStore interface {
	Create(ctx context.Context, obj *model.Config) error
	Update(ctx context.Context, obj *model.Config, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.Config, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.Config, error)

	ConfigExpansion
}

// ConfigExpansion 定义了 Config 操作的扩展方法.
type ConfigExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListConfigRequest) (int64, []*model.Config, error)
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

type configStore struct {
	*genericstore.Store[model.Config]
}

var _ ConfigStore = (*configStore)(nil)

func NewConfigStore(store *datastore) *configStore {
	return &configStore{
		Store: genericstore.NewStore[model.Config](store, NewLogger()),
	}
}

// ListWithRequest 根据请求参数列表查询.
func (s *configStore) ListWithRequest(ctx context.Context, req *v1.ListConfigRequest) (int64, []*model.Config, error) {
	// 构建查询条件
	opts := where.NewWhere()

	if req.Name != nil {
		opts = opts.F("name", *req.Name)
	}
	if req.Description != nil {
		opts = opts.F("description", *req.Description)
	}
	if req.Key != nil {
		opts = opts.F("key", *req.Key)
	}

	// 处理分页
	db := s.DB(ctx, opts)
	var ret []*model.Config
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// CreateInBatch 批量创建.
func (s *configStore) CreateInBatch(ctx context.Context, configs []*model.Config) error {
	return s.DB(ctx).CreateInBatches(configs, global.CreateBatchSize).Error
}

// CreateIfNotExist 如果不存在则创建.
func (s *configStore) CreateIfNotExist(ctx context.Context, config *model.Config) error {
	return s.DB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(config).
		Error
}

// FirstOrCreate 首先查找，不存在则创建.
func (s *configStore) FirstOrCreate(ctx context.Context, where any, config *model.Config) error {
	return s.DB(ctx).
		Where(where).
		Attrs(config).
		FirstOrCreate(config).
		Error
}

// UpdateOrCreate 更新或创建.
func (s *configStore) UpdateOrCreate(ctx context.Context, where any, config *model.Config) error {
	return s.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Config
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		config.ID = exist.ID
		return tx.Omit("CreatedAt").Save(config).Error
	})
}

// Upsert 创建或更新.
func (s *configStore) Upsert(ctx context.Context, config *model.Config, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.DB(ctx).Clauses(do).Create(config).Error
}

// DeleteInBatch 批量删除.
func (s *configStore) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.DB(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Config{}).
		Error
}

// GetObject 获取 JSON 对象.
func (s *configStore) GetObject(ctx context.Context, key model.CfgKey, resp any) error {
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

// GetServerConfig 获取服务器配置.
func (s *configStore) GetServerConfig(ctx context.Context) (*model.ServerConfig, error) {
	data := model.ServerConfig{
		Status: model.ServerStatusOK,
	}

	err := s.GetObject(ctx, model.CfgKeyServer, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// UpdateServerConfig 更新服务器配置.
func (s *configStore) UpdateServerConfig(ctx context.Context, data *model.ServerConfig) error {
	return s.DB(ctx).
		Model(&model.Config{}).
		Where(&model.Config{Key: model.CfgKeyServer}).
		Update("value", convertor.ToString(data)).
		Error
}
