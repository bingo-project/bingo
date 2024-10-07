package system

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/model"
)

type RoleStore interface {
	List(ctx context.Context, req *v1.ListRoleRequest) (int64, []*model.RoleM, error)
	Create(ctx context.Context, role *model.RoleM) error
	Get(ctx context.Context, roleName string) (*model.RoleM, error)
	Update(ctx context.Context, role *model.RoleM, fields ...string) error
	Delete(ctx context.Context, roleName string) error

	GetByNames(ctx context.Context, names []string) ([]model.RoleM, error)
	GetWithMenus(ctx context.Context, roleName string) (role *model.RoleM, err error)

	All(ctx context.Context) (ret []*model.RoleM, err error)
}

type roles struct {
	db *gorm.DB
}

var _ RoleStore = (*roles)(nil)

func NewRoles(db *gorm.DB) *roles {
	return &roles{db: db}
}

func SearchRole(req *v1.ListRoleRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Name != "" {
			db.Where("name = ?", req.Name)
		}

		return db
	}
}

func (s *roles) List(ctx context.Context, req *v1.ListRoleRequest) (count int64, ret []*model.RoleM, err error) {
	db := s.db.WithContext(ctx).
		Scopes(SearchRole(req)).
		Where("name != ?", global.RoleRoot)

	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *roles) Create(ctx context.Context, role *model.RoleM) error {
	return s.db.WithContext(ctx).Create(&role).Error
}

func (s *roles) Get(ctx context.Context, roleName string) (role *model.RoleM, err error) {
	err = s.db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error

	return
}

func (s *roles) Update(ctx context.Context, role *model.RoleM, fields ...string) error {
	err := s.db.WithContext(ctx).Model(&role).Association("Menus").Replace(role.Menus)
	if err != nil {
		return err
	}

	return s.db.WithContext(ctx).Select(fields).Save(&role).Error
}

func (s *roles) Delete(ctx context.Context, roleName string) error {
	return s.db.WithContext(ctx).Where("name = ?", roleName).Delete(&model.RoleM{}).Error
}

func (s *roles) GetByNames(ctx context.Context, names []string) (ret []model.RoleM, err error) {
	err = s.db.WithContext(ctx).Where("name IN ?", names).Find(&ret).Error

	return
}

func (s *roles) GetWithMenus(ctx context.Context, roleName string) (role *model.RoleM, err error) {
	err = s.db.WithContext(ctx).
		Preload("Menus", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort asc")
		}).
		Where("name = ?", roleName).
		First(&role).
		Error

	return
}

func (s *roles) All(ctx context.Context) (ret []*model.RoleM, err error) {
	err = s.db.WithContext(ctx).Find(&ret).Error

	return
}
