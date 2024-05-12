package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/model"
)

type UserStore interface {
	List(ctx context.Context, req *v1.ListUserRequest) (int64, []*model.UserM, error)
	Create(ctx context.Context, user *model.UserM) error
	Get(ctx context.Context, username string) (*model.UserM, error)
	Update(ctx context.Context, user *model.UserM, fields ...string) error
	Delete(ctx context.Context, username string) error

	IsExist(ctx context.Context, user *model.UserM) (exist bool, err error)
	GetByUID(ctx context.Context, uid string) (user *model.UserM, err error)
}

type users struct {
	db *gorm.DB
}

var _ UserStore = (*users)(nil)

func newUsers(db *gorm.DB) *users {
	return &users{db: db}
}

func (u *users) List(ctx context.Context, req *v1.ListUserRequest) (count int64, ret []*model.UserM, err error) {
	count, err = gormutil.Paginate(u.db.WithContext(ctx), &req.ListOptions, &ret)

	return
}

func (u *users) Create(ctx context.Context, user *model.UserM) error {
	return u.db.WithContext(ctx).Create(&user).Error
}

func (u *users) Get(ctx context.Context, username string) (user *model.UserM, err error) {
	err = u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error

	return
}

func (u *users) Update(ctx context.Context, user *model.UserM, fields ...string) error {
	return u.db.WithContext(ctx).Select(fields).Save(&user).Error
}

func (u *users) Delete(ctx context.Context, username string) error {
	return u.db.WithContext(ctx).Where("username = ?", username).Delete(&model.UserM{}).Error
}

func (u *users) IsExist(ctx context.Context, user *model.UserM) (exist bool, err error) {
	db := u.db.WithContext(ctx).Model(&model.UserM{})

	if user.UID != "" {
		db = db.Where("uid = ?", user.UID)
	}
	if user.Username != "" {
		db = db.Where("username = ?", user.Username)
	}
	if user.Email != "" {
		db = db.Where("email = ?", user.Email)
	}
	if user.Phone != "" {
		db = db.Where("phone = ?", user.Phone)
	}

	var id int
	err = db.Select("ID").Take(&id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	return id > 0, nil
}

func (u *users) GetByUID(ctx context.Context, uid string) (user *model.UserM, err error) {
	err = u.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error

	return
}
