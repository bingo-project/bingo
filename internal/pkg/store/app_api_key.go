package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// ApiKeyStore defines the interface for API key operations.
type ApiKeyStore interface {
	Create(ctx context.Context, obj *model.ApiKey) error
	Update(ctx context.Context, obj *model.ApiKey, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.ApiKey, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.ApiKey, error)

	ApiKeyExpansion
}

// ApiKeyExpansion defines additional methods for API key operations.
type ApiKeyExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListApiKeyRequest) (int64, []*model.ApiKey, error)
	GetByID(ctx context.Context, id uint) (*model.ApiKey, error)
	DeleteByID(ctx context.Context, id uint) error
	GetByAK(ctx context.Context, ak string) (*model.ApiKey, error)
}

type apiKeyStore struct {
	*genericstore.Store[model.ApiKey]
}

var _ ApiKeyStore = (*apiKeyStore)(nil)

func NewApiKeyStore(store *datastore) *apiKeyStore {
	return &apiKeyStore{
		Store: genericstore.NewStore[model.ApiKey](store, NewLogger()),
	}
}

// ListWithRequest lists API keys based on request parameters.
func (s *apiKeyStore) ListWithRequest(ctx context.Context, req *v1.ListApiKeyRequest) (int64, []*model.ApiKey, error) {
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
	if req.AccessKey != nil {
		opts = opts.F("access_key", *req.AccessKey)
	}
	if req.Status != nil {
		opts = opts.F("status", *req.Status)
	}

	db := s.DB(ctx, opts)
	var ret []*model.ApiKey
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves an API key by ID.
func (s *apiKeyStore) GetByID(ctx context.Context, id uint) (*model.ApiKey, error) {
	var ret model.ApiKey
	err := s.DB(ctx).Where("id = ?", id).First(&ret).Error

	return &ret, err
}

// DeleteByID deletes an API key by ID.
func (s *apiKeyStore) DeleteByID(ctx context.Context, id uint) error {
	return s.DB(ctx).Where("id = ?", id).Delete(&model.ApiKey{}).Error
}

// GetByAK retrieves an API key by its access key.
func (s *apiKeyStore) GetByAK(ctx context.Context, ak string) (*model.ApiKey, error) {
	var ret model.ApiKey
	err := s.DB(ctx).Where("access_key = ?", ak).First(&ret).Error

	return &ret, err
}
