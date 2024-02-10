package syscfg

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1/syscfg"
	model "bingo/internal/apiserver/model/syscfg"
)

type AppStore interface {
	List(ctx context.Context, req *v1.ListAppRequest) (int64, []*model.App, error)
	Create(ctx context.Context, app *model.App) error
	Get(ctx context.Context, ID uint) (*model.App, error)
	Update(ctx context.Context, app *model.App, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, apps []*model.App) error
	CreateIfNotExist(ctx context.Context, app *model.App) error
	FirstOrCreate(ctx context.Context, where any, app *model.App) error
	UpdateOrCreate(ctx context.Context, where any, app *model.App) error
	Upsert(ctx context.Context, app *model.App, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error
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
		if req.Name != nil {
			db.Where("name = ?", req.Name)
		}
		if req.Version != nil {
			db.Where("version = ?", req.Version)
		}
		if req.Description != nil {
			db.Where("description = ?", req.Description)
		}
		if req.AboutUs != nil {
			db.Where("about_us = ?", req.AboutUs)
		}
		if req.Logo != nil {
			db.Where("logo = ?", req.Logo)
		}
		if req.Enabled != nil {
			db.Where("enabled = ?", req.Enabled)
		}

		return db
	}
}

func (s *apps) List(ctx context.Context, req *v1.ListAppRequest) (count int64, ret []*model.App, err error) {
	db := s.db.Scopes(SearchApp(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *apps) Create(ctx context.Context, app *model.App) error {
	return s.db.Create(&app).Error
}

func (s *apps) Get(ctx context.Context, ID uint) (app *model.App, err error) {
	err = s.db.Where("id = ?", ID).First(&app).Error

	return
}

func (s *apps) Update(ctx context.Context, app *model.App, fields ...string) error {
	return s.db.Select(fields).Save(&app).Error
}

func (s *apps) Delete(ctx context.Context, ID uint) error {
	return s.db.Where("id = ?", ID).Delete(&model.App{}).Error
}

func (s *apps) CreateInBatch(ctx context.Context, apps []*model.App) error {
	return s.db.CreateInBatches(&apps, global.CreateBatchSize).Error
}

func (s *apps) CreateIfNotExist(ctx context.Context, app *model.App) error {
	return s.db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&app).
		Error
}

func (s *apps) FirstOrCreate(ctx context.Context, where any, app *model.App) error {
	return s.db.Where(where).
		Attrs(&app).
		FirstOrCreate(&app).
		Error
}

func (s *apps) UpdateOrCreate(ctx context.Context, where any, app *model.App) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
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

	return s.db.Clauses(do).
		Create(&app).
		Error
}

func (s *apps) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.Where("id IN (?)", ids).
		Delete(&model.App{}).
		Error
}
