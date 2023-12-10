package system

import (
	"context"

	"gorm.io/gorm"

	"bingo/internal/apiserver/global"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
)

type RoleMenuStore interface {
	List(ctx context.Context, req *v1.ListRoleMenuRequest) (int64, []*model.RoleMenuM, error)
	Create(ctx context.Context, roleMenu *model.RoleMenuM) error
	Get(ctx context.Context, ID uint) (*model.RoleMenuM, error)
	Update(ctx context.Context, roleMenu *model.RoleMenuM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, roleMenus []*model.RoleMenuM) error
	FirstOrCreate(ctx context.Context, where any, roleMenu *model.RoleMenuM) error

	GetMenuIDsByRoleName(ctx context.Context, roleName string) ([]uint, error)
}

type roleMenus struct {
	db *gorm.DB
}

var _ RoleMenuStore = (*roleMenus)(nil)

func NewRoleMenus(db *gorm.DB) *roleMenus {
	return &roleMenus{db: db}
}

func (s *roleMenus) List(ctx context.Context, req *v1.ListRoleMenuRequest) (count int64, ret []*model.RoleMenuM, err error) {
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

func (s *roleMenus) Create(ctx context.Context, roleMenu *model.RoleMenuM) error {
	return s.db.Create(&roleMenu).Error
}

func (s *roleMenus) Get(ctx context.Context, ID uint) (roleMenu *model.RoleMenuM, err error) {
	err = s.db.Where("id = ?", ID).First(&roleMenu).Error

	return
}

func (s *roleMenus) Update(ctx context.Context, roleMenu *model.RoleMenuM, fields ...string) error {
	return s.db.Select(fields).Save(&roleMenu).Error
}

func (s *roleMenus) Delete(ctx context.Context, ID uint) error {
	return s.db.Where("id = ?", ID).Delete(&model.RoleMenuM{}).Error
}

func (s *roleMenus) CreateInBatch(ctx context.Context, roleMenus []*model.RoleMenuM) error {
	return s.db.CreateInBatches(&roleMenus, global.CreateBatchSize).Error
}

func (s *roleMenus) FirstOrCreate(ctx context.Context, where any, roleMenu *model.RoleMenuM) error {
	return s.db.Where(where).
		Attrs(&roleMenu).
		FirstOrCreate(&roleMenu).
		Error
}

func (s *roleMenus) GetMenuIDsByRoleName(ctx context.Context, roleName string) (ret []uint, err error) {
	err = s.db.Model(&model.RoleMenuM{}).
		Select("menu_id").
		Where(&model.RoleMenuM{RoleName: roleName}).
		Find(&ret).
		Error

	return
}
