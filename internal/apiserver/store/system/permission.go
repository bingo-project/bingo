package system

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bingo/internal/pkg/model/system"
	"bingo/internal/pkg/util/helper"
)

type PermissionStore interface {
	List(ctx context.Context, offset, limit int) (int64, []*system.PermissionM, error)
	Create(ctx context.Context, permission *system.PermissionM) error
	Get(ctx context.Context, ID uint) (*system.PermissionM, error)
	Update(ctx context.Context, permission *system.PermissionM) error
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*system.PermissionM, error)
	GetByIDs(ctx context.Context, IDs []uint) (ret []*system.PermissionM, err error)
	GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error)
}

type permissions struct {
	db *gorm.DB
}

// 确保 permissions 实现了 PermissionStore 接口.
var _ PermissionStore = (*permissions)(nil)

func NewPermissions(db *gorm.DB) *permissions {
	return &permissions{db: db}
}

func (u *permissions) Create(ctx context.Context, permission *system.PermissionM) error {
	return u.db.Create(&permission).Error
}

func (u *permissions) Get(ctx context.Context, ID uint) (permission *system.PermissionM, err error) {
	err = u.db.Where("id = ?", ID).First(&permission).Error

	return
}

func (u *permissions) Update(ctx context.Context, permission *system.PermissionM) error {
	return u.db.Save(&permission).Error
}

func (u *permissions) List(ctx context.Context, offset, limit int) (count int64, ret []*system.PermissionM, err error) {
	err = u.db.Offset(offset).Limit(helper.DefaultLimit(limit)).Order("id desc").Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (u *permissions) Delete(ctx context.Context, ID uint) error {
	err := u.db.Where("id = ?", ID).Delete(&system.PermissionM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (u *permissions) All(ctx context.Context) (ret []*system.PermissionM, err error) {
	err = u.db.Find(&ret).Error

	return
}

func (u *permissions) GetByIDs(ctx context.Context, IDs []uint) (ret []*system.PermissionM, err error) {
	err = u.db.Where("id IN ?", IDs).Find(&ret).Error

	return
}

func (u *permissions) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error) {
	err = u.db.Model(&system.PermissionM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return
}
