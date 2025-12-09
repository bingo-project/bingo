package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type UserStore interface {
	Create(ctx context.Context, obj *model.UserM) error
	Update(ctx context.Context, obj *model.UserM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserM, error)

	UserExpansion
}

// UserExpansion defines additional methods for user operations.
type UserExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListUserRequest) (int64, []*model.UserM, error)
	IsExist(ctx context.Context, user *model.UserM) (bool, error)
	GetByUID(ctx context.Context, uid string) (*model.UserM, error)
	GetByUsername(ctx context.Context, username string) (*model.UserM, error)
	DeleteByUsername(ctx context.Context, username string) error
	CreateWithAccount(ctx context.Context, user *model.UserM, account *model.UserAccount) error
	FindAccounts(ctx context.Context, uid string) ([]*model.UserAccount, error)
	CountAccounts(ctx context.Context, uid string) (int64, error)
	FirstOrCreate(ctx context.Context, where *model.UserM, user *model.UserM) error
}

type userStore struct {
	*genericstore.Store[model.UserM]
}

var _ UserStore = (*userStore)(nil)

func NewUserStore(store *datastore) *userStore {
	return &userStore{
		Store: genericstore.NewStore[model.UserM](store, NewLogger()),
	}
}

// ListWithRequest lists users based on request parameters.
func (s *userStore) ListWithRequest(ctx context.Context, req *v1.ListUserRequest) (int64, []*model.UserM, error) {
	db := s.DB(ctx)
	var ret []*model.UserM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// IsExist checks if a user exists by UID, username, email, or phone.
func (s *userStore) IsExist(ctx context.Context, user *model.UserM) (bool, error) {
	db := s.DB(ctx).Model(&model.UserM{})

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
	err := db.Select("ID").Take(&id).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	return id > 0, nil
}

// GetByUID retrieves a user by UID.
func (s *userStore) GetByUID(ctx context.Context, uid string) (*model.UserM, error) {
	var user model.UserM
	err := s.DB(ctx).Where("uid = ?", uid).First(&user).Error

	return &user, err
}

// GetByUsername retrieves a user by username.
func (s *userStore) GetByUsername(ctx context.Context, username string) (*model.UserM, error) {
	var user model.UserM
	err := s.DB(ctx).Where("username = ?", username).First(&user).Error

	return &user, err
}

// DeleteByUsername deletes a user by username.
func (s *userStore) DeleteByUsername(ctx context.Context, username string) error {
	return s.DB(ctx).Where("username = ?", username).Delete(&model.UserM{}).Error
}

// CreateWithAccount creates a user with an associated account in a transaction.
func (s *userStore) CreateWithAccount(ctx context.Context, user *model.UserM, account *model.UserAccount) error {
	return s.DB(ctx).Transaction(func(tx *gorm.DB) error {
		// Check if account exists.
		err := tx.Model(&model.UserAccount{}).
			Where("provider = ?", account.Provider).
			Where("account_id = ?", account.AccountID).
			First(&account).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Update account info if exists
		if account.ID > 0 {
			tx.Where("uid = ?", account.UID).Updates(&model.UserM{
				LastLoginTime: user.LastLoginTime,
				LastLoginIP:   user.LastLoginIP,
				LastLoginType: user.LastLoginType,
			})

			return nil
		}

		// Create account
		err = tx.Create(account).Error
		if err != nil {
			return err
		}

		return tx.Create(user).Error
	})
}

// FindAccounts retrieves all accounts for a user by UID.
func (s *userStore) FindAccounts(ctx context.Context, uid string) ([]*model.UserAccount, error) {
	var ret []*model.UserAccount
	err := s.DB(ctx).
		Where("uid = ?", uid).
		Find(&ret).
		Error

	return ret, err
}

// CountAccounts counts the number of accounts for a user by UID.
func (s *userStore) CountAccounts(ctx context.Context, uid string) (int64, error) {
	var count int64
	err := s.DB(ctx).
		Model(&model.UserAccount{}).
		Where("uid = ?", uid).
		Count(&count).
		Error

	return count, err
}

// FirstOrCreate finds first record matching the given conditions or creates a new one.
func (s *userStore) FirstOrCreate(ctx context.Context, where *model.UserM, user *model.UserM) error {
	return s.DB(ctx).Where(where).FirstOrCreate(user).Error
}
