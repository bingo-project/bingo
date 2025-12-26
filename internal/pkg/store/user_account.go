package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// UserAccountStore defines the interface for user account operations.
type UserAccountStore interface {
	Create(ctx context.Context, obj *model.UserAccount) error
	Update(ctx context.Context, obj *model.UserAccount, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.UserAccount, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.UserAccount, error)

	UserAccountExpansion
}

// UserAccountExpansion defines additional methods for user account operations.
type UserAccountExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListUserAccountRequest) (int64, []*model.UserAccount, error)
	CheckExist(ctx context.Context, provider, accountID string) bool
	GetAccount(ctx context.Context, provider, accountID string) (*model.UserAccount, error)
	FirstOrCreate(ctx context.Context, where model.UserAccount, account *model.UserAccount) error
	FindByUID(ctx context.Context, uid string) ([]*model.UserAccount, error)
	FindByUIDAndProvider(ctx context.Context, uid, provider string) (*model.UserAccount, error)
	CountByUID(ctx context.Context, uid string) (int64, error)
	DeleteByID(ctx context.Context, id uint) error
}

type userAccountStore struct {
	*genericstore.Store[model.UserAccount]
}

var _ UserAccountStore = (*userAccountStore)(nil)

func NewUserAccountStore(store *datastore) *userAccountStore {
	return &userAccountStore{
		Store: genericstore.NewStore[model.UserAccount](store, NewLogger()),
	}
}

// ListWithRequest lists user accounts based on request parameters.
func (s *userAccountStore) ListWithRequest(ctx context.Context, req *v1.ListUserAccountRequest) (int64, []*model.UserAccount, error) {
	opts := where.NewWhere()

	if req.UID != nil {
		opts = opts.F("uid", *req.UID)
	}
	if req.Provider != nil {
		opts = opts.F("provider", *req.Provider)
	}
	if req.AccountID != nil {
		opts = opts.F("account_id", *req.AccountID)
	}
	if req.Username != nil {
		opts = opts.F("username", *req.Username)
	}
	if req.Nickname != nil {
		opts = opts.F("nickname", *req.Nickname)
	}
	if req.Email != nil {
		opts = opts.F("email", *req.Email)
	}
	if req.Bio != nil {
		opts = opts.F("bio", *req.Bio)
	}
	if req.Avatar != nil {
		opts = opts.F("avatar", *req.Avatar)
	}

	db := s.DB(ctx, opts)
	var ret []*model.UserAccount
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// CheckExist checks if a user account exists by provider and account ID.
func (s *userAccountStore) CheckExist(ctx context.Context, provider, accountID string) bool {
	var id int64
	s.DB(ctx).
		Model(&model.UserAccount{}).
		Where("provider = ?", provider).
		Where("account_id = ?", accountID).
		Select("id").
		Take(&id)

	return id > 0
}

// GetAccount retrieves a user account by provider and account ID.
func (s *userAccountStore) GetAccount(ctx context.Context, provider, accountID string) (*model.UserAccount, error) {
	var ret model.UserAccount
	err := s.DB(ctx).
		Model(&model.UserAccount{}).
		Where("provider = ?", provider).
		Where("account_id = ?", accountID).
		Take(&ret).
		Error

	return &ret, err
}

// FirstOrCreate finds first record matching the given conditions or creates a new one.
func (s *userAccountStore) FirstOrCreate(ctx context.Context, where model.UserAccount, account *model.UserAccount) error {
	return s.DB(ctx).Where(where).FirstOrCreate(account).Error
}

// FindByUID retrieves all accounts for a user by UID.
func (s *userAccountStore) FindByUID(ctx context.Context, uid string) ([]*model.UserAccount, error) {
	var ret []*model.UserAccount
	err := s.DB(ctx).Where("uid = ?", uid).Find(&ret).Error
	return ret, err
}

// FindByUIDAndProvider retrieves a user account by UID and provider.
func (s *userAccountStore) FindByUIDAndProvider(ctx context.Context, uid, provider string) (*model.UserAccount, error) {
	var ret model.UserAccount
	err := s.DB(ctx).Where("uid = ?", uid).Where("provider = ?", provider).Take(&ret).Error
	return &ret, err
}

// CountByUID counts the number of accounts for a user by UID.
func (s *userAccountStore) CountByUID(ctx context.Context, uid string) (int64, error) {
	var count int64
	err := s.DB(ctx).Model(&model.UserAccount{}).Where("uid = ?", uid).Count(&count).Error
	return count, err
}

// DeleteByID deletes a user account by ID.
func (s *userAccountStore) DeleteByID(ctx context.Context, id uint) error {
	return s.DB(ctx).Delete(&model.UserAccount{}, id).Error
}
