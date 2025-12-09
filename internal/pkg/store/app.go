package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AppStore defines the interface for app operations.
type AppStore interface {
	Create(ctx context.Context, obj *model.App) error
	Update(ctx context.Context, obj *model.App, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.App, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.App, error)

	AppExpansion
}

// AppExpansion defines additional methods for app operations.
type AppExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListAppRequest) (int64, []*model.App, error)
	GetByAppID(ctx context.Context, appID string) (*model.App, error)
	DeleteByAppID(ctx context.Context, appID string) error
}

type appStore struct {
	*genericstore.Store[model.App]
}

var _ AppStore = (*appStore)(nil)

func NewAppStore(store *datastore) *appStore {
	return &appStore{
		Store: genericstore.NewStore[model.App](store, NewLogger()),
	}
}

// ListWithRequest lists apps based on request parameters.
func (s *appStore) ListWithRequest(ctx context.Context, req *v1.ListAppRequest) (int64, []*model.App, error) {
	opts := where.NewWhere()

	if req.UID != nil {
		opts = opts.F("uid", *req.UID)
	}
	if req.AppID != nil {
		opts = opts.F("app_id", *req.AppID)
	}
	if req.Name != nil {
		opts = opts.F("name", *req.Name)
	}
	if req.Status != nil {
		opts = opts.F("status", *req.Status)
	}
	if req.Description != nil {
		opts = opts.F("description", *req.Description)
	}
	if req.Logo != nil {
		opts = opts.F("logo", *req.Logo)
	}

	db := s.DB(ctx, opts)
	var ret []*model.App
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByAppID retrieves an app by its app ID.
func (s *appStore) GetByAppID(ctx context.Context, appID string) (*model.App, error) {
	var ret model.App
	err := s.DB(ctx).Where("app_id = ?", appID).First(&ret).Error

	return &ret, err
}

// DeleteByAppID deletes an app by its app ID.
func (s *appStore) DeleteByAppID(ctx context.Context, appID string) error {
	return s.DB(ctx).Where("app_id = ?", appID).Delete(&model.App{}).Error
}
