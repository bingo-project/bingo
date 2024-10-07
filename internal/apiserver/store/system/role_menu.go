package system

import (
	"context"

	"github.com/ahmetb/go-linq/v3"
	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/global"
	model2 "bingo/internal/pkg/model"
)

type RoleMenuStore interface {
	List(ctx context.Context, req *v1.ListRoleMenuRequest) (int64, []*model2.RoleMenuM, error)
	Create(ctx context.Context, roleMenu *model2.RoleMenuM) error
	Get(ctx context.Context, ID uint) (*model2.RoleMenuM, error)
	Update(ctx context.Context, roleMenu *model2.RoleMenuM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, roleMenus []*model2.RoleMenuM) error
	FirstOrCreate(ctx context.Context, where any, roleMenu *model2.RoleMenuM) error

	GetMenuIDsByRoleName(ctx context.Context, roleName string) ([]uint, error)
	GetMenuIDsByRoleNameWithParent(ctx context.Context, roleName string) (ret []uint, err error)
}

type roleMenus struct {
	db *gorm.DB
}

var _ RoleMenuStore = (*roleMenus)(nil)

func NewRoleMenus(db *gorm.DB) *roleMenus {
	return &roleMenus{db: db}
}

func (s *roleMenus) List(ctx context.Context, req *v1.ListRoleMenuRequest) (count int64, ret []*model2.RoleMenuM, err error) {
	count, err = gormutil.Paginate(s.db.WithContext(ctx), &req.ListOptions, &ret)

	return
}

func (s *roleMenus) Create(ctx context.Context, roleMenu *model2.RoleMenuM) error {
	return s.db.WithContext(ctx).Create(&roleMenu).Error
}

func (s *roleMenus) Get(ctx context.Context, ID uint) (roleMenu *model2.RoleMenuM, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&roleMenu).Error

	return
}

func (s *roleMenus) Update(ctx context.Context, roleMenu *model2.RoleMenuM, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&roleMenu).Error
}

func (s *roleMenus) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model2.RoleMenuM{}).Error
}

func (s *roleMenus) CreateInBatch(ctx context.Context, roleMenus []*model2.RoleMenuM) error {
	return s.db.WithContext(ctx).CreateInBatches(&roleMenus, global.CreateBatchSize).Error
}

func (s *roleMenus) FirstOrCreate(ctx context.Context, where any, roleMenu *model2.RoleMenuM) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&roleMenu).
		FirstOrCreate(&roleMenu).
		Error
}

func (s *roleMenus) GetMenuIDsByRoleName(ctx context.Context, roleName string) (ret []uint, err error) {
	err = s.db.WithContext(ctx).
		Model(&model2.RoleMenuM{}).
		Select("menu_id").
		Where(&model2.RoleMenuM{RoleName: roleName}).
		Find(&ret).
		Error

	return
}

func (s *roleMenus) GetMenuIDsByRoleNameWithParent(ctx context.Context, roleName string) (ret []uint, err error) {
	var menuIDs []uint
	err = s.db.WithContext(ctx).
		Model(&model2.RoleMenuM{}).
		Select("menu_id").
		Where(&model2.RoleMenuM{RoleName: roleName}).
		Find(&menuIDs).
		Error
	if err != nil {
		return
	}

	var parentIDs []uint
	err = s.db.WithContext(ctx).
		Model(&model2.MenuM{}).
		Where("id IN (?)", menuIDs).
		Select("parent_id").
		Find(&parentIDs).
		Error

	// Union menuIDs & parentIDs
	linq.From(menuIDs).
		Union(linq.From(parentIDs)).
		ToSlice(&ret)

	return
}
