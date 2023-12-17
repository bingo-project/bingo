package system

import (
	"context"

	"gorm.io/gorm"

	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
)

type RoleStore interface {
	List(ctx context.Context, req *v1.ListRoleRequest) (int64, []*model.RoleM, error)
	Create(ctx context.Context, role *model.RoleM) error
	Get(ctx context.Context, roleName string) (*model.RoleM, error)
	Update(ctx context.Context, role *model.RoleM, fields ...string) error
	Delete(ctx context.Context, roleName string) error

	GetByNames(ctx context.Context, names []string) ([]model.RoleM, error)
	GetWithMenus(ctx context.Context, roleName string) (role *model.RoleM, err error)
}

type roles struct {
	db *gorm.DB
}

var _ RoleStore = (*roles)(nil)

func NewRoles(db *gorm.DB) *roles {
	return &roles{db: db}
}

func (u *roles) List(ctx context.Context, req *v1.ListRoleRequest) (count int64, ret []*model.RoleM, err error) {
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

func (u *roles) Create(ctx context.Context, role *model.RoleM) error {
	return u.db.Create(&role).Error
}

func (u *roles) Get(ctx context.Context, roleName string) (role *model.RoleM, err error) {
	err = u.db.Where("name = ?", roleName).First(&role).Error

	return
}

func (u *roles) Update(ctx context.Context, role *model.RoleM, fields ...string) error {
	if len(role.Menus) > 0 {
		err := u.db.Model(&role).Association("Menus").Replace(role.Menus)
		if err != nil {
			return err
		}
	}

	return u.db.Select(fields).Save(&role).Error
}

func (u *roles) Delete(ctx context.Context, roleName string) error {
	return u.db.Where("name = ?", roleName).Delete(&model.RoleM{}).Error
}

func (u *roles) GetByNames(ctx context.Context, names []string) (ret []model.RoleM, err error) {
	err = u.db.Where("name IN ?", names).Find(&ret).Error

	return
}

func (u *roles) GetWithMenus(ctx context.Context, roleName string) (role *model.RoleM, err error) {
	err = u.db.Preload("Menus",
		func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Where("name = ?", roleName).
		First(&role).
		Error

	return
}
