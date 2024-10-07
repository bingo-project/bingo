package system

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
)

type MenuStore interface {
	List(ctx context.Context, req *v1.ListMenuRequest) (int64, []*model.MenuM, error)
	Create(ctx context.Context, menu *model.MenuM) error
	Get(ctx context.Context, ID uint) (*model.MenuM, error)
	Update(ctx context.Context, menu *model.MenuM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, menus []*model.MenuM) error
	FirstOrCreate(ctx context.Context, where any, menu *model.MenuM) error
	UpdateOrCreate(ctx context.Context, where any, menu *model.MenuM) error

	All(ctx context.Context) (ret []*model.MenuM, err error)
	GetByIDs(ctx context.Context, ids []uint) (ret []*model.MenuM, err error)
	GetByParentID(ctx context.Context, parentID uint) (ret []*model.MenuM, err error)

	FilterByParentID(ctx context.Context, all []*model.MenuM, parentID uint) (ret []*model.MenuM, err error)
	GetChildren(ctx context.Context, all []*model.MenuM, menuM *model.MenuM) error
	Tree(ctx context.Context, all []*model.MenuM) (ret []*model.MenuM, err error)
}

type menus struct {
	db *gorm.DB
}

var _ MenuStore = (*menus)(nil)

func NewMenus(db *gorm.DB) *menus {
	return &menus{db: db}
}

func (s *menus) List(ctx context.Context, req *v1.ListMenuRequest) (count int64, ret []*model.MenuM, err error) {
	count, err = gormutil.Paginate(s.db.WithContext(ctx), &req.ListOptions, &ret)

	return
}

func (s *menus) Create(ctx context.Context, menu *model.MenuM) error {
	return s.db.WithContext(ctx).Create(&menu).Error
}

func (s *menus) Get(ctx context.Context, ID uint) (menu *model.MenuM, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&menu).Error

	return
}

func (s *menus) Update(ctx context.Context, menu *model.MenuM, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&menu).Error
}

func (s *menus) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.MenuM{}).Error
}

func (s *menus) CreateInBatch(ctx context.Context, menus []*model.MenuM) error {
	return s.db.WithContext(ctx).CreateInBatches(&menus, global.CreateBatchSize).Error
}

func (s *menus) FirstOrCreate(ctx context.Context, where any, menu *model.MenuM) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&menu).
		FirstOrCreate(&menu).
		Error
}

func (s *menus) UpdateOrCreate(ctx context.Context, where any, menu *model.MenuM) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.MenuM
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		menu.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&menu).Error
	})
}

func (s *menus) All(ctx context.Context) (ret []*model.MenuM, err error) {
	err = s.db.WithContext(ctx).Order("sort asc").Find(&ret).Error

	return
}

func (s *menus) GetByIDs(ctx context.Context, ids []uint) (ret []*model.MenuM, err error) {
	err = s.db.WithContext(ctx).
		Where("id IN ?", ids).
		Order("sort asc").
		Find(&ret).
		Error

	return
}

func (s *menus) GetByParentID(ctx context.Context, parentID uint) (ret []*model.MenuM, err error) {
	err = s.db.WithContext(ctx).Where(&model.MenuM{ParentID: parentID}).Find(&ret).Error

	return
}

func (s *menus) FilterByParentID(ctx context.Context, all []*model.MenuM, parentID uint) (ret []*model.MenuM, err error) {
	for _, item := range all {
		if item.ParentID != parentID {
			continue
		}

		ret = append(ret, item)
	}

	return
}

func (s *menus) GetChildren(ctx context.Context, all []*model.MenuM, menuM *model.MenuM) error {
	children, err := s.FilterByParentID(ctx, all, menuM.ID)
	if err != nil {
		return err
	}

	if len(children) == 0 {
		return nil
	}

	menuM.Children = children
	for key := range menuM.Children {
		item := menuM.Children[key]
		err := s.GetChildren(ctx, all, item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *menus) Tree(ctx context.Context, all []*model.MenuM) (ret []*model.MenuM, err error) {
	ret, err = s.FilterByParentID(ctx, all, 0)
	if err != nil {
		return
	}

	if len(ret) == 0 {
		return
	}

	for key := range ret {
		item := ret[key]
		err := s.GetChildren(ctx, all, item)
		if err != nil {
			return ret, err
		}
	}

	return
}
