package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	v1 "bingo/internal/apiserver/http/request/v1"
	model2 "bingo/internal/pkg/model"
)

type UserStore interface {
	List(ctx context.Context, req *v1.ListUserRequest) (int64, []*model2.UserM, error)
	Create(ctx context.Context, user *model2.UserM) error
	Get(ctx context.Context, username string) (*model2.UserM, error)
	Update(ctx context.Context, user *model2.UserM, fields ...string) error
	Delete(ctx context.Context, username string) error

	FirstOrCreate(ctx context.Context, where any, user *model2.UserM) error

	IsExist(ctx context.Context, user *model2.UserM) (exist bool, err error)
	GetByUID(ctx context.Context, uid string) (user *model2.UserM, err error)

	CreateWithAccount(ctx context.Context, user *model2.UserM, account *model2.UserAccount) error
	FindAccounts(ctx context.Context, uid string) ([]*model2.UserAccount, error)
	CountAccounts(ctx context.Context, uid string) (ret int64, err error)
}

type users struct {
	db *gorm.DB
}

var _ UserStore = (*users)(nil)

func newUsers(db *gorm.DB) *users {
	return &users{db: db}
}

func (u *users) List(ctx context.Context, req *v1.ListUserRequest) (count int64, ret []*model2.UserM, err error) {
	count, err = gormutil.Paginate(u.db.WithContext(ctx), &req.ListOptions, &ret)

	return
}

func (u *users) Create(ctx context.Context, user *model2.UserM) error {
	return u.db.WithContext(ctx).Create(&user).Error
}

func (u *users) Get(ctx context.Context, username string) (user *model2.UserM, err error) {
	err = u.db.WithContext(ctx).Where("username = ?", username).First(&user).Error

	return
}

func (u *users) Update(ctx context.Context, user *model2.UserM, fields ...string) error {
	return u.db.WithContext(ctx).Select(fields).Save(&user).Error
}

func (u *users) Delete(ctx context.Context, username string) error {
	return u.db.WithContext(ctx).Where("username = ?", username).Delete(&model2.UserM{}).Error
}

func (u *users) FirstOrCreate(ctx context.Context, where any, user *model2.UserM) error {
	return u.db.WithContext(ctx).
		Where(where).
		Attrs(&user).
		FirstOrCreate(&user).
		Error
}

func (u *users) IsExist(ctx context.Context, user *model2.UserM) (exist bool, err error) {
	db := u.db.WithContext(ctx).Model(&model2.UserM{})

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

func (u *users) GetByUID(ctx context.Context, uid string) (user *model2.UserM, err error) {
	err = u.db.WithContext(ctx).Where("uid = ?", uid).First(&user).Error

	return
}

func (u *users) CreateWithAccount(ctx context.Context, user *model2.UserM, account *model2.UserAccount) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check exist.
		err := tx.Model(&model2.UserAccount{}).
			Where("provider = ?", account.Provider).
			Where("account_id = ?", account.AccountID).
			First(&account).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Update account info
		if account.ID > 0 {
			tx.Where("uid = ?", account.UID).Updates(&model2.UserM{
				LastLoginTime: user.LastLoginTime,
				LastLoginIP:   user.LastLoginIP,
				LastLoginType: user.LastLoginType,
			})

			return nil
		}

		// Create user
		err = tx.Create(account).Error
		if err != nil {
			return err
		}

		return tx.Create(user).Error
	})
}

func (u *users) FindAccounts(ctx context.Context, uid string) (ret []*model2.UserAccount, err error) {
	err = u.db.WithContext(ctx).
		Where("uid = ?", uid).
		Find(&ret).
		Error

	return
}

func (u *users) CountAccounts(ctx context.Context, uid string) (ret int64, err error) {
	err = u.db.WithContext(ctx).
		Model(&model2.UserAccount{}).
		Where("uid = ?", uid).
		Count(&ret).
		Error

	return
}
