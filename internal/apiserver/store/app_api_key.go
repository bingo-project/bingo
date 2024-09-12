package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1"
	model "bingo/internal/apiserver/model"
)

type ApiKeyStore interface {
	List(ctx context.Context, req *v1.ListApiKeyRequest) (int64, []*model.ApiKey, error)
	Create(ctx context.Context, apiKey *model.ApiKey) error
	Get(ctx context.Context, ID uint) (*model.ApiKey, error)
	Update(ctx context.Context, apiKey *model.ApiKey, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, apiKeys []*model.ApiKey) error
	CreateIfNotExist(ctx context.Context, apiKey *model.ApiKey) error
	FirstOrCreate(ctx context.Context, where any, apiKey *model.ApiKey) error
	UpdateOrCreate(ctx context.Context, where any, apiKey *model.ApiKey) error
	Upsert(ctx context.Context, apiKey *model.ApiKey, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	GetByAK(ctx context.Context, ak string) (*model.ApiKey, error)
}

type apiKeys struct {
	db *gorm.DB
}

var _ ApiKeyStore = (*apiKeys)(nil)

func NewApiKeys(db *gorm.DB) *apiKeys {
	return &apiKeys{db: db}
}

func SearchApiKey(req *v1.ListApiKeyRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.UID != nil {
			db.Where("uid = ?", req.UID)
		}
		if req.AppID != nil {
			db.Where("app_id = ?", req.AppID)
		}
		if req.Name != nil {
			db.Where("name = ?", req.Name)
		}
		if req.AccessKey != nil {
			db.Where("access_key = ?", req.AccessKey)
		}
		if req.Status != nil {
			db.Where("status = ?", req.Status)
		}

		return db
	}
}

func (s *apiKeys) List(ctx context.Context, req *v1.ListApiKeyRequest) (count int64, ret []*model.ApiKey, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchApiKey(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *apiKeys) Create(ctx context.Context, apiKey *model.ApiKey) error {
	return s.db.WithContext(ctx).Create(&apiKey).Error
}

func (s *apiKeys) Get(ctx context.Context, ID uint) (apiKey *model.ApiKey, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&apiKey).Error

	return
}

func (s *apiKeys) Update(ctx context.Context, apiKey *model.ApiKey, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&apiKey).Error
}

func (s *apiKeys) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.ApiKey{}).Error
}

func (s *apiKeys) CreateInBatch(ctx context.Context, apiKeys []*model.ApiKey) error {
	return s.db.WithContext(ctx).CreateInBatches(&apiKeys, global.CreateBatchSize).Error
}

func (s *apiKeys) CreateIfNotExist(ctx context.Context, apiKey *model.ApiKey) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&apiKey).
		Error
}

func (s *apiKeys) FirstOrCreate(ctx context.Context, where any, apiKey *model.ApiKey) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&apiKey).
		FirstOrCreate(&apiKey).
		Error
}

func (s *apiKeys) UpdateOrCreate(ctx context.Context, where any, apiKey *model.ApiKey) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.ApiKey
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		apiKey.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&apiKey).Error
	})
}

func (s *apiKeys) Upsert(ctx context.Context, apiKey *model.ApiKey, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&apiKey).
		Error
}

func (s *apiKeys) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.ApiKey{}).
		Error
}

func (s *apiKeys) GetByAK(ctx context.Context, ak string) (apiKey *model.ApiKey, err error) {
	err = s.db.WithContext(ctx).Where("access_key = ?", ak).First(&apiKey).Error

	return
}
