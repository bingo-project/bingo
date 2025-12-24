// ABOUTME: System menu store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for menu management including tree building.

package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// SysMenuStore defines the interface for system menu operations.
type SysMenuStore interface {
	Create(ctx context.Context, obj *model.MenuM) error
	Update(ctx context.Context, obj *model.MenuM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.MenuM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.MenuM, error)

	SysMenuExpansion
}

// SysMenuExpansion defines additional methods for system menu operations.
type SysMenuExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListMenuRequest) (int64, []*model.MenuM, error)
	GetByID(ctx context.Context, id uint) (*model.MenuM, error)
	GetByIDWithApis(ctx context.Context, id uint) (*model.MenuM, error)
	DeleteByID(ctx context.Context, id uint) error
	All(ctx context.Context) ([]*model.MenuM, error)
	AllEnabled(ctx context.Context) ([]*model.MenuM, error)
	AllWithApis(ctx context.Context) ([]*model.MenuM, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.MenuM, error)
	GetByIDsWithApis(ctx context.Context, ids []uint) ([]*model.MenuM, error)
	GetByParentID(ctx context.Context, parentID uint) ([]*model.MenuM, error)
	FilterByParentID(ctx context.Context, all []*model.MenuM, parentID uint) ([]*model.MenuM, error)
	GetChildren(ctx context.Context, all []*model.MenuM, menu *model.MenuM) error
	Tree(ctx context.Context, all []*model.MenuM) ([]*model.MenuM, error)
	CreateWithApis(ctx context.Context, menu *model.MenuM) error
	FirstOrCreateWithApis(ctx context.Context, menu *model.MenuM) error
	UpdateWithApis(ctx context.Context, menu *model.MenuM) error
}

type sysMenuStore struct {
	*genericstore.Store[model.MenuM]
}

var _ SysMenuStore = (*sysMenuStore)(nil)

func NewSysMenuStore(store *datastore) *sysMenuStore {
	return &sysMenuStore{
		Store: genericstore.NewStore[model.MenuM](store, NewLogger()),
	}
}

// ListWithRequest lists menus based on request parameters.
func (s *sysMenuStore) ListWithRequest(ctx context.Context, req *v1.ListMenuRequest) (int64, []*model.MenuM, error) {
	db := s.DB(ctx)
	var ret []*model.MenuM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByID retrieves a menu by ID.
func (s *sysMenuStore) GetByID(ctx context.Context, id uint) (*model.MenuM, error) {
	return s.Get(ctx, where.F("id", id))
}

// GetByIDWithApis retrieves a menu by ID with preloaded Apis.
func (s *sysMenuStore) GetByIDWithApis(ctx context.Context, id uint) (*model.MenuM, error) {
	var ret model.MenuM
	err := s.DB(ctx).Preload("Apis").Where("id = ?", id).First(&ret).Error

	return &ret, err
}

// DeleteByID deletes a menu by ID.
func (s *sysMenuStore) DeleteByID(ctx context.Context, id uint) error {
	return s.Delete(ctx, where.F("id", id))
}

// All retrieves all menus ordered by sort.
func (s *sysMenuStore) All(ctx context.Context) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).Order("sort asc").Find(&ret).Error

	return ret, err
}

// AllEnabled retrieves all enabled menus ordered by sort.
// Hidden field is not filtered here as it's a frontend display concern.
func (s *sysMenuStore) AllEnabled(ctx context.Context) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Where("status = ?", "enabled").
		Order("sort asc").
		Find(&ret).
		Error

	return ret, err
}

// AllWithApis retrieves all menus with preloaded Apis.
func (s *sysMenuStore) AllWithApis(ctx context.Context) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Preload("Apis").
		Order("sort asc").
		Find(&ret).
		Error

	return ret, err
}

// GetByIDs retrieves enabled menus by IDs.
// Hidden field is not filtered here as it's a frontend display concern.
func (s *sysMenuStore) GetByIDs(ctx context.Context, ids []uint) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Where("id IN ?", ids).
		Where("status = ?", "enabled").
		Order("sort asc").
		Find(&ret).
		Error

	return ret, err
}

// GetByIDsWithApis retrieves menus by IDs with preloaded Apis.
func (s *sysMenuStore) GetByIDsWithApis(ctx context.Context, ids []uint) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Preload("Apis").
		Where("id IN ?", ids).
		Order("sort asc").
		Find(&ret).
		Error

	return ret, err
}

// GetByParentID retrieves menus by parent ID.
func (s *sysMenuStore) GetByParentID(ctx context.Context, parentID uint) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).Where(&model.MenuM{ParentID: parentID}).Find(&ret).Error

	return ret, err
}

// FilterByParentID filters menus by parent ID from a given list.
func (s *sysMenuStore) FilterByParentID(_ context.Context, all []*model.MenuM, parentID uint) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	for _, item := range all {
		if item.ParentID != parentID {
			continue
		}
		ret = append(ret, item)
	}

	return ret, nil
}

// GetChildren recursively gets children menus and assigns them to the menu's Children field.
func (s *sysMenuStore) GetChildren(ctx context.Context, all []*model.MenuM, menu *model.MenuM) error {
	children, err := s.FilterByParentID(ctx, all, menu.ID)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return nil
	}

	menu.Children = children
	for key := range menu.Children {
		item := menu.Children[key]
		err := s.GetChildren(ctx, all, item)
		if err != nil {
			return err
		}
	}

	return nil
}

// Tree builds a menu tree from a flat list of menus.
func (s *sysMenuStore) Tree(ctx context.Context, all []*model.MenuM) ([]*model.MenuM, error) {
	ret, err := s.FilterByParentID(ctx, all, 0)
	if err != nil {
		return nil, err
	}

	if len(ret) == 0 {
		return ret, nil
	}

	for key := range ret {
		item := ret[key]
		err := s.GetChildren(ctx, all, item)
		if err != nil {
			return ret, err
		}
	}

	return ret, nil
}

// CreateWithApis creates a menu with associated APIs.
func (s *sysMenuStore) CreateWithApis(ctx context.Context, menu *model.MenuM) error {
	return s.DB(ctx).Create(menu).Error
}

// FirstOrCreateWithApis finds or creates a menu by name, and updates API associations.
func (s *sysMenuStore) FirstOrCreateWithApis(ctx context.Context, menu *model.MenuM) error {
	var existing model.MenuM
	result := s.DB(ctx).Where("name = ?", menu.Name).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		// Menu exists, update ID for parent lookup
		menu.ID = existing.ID
		return nil
	}

	// Create new menu with APIs
	return s.DB(ctx).Create(menu).Error
}

// UpdateWithApis updates a menu and replaces its API associations.
func (s *sysMenuStore) UpdateWithApis(ctx context.Context, menu *model.MenuM) error {
	db := s.DB(ctx)

	// Replace API associations
	if err := db.Model(menu).Association("Apis").Replace(menu.Apis); err != nil {
		return err
	}

	// Update menu fields
	return db.Save(menu).Error
}
