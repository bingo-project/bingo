package system

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bingo/internal/pkg/model/system"
	"bingo/internal/pkg/util/helper"
)

type RoleStore interface {
	List(ctx context.Context, offset, limit int) (int64, []*system.RoleM, error)
	Create(ctx context.Context, role *system.RoleM) error
	Get(ctx context.Context, roleName string) (*system.RoleM, error)
	Update(ctx context.Context, role *system.RoleM) error
	Delete(ctx context.Context, roleName string) error
}

type roles struct {
	db *gorm.DB
}

// 确保 roles 实现了 RoleStore 接口.
var _ RoleStore = (*roles)(nil)

func NewRoles(db *gorm.DB) *roles {
	return &roles{db: db}
}

func (u *roles) Create(ctx context.Context, role *system.RoleM) error {
	return u.db.Create(&role).Error
}

func (u *roles) Get(ctx context.Context, roleName string) (role *system.RoleM, err error) {
	err = u.db.Where("name = ?", roleName).First(&role).Error

	return
}

func (u *roles) Update(ctx context.Context, role *system.RoleM) error {
	return u.db.Save(&role).Error
}

func (u *roles) List(ctx context.Context, offset, limit int) (count int64, ret []*system.RoleM, err error) {
	err = u.db.Offset(offset).Limit(helper.DefaultLimit(limit)).Order("id desc").Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (u *roles) Delete(ctx context.Context, roleName string) error {
	err := u.db.Where("name = ?", roleName).Delete(&system.RoleM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}
