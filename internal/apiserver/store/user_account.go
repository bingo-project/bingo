package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/global"
	"bingo/internal/pkg/model"
)

type UserAccountStore interface {
	List(ctx context.Context, req *v1.ListUserAccountRequest) (int64, []*model.UserAccount, error)
	Create(ctx context.Context, userAccount *model.UserAccount) error
	Get(ctx context.Context, ID uint) (*model.UserAccount, error)
	Update(ctx context.Context, userAccount *model.UserAccount, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, userAccounts []*model.UserAccount) error
	CreateIfNotExist(ctx context.Context, userAccount *model.UserAccount) error
	FirstOrCreate(ctx context.Context, where any, userAccount *model.UserAccount) error
	UpdateOrCreate(ctx context.Context, where any, userAccount *model.UserAccount) error
	Upsert(ctx context.Context, userAccount *model.UserAccount, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	CheckExist(ctx context.Context, provider, accountID string) bool
	GetAccount(ctx context.Context, provider, accountID string) (ret *model.UserAccount, err error)
}

type userAccounts struct {
	db *gorm.DB
}

var _ UserAccountStore = (*userAccounts)(nil)

func NewUserAccounts(db *gorm.DB) *userAccounts {
	return &userAccounts{db: db}
}

func SearchUserAccount(req *v1.ListUserAccountRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.UID != nil {
			db.Where("uid = ?", req.UID)
		}
		if req.Provider != nil {
			db.Where("provider = ?", req.Provider)
		}
		if req.AccountID != nil {
			db.Where("account_id = ?", req.AccountID)
		}
		if req.Username != nil {
			db.Where("username = ?", req.Username)
		}
		if req.Nickname != nil {
			db.Where("nickname = ?", req.Nickname)
		}
		if req.Email != nil {
			db.Where("email = ?", req.Email)
		}
		if req.Bio != nil {
			db.Where("bio = ?", req.Bio)
		}
		if req.Avatar != nil {
			db.Where("avatar = ?", req.Avatar)
		}

		return db
	}
}

func (s *userAccounts) List(ctx context.Context, req *v1.ListUserAccountRequest) (count int64, ret []*model.UserAccount, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchUserAccount(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *userAccounts) Create(ctx context.Context, userAccount *model.UserAccount) error {
	return s.db.WithContext(ctx).Create(&userAccount).Error
}

func (s *userAccounts) Get(ctx context.Context, ID uint) (userAccount *model.UserAccount, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&userAccount).Error

	return
}

func (s *userAccounts) Update(ctx context.Context, userAccount *model.UserAccount, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&userAccount).Error
}

func (s *userAccounts) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.UserAccount{}).Error
}

func (s *userAccounts) CreateInBatch(ctx context.Context, userAccounts []*model.UserAccount) error {
	return s.db.WithContext(ctx).CreateInBatches(&userAccounts, global.CreateBatchSize).Error
}

func (s *userAccounts) CreateIfNotExist(ctx context.Context, userAccount *model.UserAccount) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&userAccount).
		Error
}

func (s *userAccounts) FirstOrCreate(ctx context.Context, where any, userAccount *model.UserAccount) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&userAccount).
		FirstOrCreate(&userAccount).
		Error
}

func (s *userAccounts) UpdateOrCreate(ctx context.Context, where any, userAccount *model.UserAccount) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.UserAccount
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		userAccount.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&userAccount).Error
	})
}

func (s *userAccounts) Upsert(ctx context.Context, userAccount *model.UserAccount, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&userAccount).
		Error
}

func (s *userAccounts) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.UserAccount{}).
		Error
}

func (s *userAccounts) CheckExist(ctx context.Context, provider, accountID string) bool {
	var id int64
	s.db.WithContext(ctx).
		Model(&model.UserAccount{}).
		Where("provider = ?", provider).
		Where("account_id = ?", accountID).
		Select("id").
		Take(&id)

	return id > 0
}

func (s *userAccounts) GetAccount(ctx context.Context, provider, accountID string) (ret *model.UserAccount, err error) {
	err = s.db.WithContext(ctx).
		Model(&model.UserAccount{}).
		Where("provider = ?", provider).
		Where("account_id = ?", accountID).
		Take(&ret).
		Error

	return
}
