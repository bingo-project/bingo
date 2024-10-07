package store

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
	"bingo/internal/pkg/model"
	"bingo/pkg/api/apiserver/v1"
)

type AppStore interface {
	List(ctx context.Context, req *v1.ListAppRequest) (int64, []*model.App, error)
	Create(ctx context.Context, app *model.App) error
	Get(ctx context.Context, appID string) (*model.App, error)
	Update(ctx context.Context, app *model.App, fields ...string) error
	Delete(ctx context.Context, appID string) error

	CreateInBatch(ctx context.Context, apps []*model.App) error
	CreateIfNotExist(ctx context.Context, app *model.App) error
	FirstOrCreate(ctx context.Context, where any, app *model.App) error
	UpdateOrCreate(ctx context.Context, where any, app *model.App) error
	Upsert(ctx context.Context, app *model.App, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	GetByAppID(ctx context.Context, appID string) (*model.App, error)
}

type apps struct {
	db *gorm.DB
}

var _ AppStore = (*apps)(nil)

func NewApps(db *gorm.DB) *apps {
	return &apps{db: db}
}

func SearchApp(req *v1.ListAppRequest) func(db *gorm.DB) *gorm.DB {
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
		if req.Status != nil {
			db.Where("status = ?", req.Status)
		}
		if req.Description != nil {
			db.Where("description = ?", req.Description)
		}
		if req.Logo != nil {
			db.Where("logo = ?", req.Logo)
		}

		return db
	}
}

func (s *apps) List(ctx context.Context, req *v1.ListAppRequest) (count int64, ret []*model.App, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchApp(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *apps) Create(ctx context.Context, app *model.App) error {
	return s.db.WithContext(ctx).Create(&app).Error
}

func (s *apps) Get(ctx context.Context, appID string) (app *model.App, err error) {
	err = s.db.WithContext(ctx).Where("app_id = ?", appID).First(&app).Error

	return
}

func (s *apps) Update(ctx context.Context, app *model.App, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&app).Error
}

func (s *apps) Delete(ctx context.Context, appID string) error {
	return s.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&model.App{}).Error
}

func (s *apps) CreateInBatch(ctx context.Context, apps []*model.App) error {
	return s.db.WithContext(ctx).CreateInBatches(&apps, global.CreateBatchSize).Error
}

func (s *apps) CreateIfNotExist(ctx context.Context, app *model.App) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&app).
		Error
}

func (s *apps) FirstOrCreate(ctx context.Context, where any, app *model.App) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&app).
		FirstOrCreate(&app).
		Error
}

func (s *apps) UpdateOrCreate(ctx context.Context, where any, app *model.App) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.App
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		app.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&app).Error
	})
}

func (s *apps) Upsert(ctx context.Context, app *model.App, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&app).
		Error
}

func (s *apps) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.App{}).
		Error
}

func (s *apps) GetByAppID(ctx context.Context, appID string) (ret *model.App, err error) {
	err = s.db.WithContext(ctx).Where("app_id = ?", appID).First(&ret).Error

	return
}
