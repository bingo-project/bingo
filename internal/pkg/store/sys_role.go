// ABOUTME: System role store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for role management.

package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	"bingo/internal/pkg/known"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// SysRoleStore defines the interface for system role operations.
type SysRoleStore interface {
	Create(ctx context.Context, obj *model.RoleM) error
	Update(ctx context.Context, obj *model.RoleM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.RoleM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.RoleM, error)

	SysRoleExpansion
}

// SysRoleExpansion defines additional methods for system role operations.
type SysRoleExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListRoleRequest) (int64, []*model.RoleM, error)
	GetByName(ctx context.Context, roleName string) (*model.RoleM, error)
	DeleteByName(ctx context.Context, roleName string) error
	GetByNames(ctx context.Context, names []string) ([]model.RoleM, error)
	GetWithMenus(ctx context.Context, roleName string) (*model.RoleM, error)
	All(ctx context.Context) ([]*model.RoleM, error)
	UpdateWithMenus(ctx context.Context, role *model.RoleM, fields ...string) error
}

type sysRoleStore struct {
	*genericstore.Store[model.RoleM]
}

var _ SysRoleStore = (*sysRoleStore)(nil)

func NewSysRoleStore(store *datastore) *sysRoleStore {
	return &sysRoleStore{
		Store: genericstore.NewStore[model.RoleM](store, NewLogger()),
	}
}

// ListWithRequest lists roles based on request parameters.
func (s *sysRoleStore) ListWithRequest(ctx context.Context, req *v1.ListRoleRequest) (int64, []*model.RoleM, error) {
	opts := where.NewWhere()

	if req.Name != "" {
		opts = opts.F("name", req.Name)
	}

	db := s.DB(ctx, opts).Where("name != ?", known.RoleRoot)
	var ret []*model.RoleM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByName retrieves a role by name.
func (s *sysRoleStore) GetByName(ctx context.Context, roleName string) (*model.RoleM, error) {
	return s.Get(ctx, where.F("name", roleName))
}

// DeleteByName deletes a role by name.
func (s *sysRoleStore) DeleteByName(ctx context.Context, roleName string) error {
	return s.Delete(ctx, where.F("name", roleName))
}

// GetByNames retrieves roles by names.
func (s *sysRoleStore) GetByNames(ctx context.Context, names []string) ([]model.RoleM, error) {
	var ret []model.RoleM
	err := s.DB(ctx).Where("name IN ?", names).Find(&ret).Error

	return ret, err
}

// GetWithMenus retrieves a role with its menus preloaded.
func (s *sysRoleStore) GetWithMenus(ctx context.Context, roleName string) (*model.RoleM, error) {
	var ret model.RoleM
	err := s.DB(ctx).
		Preload("Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Where("name = ?", roleName).
		First(&ret).
		Error

	return &ret, err
}

// All retrieves all roles.
func (s *sysRoleStore) All(ctx context.Context) ([]*model.RoleM, error) {
	var ret []*model.RoleM
	err := s.DB(ctx).Find(&ret).Error

	return ret, err
}

// UpdateWithMenus updates a role and replaces its menus association.
func (s *sysRoleStore) UpdateWithMenus(ctx context.Context, role *model.RoleM, fields ...string) error {
	err := s.DB(ctx).Model(role).Association("Menus").Replace(role.Menus)
	if err != nil {
		return err
	}

	return s.Update(ctx, role, fields...)
}
