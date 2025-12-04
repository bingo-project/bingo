// ABOUTME: System API store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for API management.

package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// SysApiStore defines the interface for system API operations.
type SysApiStore interface {
	Create(ctx context.Context, obj *model.ApiM) error
	Update(ctx context.Context, obj *model.ApiM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.ApiM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.ApiM, error)
	FirstOrCreate(ctx context.Context, where any, obj *model.ApiM) error

	SysApiExpansion
}

// SysApiExpansion defines additional methods for system API operations.
type SysApiExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListApiRequest) (int64, []*model.ApiM, error)
	GetByID(ctx context.Context, id uint) (*model.ApiM, error)
	DeleteByID(ctx context.Context, id uint) error
	All(ctx context.Context) ([]*model.ApiM, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.ApiM, error)
	GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) ([]uint, error)
}

type sysApiStore struct {
	*genericstore.Store[model.ApiM]
}

var _ SysApiStore = (*sysApiStore)(nil)

func NewSysApiStore(store *datastore) *sysApiStore {
	return &sysApiStore{
		Store: genericstore.NewStore[model.ApiM](store, NewLogger()),
	}
}

// ListWithRequest lists APIs based on request parameters.
func (s *sysApiStore) ListWithRequest(ctx context.Context, req *v1.ListApiRequest) (int64, []*model.ApiM, error) {
	opts := where.NewWhere()

	if req.Method != "" {
		opts = opts.F("method", req.Method)
	}
	if req.Path != "" {
		opts = opts.Q("path like ?", "%"+req.Path+"%")
	}
	if req.Group != "" {
		opts = opts.F("`group`", req.Group)
	}

	db := s.DB(ctx, opts)
	var ret []*model.ApiM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves an API by ID.
func (s *sysApiStore) GetByID(ctx context.Context, id uint) (*model.ApiM, error) {
	return s.Get(ctx, where.F("id", id))
}

// DeleteByID deletes an API by ID.
func (s *sysApiStore) DeleteByID(ctx context.Context, id uint) error {
	return s.Delete(ctx, where.F("id", id))
}

// All retrieves all APIs.
func (s *sysApiStore) All(ctx context.Context) ([]*model.ApiM, error) {
	var ret []*model.ApiM
	err := s.DB(ctx).Find(&ret).Error

	return ret, err
}

// GetByIDs retrieves APIs by IDs.
func (s *sysApiStore) GetByIDs(ctx context.Context, ids []uint) ([]*model.ApiM, error) {
	var ret []*model.ApiM
	err := s.DB(ctx).Where("id IN ?", ids).Find(&ret).Error

	return ret, err
}

// GetIDsByPathAndMethod retrieves API IDs by path and method pairs.
func (s *sysApiStore) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) ([]uint, error) {
	var ret []uint
	err := s.DB(ctx).
		Model(&model.ApiM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return ret, err
}
