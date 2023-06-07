package system

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"bingo/internal/pkg/model/system"
	"bingo/internal/pkg/util/helper"
)

type AdminStore interface {
	List(ctx context.Context, offset, limit int) (int64, []*system.AdminM, error)
	Create(ctx context.Context, admin *system.AdminM) error
	Get(ctx context.Context, username string) (*system.AdminM, error)
	Update(ctx context.Context, admin *system.AdminM) error
	Delete(ctx context.Context, username string) error

	InitData(ctx context.Context) error
	CheckExist(ctx context.Context, admin *system.AdminM) (exist bool, err error)
	HasRole(ctx context.Context, admin *system.AdminM, roleName string) bool
	GetUserInfo(ctx context.Context, username string) (admin *system.AdminM, err error)
}

type admins struct {
	db *gorm.DB
}

// 确保 admins 实现了 AdminStore 接口.
var _ AdminStore = (*admins)(nil)

func NewAdmins(db *gorm.DB) *admins {
	return &admins{db: db}
}

func (u *admins) Create(ctx context.Context, admin *system.AdminM) error {
	return u.db.Create(&admin).Error
}

func (u *admins) Get(ctx context.Context, username string) (admin *system.AdminM, err error) {
	err = u.db.Where("username = ?", username).First(&admin).Error

	return
}

func (u *admins) Update(ctx context.Context, admin *system.AdminM) error {
	if len(admin.Roles) > 0 {
		err := u.db.Model(&admin).Association("Roles").Replace(admin.Roles)
		if err != nil {
			return err
		}
	}

	return u.db.Save(&admin).Error
}

func (u *admins) List(ctx context.Context, offset, limit int) (count int64, ret []*system.AdminM, err error) {
	err = u.db.Offset(offset).
		Limit(helper.DefaultLimit(limit)).
		Order("id desc").
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (u *admins) Delete(ctx context.Context, username string) error {
	err := u.db.Where("username = ?", username).Delete(&system.AdminM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (u *admins) InitData(ctx context.Context) error {
	admin := system.AdminM{
		Username: "root",
		Password: "123456",
		Nickname: "Root",
		Email:    "root@root.com",
		Phone:    "18800000000",
		RoleName: "root",
	}

	// Check exist
	resp, err := u.Get(ctx, admin.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if resp.ID > 0 {
		return errors.New("admin:" + admin.Username + " already exist")
	}

	return u.db.Create(&admin).Error
}

func (u *admins) CheckExist(ctx context.Context, admin *system.AdminM) (exist bool, err error) {
	var id uint

	if admin.Username != "" {
		u.db.Model(&admin).Where("username = ?", admin.Username).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	if admin.Email != "" {
		u.db.Model(&admin).Where("email = ?", admin.Email).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	u.db.Model(&admin).Where("phone = ?", admin.Phone).Select("id").First(&id)

	return id > 0, nil
}

func (u *admins) HasRole(ctx context.Context, admin *system.AdminM, roleName string) bool {
	count := u.db.Model(&admin).Where("role_name = ?", roleName).Association("Roles").Count()

	return count > 0
}

func (u *admins) GetUserInfo(ctx context.Context, username string) (admin *system.AdminM, err error) {
	err = u.db.Preload("Role").
		Preload("Roles").
		Where("username = ?", username).
		First(&admin).
		Error

	return
}
