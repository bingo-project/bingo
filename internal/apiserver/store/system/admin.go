package system

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
)

type AdminStore interface {
	List(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.AdminM, error)
	Create(ctx context.Context, admin *model.AdminM) error
	Get(ctx context.Context, username string) (*model.AdminM, error)
	Update(ctx context.Context, admin *model.AdminM) error
	Delete(ctx context.Context, username string) error

	InitData(ctx context.Context) error
	CheckExist(ctx context.Context, admin *model.AdminM) (exist bool, err error)
	HasRole(ctx context.Context, admin *model.AdminM, roleName string) bool
	GetUserInfo(ctx context.Context, username string) (admin *model.AdminM, err error)
}

type admins struct {
	db *gorm.DB
}

var _ AdminStore = (*admins)(nil)

func NewAdmins(db *gorm.DB) *admins {
	return &admins{db: db}
}

func SearchAdmin(req *v1.ListAdminRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Username != "" {
			db.Where("username = ?", req.Username)
		}
		if req.Nickname != "" {
			db.Where("nickname like ?", "%"+req.Nickname+"%")
		}
		if req.Status != nil {
			db.Where("status = ?", req.Status)
		}
		if req.RoleName != "" {
			db.Where("role_name = ?", req.RoleName)
		}
		if req.Email != "" {
			db.Where("email = ?", req.Email)
		}
		if req.Phone != "" {
			db.Where("phone = ?", req.Phone)
		}

		return db
	}
}

func (u *admins) List(ctx context.Context, req *v1.ListAdminRequest) (count int64, ret []*model.AdminM, err error) {
	db := u.db.Preload("Role").
		Preload("Roles").
		Scopes(SearchAdmin(req))

	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (u *admins) Create(ctx context.Context, admin *model.AdminM) error {
	return u.db.Create(&admin).Error
}

func (u *admins) Get(ctx context.Context, username string) (admin *model.AdminM, err error) {
	err = u.db.Where("username = ?", username).First(&admin).Error

	return
}

func (u *admins) Update(ctx context.Context, admin *model.AdminM) error {
	err := u.db.Model(&admin).Association("Roles").Replace(admin.Roles)
	if err != nil {
		return err
	}

	return u.db.Save(&admin).Error
}

func (u *admins) Delete(ctx context.Context, username string) error {
	err := u.db.Where("username = ?", username).Delete(&model.AdminM{}).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (u *admins) InitData(ctx context.Context) error {
	admin := model.AdminM{
		Username: "root",
		Password: "123456",
		Nickname: "Root",
		Email:    nil,
		Phone:    nil,
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

func (u *admins) CheckExist(ctx context.Context, admin *model.AdminM) (exist bool, err error) {
	var id uint

	if admin.Username != "" {
		u.db.Model(&admin).Where("username = ?", admin.Username).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	if admin.Email != nil {
		u.db.Model(&admin).Where("email = ?", admin.Email).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	u.db.Model(&admin).Where("phone = ?", admin.Phone).Select("id").First(&id)

	return id > 0, nil
}

func (u *admins) HasRole(ctx context.Context, admin *model.AdminM, roleName string) bool {
	count := u.db.Model(&admin).Where("role_name = ?", roleName).Association("Roles").Count()

	return count > 0
}

func (u *admins) GetUserInfo(ctx context.Context, username string) (admin *model.AdminM, err error) {
	err = u.db.Preload("Role").
		Preload("Roles").
		Where("username = ?", username).
		First(&admin).
		Error

	return
}
