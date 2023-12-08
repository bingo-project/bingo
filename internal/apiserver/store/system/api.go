package system

import (
	"context"

	"gorm.io/gorm"

	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
)

type ApiStore interface {
	List(ctx context.Context, req *v1.ListApiRequest) (int64, []*model.ApiM, error)
	Create(ctx context.Context, api *model.ApiM) error
	Get(ctx context.Context, ID uint) (*model.ApiM, error)
	Update(ctx context.Context, api *model.ApiM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*model.ApiM, error)
	GetByIDs(ctx context.Context, IDs []uint) (ret []*model.ApiM, err error)
	GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error)
}

type apis struct {
	db *gorm.DB
}

var _ ApiStore = (*apis)(nil)

func NewApis(db *gorm.DB) *apis {
	return &apis{db: db}
}

func (u *apis) List(ctx context.Context, req *v1.ListApiRequest) (count int64, ret []*model.ApiM, err error) {
	// Order
	if req.Order == "" {
		req.Order = "id"
	}

	// Sort
	if req.Sort == "" {
		req.Sort = "desc"
	}

	err = u.db.Offset(req.Offset).
		Limit(helper.DefaultLimit(req.Limit)).
		Order(req.Order + " " + req.Sort).
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (u *apis) Create(ctx context.Context, api *model.ApiM) error {
	return u.db.Create(&api).Error
}

func (u *apis) Get(ctx context.Context, ID uint) (api *model.ApiM, err error) {
	err = u.db.Where("id = ?", ID).First(&api).Error

	return
}

func (u *apis) Update(ctx context.Context, api *model.ApiM, fields ...string) error {
	return u.db.Select(fields).Save(&api).Error
}

func (u *apis) Delete(ctx context.Context, ID uint) error {
	return u.db.Where("id = ?", ID).Delete(&model.ApiM{}).Error
}

func (u *apis) All(ctx context.Context) (ret []*model.ApiM, err error) {
	err = u.db.Find(&ret).Error

	return
}

func (u *apis) GetByIDs(ctx context.Context, IDs []uint) (ret []*model.ApiM, err error) {
	err = u.db.Where("id IN ?", IDs).Find(&ret).Error

	return
}

func (u *apis) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error) {
	err = u.db.Model(&model.ApiM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return
}
