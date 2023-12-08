package system

import (
	"context"

	"gorm.io/gorm"

	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
)

type PermissionStore interface {
	List(ctx context.Context, req *v1.ListPermissionRequest) (int64, []*model.PermissionM, error)
	Create(ctx context.Context, permission *model.PermissionM) error
	Get(ctx context.Context, ID uint) (*model.PermissionM, error)
	Update(ctx context.Context, permission *model.PermissionM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*model.PermissionM, error)
	GetByIDs(ctx context.Context, IDs []uint) (ret []*model.PermissionM, err error)
	GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error)
}

type permissions struct {
	db *gorm.DB
}

var _ PermissionStore = (*permissions)(nil)

func NewPermissions(db *gorm.DB) *permissions {
	return &permissions{db: db}
}

func (u *permissions) List(ctx context.Context, req *v1.ListPermissionRequest) (count int64, ret []*model.PermissionM, err error) {
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

func (u *permissions) Create(ctx context.Context, permission *model.PermissionM) error {
	return u.db.Create(&permission).Error
}

func (u *permissions) Get(ctx context.Context, ID uint) (permission *model.PermissionM, err error) {
	err = u.db.Where("id = ?", ID).First(&permission).Error

	return
}

func (u *permissions) Update(ctx context.Context, permission *model.PermissionM, fields ...string) error {
	return u.db.Select(fields).Save(&permission).Error
}

func (u *permissions) Delete(ctx context.Context, ID uint) error {
	return u.db.Where("id = ?", ID).Delete(&model.PermissionM{}).Error
}

func (u *permissions) All(ctx context.Context) (ret []*model.PermissionM, err error) {
	err = u.db.Find(&ret).Error

	return
}

func (u *permissions) GetByIDs(ctx context.Context, IDs []uint) (ret []*model.PermissionM, err error) {
	err = u.db.Where("id IN ?", IDs).Find(&ret).Error

	return
}

func (u *permissions) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error) {
	err = u.db.Model(&model.PermissionM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return
}
