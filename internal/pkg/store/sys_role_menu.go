// ABOUTME: System role-menu store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for role-menu associations.

package store

import (
	"context"

	linq "github.com/ahmetb/go-linq/v3"
	"github.com/bingo-project/component-base/util/gormutil"

	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// SysRoleMenuStore defines the interface for system role-menu operations.
type SysRoleMenuStore interface {
	Create(ctx context.Context, obj *model.RoleMenuM) error
	Update(ctx context.Context, obj *model.RoleMenuM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.RoleMenuM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.RoleMenuM, error)

	SysRoleMenuExpansion
}

// SysRoleMenuExpansion defines additional methods for system role-menu operations.
type SysRoleMenuExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListRoleMenuRequest) (int64, []*model.RoleMenuM, error)
	GetByID(ctx context.Context, id uint) (*model.RoleMenuM, error)
	DeleteByID(ctx context.Context, id uint) error
	GetMenuIDsByRoleName(ctx context.Context, roleName string) ([]uint, error)
	GetMenuIDsByRoleNameWithParent(ctx context.Context, roleName string) ([]uint, error)
}

type sysRoleMenuStore struct {
	*genericstore.Store[model.RoleMenuM]
}

var _ SysRoleMenuStore = (*sysRoleMenuStore)(nil)

func NewSysRoleMenuStore(store *datastore) *sysRoleMenuStore {
	return &sysRoleMenuStore{
		Store: genericstore.NewStore[model.RoleMenuM](store, NewLogger()),
	}
}

// ListWithRequest lists role-menus based on request parameters.
func (s *sysRoleMenuStore) ListWithRequest(ctx context.Context, req *v1.ListRoleMenuRequest) (int64, []*model.RoleMenuM, error) {
	db := s.DB(ctx)
	var ret []*model.RoleMenuM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves a role-menu by ID.
func (s *sysRoleMenuStore) GetByID(ctx context.Context, id uint) (*model.RoleMenuM, error) {
	return s.Get(ctx, where.F("id", id))
}

// DeleteByID deletes a role-menu by ID.
func (s *sysRoleMenuStore) DeleteByID(ctx context.Context, id uint) error {
	return s.Delete(ctx, where.F("id", id))
}

// GetMenuIDsByRoleName retrieves menu IDs by role name.
func (s *sysRoleMenuStore) GetMenuIDsByRoleName(ctx context.Context, roleName string) ([]uint, error) {
	var ret []uint
	err := s.DB(ctx).
		Model(&model.RoleMenuM{}).
		Select("menu_id").
		Where(&model.RoleMenuM{RoleName: roleName}).
		Find(&ret).
		Error

	return ret, err
}

// GetMenuIDsByRoleNameWithParent retrieves menu IDs including parent IDs by role name.
func (s *sysRoleMenuStore) GetMenuIDsByRoleNameWithParent(ctx context.Context, roleName string) ([]uint, error) {
	var menuIDs []uint
	err := s.DB(ctx).
		Model(&model.RoleMenuM{}).
		Select("menu_id").
		Where(&model.RoleMenuM{RoleName: roleName}).
		Find(&menuIDs).
		Error
	if err != nil {
		return nil, err
	}

	var parentIDs []uint
	err = s.DB(ctx).
		Model(&model.MenuM{}).
		Where("id IN (?)", menuIDs).
		Where("hidden = ?", false).
		Select("parent_id").
		Find(&parentIDs).
		Error
	if err != nil {
		return nil, err
	}

	// Union menuIDs & parentIDs
	var ret []uint
	linq.From(menuIDs).
		Union(linq.From(parentIDs)).
		ToSlice(&ret)

	return ret, nil
}
