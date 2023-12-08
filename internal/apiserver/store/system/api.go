package system

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/util/helper"
	v1 "bingo/pkg/api/bingo/v1"
)

type ApiStore interface {
	List(ctx context.Context, req *v1.ListApiRequest) (int64, []*model.ApiM, error)
	Create(ctx context.Context, api *model.ApiM) error
	Get(ctx context.Context, ID uint) (*model.ApiM, error)
	Update(ctx context.Context, api *model.ApiM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, apis []*model.ApiM) error
	FirstOrCreate(ctx context.Context, where any, api *model.ApiM) error
	UpdateOrCreate(ctx context.Context, where any, api *model.ApiM) error

	All(ctx context.Context) ([]*model.ApiM, error)
	GetByIDs(ctx context.Context, IDs []uint) (ret []*model.ApiM, err error)
	GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error)
}

type apis struct {
	db *gorm.DB
}

var _ ApiStore = (*apis)(nil)

func NewApis(db *gorm.DB) *apis {
	return &apis{db: db}
}

func (s *apis) List(ctx context.Context, req *v1.ListApiRequest) (count int64, ret []*model.ApiM, err error) {
	// Order
	if req.Order == "" {
		req.Order = "id"
	}

	// Sort
	if req.Sort == "" {
		req.Sort = "desc"
	}

	err = s.db.Offset(req.Offset).
		Limit(helper.DefaultLimit(req.Limit)).
		Order(req.Order + " " + req.Sort).
		Find(&ret).
		Offset(-1).
		Limit(-1).
		Count(&count).
		Error

	return
}

func (s *apis) Create(ctx context.Context, api *model.ApiM) error {
	return s.db.Create(&api).Error
}

func (s *apis) Get(ctx context.Context, ID uint) (api *model.ApiM, err error) {
	err = s.db.Where("id = ?", ID).First(&api).Error

	return
}

func (s *apis) Update(ctx context.Context, api *model.ApiM, fields ...string) error {
	return s.db.Select(fields).Save(&api).Error
}

func (s *apis) Delete(ctx context.Context, ID uint) error {
	return s.db.Where("id = ?", ID).Delete(&model.ApiM{}).Error
}

func (s *apis) CreateInBatch(ctx context.Context, apis []*model.ApiM) error {
	return s.db.CreateInBatches(&apis, global.CreateBatchSize).Error
}

func (s *apis) FirstOrCreate(ctx context.Context, where any, api *model.ApiM) error {
	return s.db.Where(where).
		Attrs(&api).
		FirstOrCreate(&api).
		Error
}

func (s *apis) UpdateOrCreate(ctx context.Context, where any, api *model.ApiM) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var exist model.ApiM
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		api.ID = exist.ID

		return tx.Save(&api).Error
	})
}

func (s *apis) All(ctx context.Context) (ret []*model.ApiM, err error) {
	err = s.db.Find(&ret).Error

	return
}

func (s *apis) GetByIDs(ctx context.Context, IDs []uint) (ret []*model.ApiM, err error) {
	err = s.db.Where("id IN ?", IDs).Find(&ret).Error

	return
}

func (s *apis) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error) {
	err = s.db.Model(&model.ApiM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return
}
