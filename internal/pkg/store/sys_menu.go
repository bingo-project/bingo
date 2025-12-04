// ABOUTME: System menu store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for menu management including tree building.

package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
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
	DeleteByID(ctx context.Context, id uint) error
	All(ctx context.Context) ([]*model.MenuM, error)
	AllEnabled(ctx context.Context) ([]*model.MenuM, error)
	GetByIDs(ctx context.Context, ids []uint) ([]*model.MenuM, error)
	GetByParentID(ctx context.Context, parentID uint) ([]*model.MenuM, error)
	FilterByParentID(ctx context.Context, all []*model.MenuM, parentID uint) ([]*model.MenuM, error)
	GetChildren(ctx context.Context, all []*model.MenuM, menu *model.MenuM) error
	Tree(ctx context.Context, all []*model.MenuM) ([]*model.MenuM, error)
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

// AllEnabled retrieves all enabled (visible) menus ordered by sort.
func (s *sysMenuStore) AllEnabled(ctx context.Context) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Where("hidden = ?", false).
		Order("sort asc").
		Find(&ret).
		Error

	return ret, err
}

// GetByIDs retrieves enabled menus by IDs.
func (s *sysMenuStore) GetByIDs(ctx context.Context, ids []uint) ([]*model.MenuM, error) {
	var ret []*model.MenuM
	err := s.DB(ctx).
		Where("id IN ?", ids).
		Where("hidden = ?", false).
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
