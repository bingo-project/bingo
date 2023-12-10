package system

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
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
	GetByIDs(ctx context.Context, ids []uint) (ret []model.MenuM, err error)
}

type menus struct {
	db *gorm.DB
}

var _ MenuStore = (*menus)(nil)

func NewMenus(db *gorm.DB) *menus {
	return &menus{db: db}
}

func (s *menus) List(ctx context.Context, req *v1.ListMenuRequest) (count int64, ret []*model.MenuM, err error) {
	// Order
	if req.Order == "" {
		req.Order = "id"
	}

	// Sort
	if req.Sort == "" {
		req.Sort = "desc"
	}

	err = s.db.Offset(req.Offset).
		Limit(helper.DefaultLimit(req.Limit)).
		Order(req.Order + " " + req.Sort).
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (s *menus) Create(ctx context.Context, menu *model.MenuM) error {
	return s.db.Create(&menu).Error
}

func (s *menus) Get(ctx context.Context, ID uint) (menu *model.MenuM, err error) {
	err = s.db.Where("id = ?", ID).First(&menu).Error

	return
}

func (s *menus) Update(ctx context.Context, menu *model.MenuM, fields ...string) error {
	return s.db.Select(fields).Save(&menu).Error
}

func (s *menus) Delete(ctx context.Context, ID uint) error {
	return s.db.Where("id = ?", ID).Delete(&model.MenuM{}).Error
}

func (s *menus) CreateInBatch(ctx context.Context, menus []*model.MenuM) error {
	return s.db.CreateInBatches(&menus, global.CreateBatchSize).Error
}

func (s *menus) FirstOrCreate(ctx context.Context, where any, menu *model.MenuM) error {
	return s.db.Where(where).
		Attrs(&menu).
		FirstOrCreate(&menu).
		Error
}

func (s *menus) UpdateOrCreate(ctx context.Context, where any, menu *model.MenuM) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var exist model.MenuM
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		menu.ID = exist.ID

		return tx.Save(&menu).Error
	})
}

func (s *menus) All(ctx context.Context) (ret []*model.MenuM, err error) {
	err = s.db.Find(&ret).Error

	return
}

func (s *menus) GetByIDs(ctx context.Context, ids []uint) (ret []model.MenuM, err error) {
	err = s.db.Where("id IN ?", ids).Find(&ret).Error

	return
}
