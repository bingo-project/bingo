package syscfg

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1/syscfg"
	model "bingo/internal/pkg/model/syscfg"
)

type AppVersionStore interface {
	List(ctx context.Context, req *v1.ListAppVersionRequest) (int64, []*model.AppVersion, error)
	Create(ctx context.Context, app *model.AppVersion) error
	Get(ctx context.Context, ID uint) (*model.AppVersion, error)
	Update(ctx context.Context, app *model.AppVersion, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, apps []*model.AppVersion) error
	CreateIfNotExist(ctx context.Context, app *model.AppVersion) error
	FirstOrCreate(ctx context.Context, where any, app *model.AppVersion) error
	UpdateOrCreate(ctx context.Context, where any, app *model.AppVersion) error
	Upsert(ctx context.Context, app *model.AppVersion, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error
}

type appVersions struct {
	db *gorm.DB
}

var _ AppVersionStore = (*appVersions)(nil)

func NewAppVersions(db *gorm.DB) *appVersions {
	return &appVersions{db: db}
}

func SearchApp(req *v1.ListAppVersionRequest) func(db *gorm.DB) *gorm.DB {
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

func (s *appVersions) List(ctx context.Context, req *v1.ListAppVersionRequest) (count int64, ret []*model.AppVersion, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchApp(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *appVersions) Create(ctx context.Context, app *model.AppVersion) error {
	return s.db.WithContext(ctx).Create(&app).Error
}

func (s *appVersions) Get(ctx context.Context, ID uint) (app *model.AppVersion, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&app).Error

	return
}

func (s *appVersions) Update(ctx context.Context, app *model.AppVersion, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&app).Error
}

func (s *appVersions) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.AppVersion{}).Error
}

func (s *appVersions) CreateInBatch(ctx context.Context, apps []*model.AppVersion) error {
	return s.db.WithContext(ctx).CreateInBatches(&apps, global.CreateBatchSize).Error
}

func (s *appVersions) CreateIfNotExist(ctx context.Context, app *model.AppVersion) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&app).
		Error
}

func (s *appVersions) FirstOrCreate(ctx context.Context, where any, app *model.AppVersion) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&app).
		FirstOrCreate(&app).
		Error
}

func (s *appVersions) UpdateOrCreate(ctx context.Context, where any, app *model.AppVersion) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.AppVersion
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

func (s *appVersions) Upsert(ctx context.Context, app *model.AppVersion, fields ...string) error {
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

func (s *appVersions) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.AppVersion{}).
		Error
}
