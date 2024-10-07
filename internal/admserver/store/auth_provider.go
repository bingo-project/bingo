package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
)

type AuthProviderStore interface {
	List(ctx context.Context, req *v1.ListAuthProviderRequest) (int64, []*model.AuthProvider, error)
	Create(ctx context.Context, authProvider *model.AuthProvider) error
	Get(ctx context.Context, ID uint) (*model.AuthProvider, error)
	Update(ctx context.Context, authProvider *model.AuthProvider, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, authProviders []*model.AuthProvider) error
	CreateIfNotExist(ctx context.Context, authProvider *model.AuthProvider) error
	FirstOrCreate(ctx context.Context, where any, authProvider *model.AuthProvider) error
	UpdateOrCreate(ctx context.Context, where any, authProvider *model.AuthProvider) error
	Upsert(ctx context.Context, authProvider *model.AuthProvider, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	FindEnabled(ctx context.Context) (ret []*model.AuthProvider, err error)
	FirstEnabled(ctx context.Context, provider string) (authProvider *model.AuthProvider, err error)
}

type authProviders struct {
	db *gorm.DB
}

var _ AuthProviderStore = (*authProviders)(nil)

func NewAuthProviders(db *gorm.DB) *authProviders {
	return &authProviders{db: db}
}

func SearchAuthProvider(req *v1.ListAuthProviderRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Name != nil {
			db.Where("name = ?", req.Name)
		}
		if req.Status != nil {
			db.Where("status = ?", req.Status)
		}
		if req.IsDefault != nil {
			db.Where("is_default = ?", req.IsDefault)
		}

		return db
	}
}

func (s *authProviders) List(ctx context.Context, req *v1.ListAuthProviderRequest) (count int64, ret []*model.AuthProvider, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchAuthProvider(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *authProviders) Create(ctx context.Context, authProvider *model.AuthProvider) error {
	return s.db.WithContext(ctx).Create(&authProvider).Error
}

func (s *authProviders) Get(ctx context.Context, ID uint) (authProvider *model.AuthProvider, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&authProvider).Error

	return
}

func (s *authProviders) Update(ctx context.Context, authProvider *model.AuthProvider, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&authProvider).Error
}

func (s *authProviders) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.AuthProvider{}).Error
}

func (s *authProviders) CreateInBatch(ctx context.Context, authProviders []*model.AuthProvider) error {
	return s.db.WithContext(ctx).CreateInBatches(&authProviders, global.CreateBatchSize).Error
}

func (s *authProviders) CreateIfNotExist(ctx context.Context, authProvider *model.AuthProvider) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&authProvider).
		Error
}

func (s *authProviders) FirstOrCreate(ctx context.Context, where any, authProvider *model.AuthProvider) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&authProvider).
		FirstOrCreate(&authProvider).
		Error
}

func (s *authProviders) UpdateOrCreate(ctx context.Context, where any, authProvider *model.AuthProvider) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.AuthProvider
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		authProvider.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&authProvider).Error
	})
}

func (s *authProviders) Upsert(ctx context.Context, authProvider *model.AuthProvider, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&authProvider).
		Error
}

func (s *authProviders) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.AuthProvider{}).
		Error
}

func (s *authProviders) FindEnabled(ctx context.Context) (ret []*model.AuthProvider, err error) {
	err = s.db.WithContext(ctx).
		Where(&model.AuthProvider{Status: model.AuthProviderStatusEnabled}).
		Find(&ret).
		Error

	return
}

func (s *authProviders) FirstEnabled(ctx context.Context, provider string) (authProvider *model.AuthProvider, err error) {
	err = s.db.WithContext(ctx).
		Where("name = ?", provider).
		Where("status = ?", model.AuthProviderStatusEnabled).
		First(&authProvider).
		Error

	return
}
