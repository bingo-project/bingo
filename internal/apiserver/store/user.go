package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
)

type UserStore interface {
	List(ctx context.Context, req *v1.ListUserRequest) (int64, []*model.UserM, error)
	Create(ctx context.Context, user *model.UserM) error
	Get(ctx context.Context, username string) (*model.UserM, error)
	Update(ctx context.Context, user *model.UserM, fields ...string) error
	Delete(ctx context.Context, username string) error
}

type users struct {
	db *gorm.DB
}

var _ UserStore = (*users)(nil)

func newUsers(db *gorm.DB) *users {
	return &users{db: db}
}

func (u *users) List(ctx context.Context, req *v1.ListUserRequest) (count int64, ret []*model.UserM, err error) {
	count, err = gormutil.Paginate(u.db, &req.ListOptions, &ret)

	return
}

func (u *users) Create(ctx context.Context, user *model.UserM) error {
	return u.db.Create(&user).Error
}

func (u *users) Get(ctx context.Context, username string) (user *model.UserM, err error) {
	err = u.db.Where("username = ?", username).First(&user).Error

	return
}

func (u *users) Update(ctx context.Context, user *model.UserM, fields ...string) error {
	return u.db.Select(fields).Save(&user).Error
}

func (u *users) Delete(ctx context.Context, username string) error {
	return u.db.Where("username = ?", username).Delete(&model.UserM{}).Error
}
