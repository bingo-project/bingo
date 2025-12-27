package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AuthProviderStore defines the interface for auth provider operations.
type AuthProviderStore interface {
	Create(ctx context.Context, obj *model.AuthProvider) error
	Update(ctx context.Context, obj *model.AuthProvider, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AuthProvider, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AuthProvider, error)

	AuthProviderExpansion
}

// AuthProviderExpansion defines additional methods for auth provider operations.
type AuthProviderExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListAuthProviderRequest) (int64, []*model.AuthProvider, error)
	GetByID(ctx context.Context, id uint) (*model.AuthProvider, error)
	DeleteByID(ctx context.Context, id uint) error
	FindEnabled(ctx context.Context) ([]*model.AuthProvider, error)
	FirstEnabled(ctx context.Context, provider string) (*model.AuthProvider, error)
	FindByName(ctx context.Context, name string) (*model.AuthProvider, error)
}

type authProviderStore struct {
	*genericstore.Store[model.AuthProvider]
}

var _ AuthProviderStore = (*authProviderStore)(nil)

func NewAuthProviderStore(store *datastore) *authProviderStore {
	return &authProviderStore{
		Store: genericstore.NewStore[model.AuthProvider](store, NewLogger()),
	}
}

// ListWithRequest lists auth providers based on request parameters.
func (s *authProviderStore) ListWithRequest(ctx context.Context, req *v1.ListAuthProviderRequest) (int64, []*model.AuthProvider, error) {
	opts := where.NewWhere()

	if req.Name != nil {
		opts = opts.F("name", *req.Name)
	}
	if req.Status != nil {
		opts = opts.F("status", *req.Status)
	}
	if req.IsDefault != nil {
		opts = opts.F("is_default", *req.IsDefault)
	}

	db := s.DB(ctx, opts)
	var ret []*model.AuthProvider
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves an auth provider by ID.
func (s *authProviderStore) GetByID(ctx context.Context, id uint) (*model.AuthProvider, error) {
	var ret model.AuthProvider
	err := s.DB(ctx).Where("id = ?", id).First(&ret).Error

	return &ret, err
}

// DeleteByID deletes an auth provider by ID.
func (s *authProviderStore) DeleteByID(ctx context.Context, id uint) error {
	return s.DB(ctx).Where("id = ?", id).Delete(&model.AuthProvider{}).Error
}

// FindEnabled returns all enabled auth providers.
func (s *authProviderStore) FindEnabled(ctx context.Context) ([]*model.AuthProvider, error) {
	var ret []*model.AuthProvider
	err := s.DB(ctx).
		Where("status = ?", model.AuthProviderStatusEnabled).
		Find(&ret).
		Error

	return ret, err
}

// FirstEnabled returns the first enabled auth provider by name.
func (s *authProviderStore) FirstEnabled(ctx context.Context, provider string) (*model.AuthProvider, error) {
	var ret model.AuthProvider
	err := s.DB(ctx).
		Where("name = ?", provider).
		Where("status = ?", model.AuthProviderStatusEnabled).
		First(&ret).
		Error

	return &ret, err
}

// FindByName finds an auth provider by name regardless of status.
func (s *authProviderStore) FindByName(ctx context.Context, name string) (*model.AuthProvider, error) {
	var ret model.AuthProvider
	err := s.DB(ctx).Where("name = ?", name).First(&ret).Error
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
