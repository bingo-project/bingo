package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	model "github.com/bingo-project/bingo/internal/pkg/model/syscfg"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1/syscfg"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AppVersionStore defines the interface for app version operations.
type AppVersionStore interface {
	Create(ctx context.Context, obj *model.AppVersion) error
	Update(ctx context.Context, obj *model.AppVersion, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AppVersion, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AppVersion, error)

	AppVersionExpansion
}

// AppVersionExpansion defines additional methods for app version operations.
type AppVersionExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListAppVersionRequest) (int64, []*model.AppVersion, error)
	GetByID(ctx context.Context, id uint) (*model.AppVersion, error)
	DeleteByID(ctx context.Context, id uint) error
}

type appVersionStore struct {
	*genericstore.Store[model.AppVersion]
}

var _ AppVersionStore = (*appVersionStore)(nil)

func NewAppVersionStore(store *datastore) *appVersionStore {
	return &appVersionStore{
		Store: genericstore.NewStore[model.AppVersion](store, NewLogger()),
	}
}

// GetByID retrieves an app version by ID.
func (s *appVersionStore) GetByID(ctx context.Context, id uint) (*model.AppVersion, error) {
	var ret model.AppVersion
	err := s.DB(ctx).Where("id = ?", id).First(&ret).Error

	return &ret, err
}

// DeleteByID deletes an app version by ID.
func (s *appVersionStore) DeleteByID(ctx context.Context, id uint) error {
	return s.DB(ctx).Where("id = ?", id).Delete(&model.AppVersion{}).Error
}

// ListWithRequest lists app versions based on request parameters.
func (s *appVersionStore) ListWithRequest(ctx context.Context, req *v1.ListAppVersionRequest) (int64, []*model.AppVersion, error) {
	opts := where.NewWhere()

	if req.Name != nil {
		opts = opts.F("name", *req.Name)
	}
	if req.Version != nil {
		opts = opts.F("version", *req.Version)
	}
	if req.Description != nil {
		opts = opts.F("description", *req.Description)
	}
	if req.AboutUs != nil {
		opts = opts.F("about_us", *req.AboutUs)
	}
	if req.Logo != nil {
		opts = opts.F("logo", *req.Logo)
	}
	if req.Enabled != nil {
		opts = opts.F("enabled", *req.Enabled)
	}

	db := s.DB(ctx, opts)
	var ret []*model.AppVersion
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}
