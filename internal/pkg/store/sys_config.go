package store

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/component-base/util/gormutil"
	"github.com/duke-git/lancet/v2/convertor"

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

// ConfigExpansion defines additional methods for config operations.
type ConfigExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListConfigRequest) (int64, []*model.Config, error)
	GetByID(ctx context.Context, id uint) (*model.Config, error)
	DeleteByID(ctx context.Context, id uint) error
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

// GetByID retrieves a config by ID.
func (s *configStore) GetByID(ctx context.Context, id uint) (*model.Config, error) {
	var ret model.Config
	err := s.DB(ctx).Where("id = ?", id).First(&ret).Error

	return &ret, err
}

// DeleteByID deletes a config by ID.
func (s *configStore) DeleteByID(ctx context.Context, id uint) error {
	return s.DB(ctx).Where("id = ?", id).Delete(&model.Config{}).Error
}

// ListWithRequest lists configs based on request parameters.
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
