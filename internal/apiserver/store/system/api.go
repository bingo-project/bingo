package system

import (
	"context"
	"errors"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/model"
)

type ApiStore interface {
	List(ctx context.Context, req *v1.ListApiRequest) (int64, []*model.ApiM, error)
	Create(ctx context.Context, api *model.ApiM) error
	Get(ctx context.Context, ID uint) (*model.ApiM, error)
	Update(ctx context.Context, api *model.ApiM, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, apis []*model.ApiM) error
	CreateIfNotExist(ctx context.Context, api *model.ApiM) error
	FirstOrCreate(ctx context.Context, where any, api *model.ApiM) error
	UpdateOrCreate(ctx context.Context, where any, api *model.ApiM) error
	Upsert(ctx context.Context, api *model.ApiM, fields ...string) error

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

func SearchApi(req *v1.ListApiRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Method != "" {
			db.Where("method = ?", req.Method)
		}
		if req.Path != "" {
			db.Where("path like ?", "%"+req.Path+"%")
		}
		if req.Group != "" {
			db.Where("`group` = ?", req.Group)
		}

		return db
	}
}

func (s *apis) List(ctx context.Context, req *v1.ListApiRequest) (count int64, ret []*model.ApiM, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchApi(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *apis) Create(ctx context.Context, api *model.ApiM) error {
	return s.db.WithContext(ctx).Create(&api).Error
}

func (s *apis) Get(ctx context.Context, ID uint) (api *model.ApiM, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&api).Error

	return
}

func (s *apis) Update(ctx context.Context, api *model.ApiM, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&api).Error
}

func (s *apis) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.ApiM{}).Error
}

func (s *apis) CreateInBatch(ctx context.Context, apis []*model.ApiM) error {
	return s.db.WithContext(ctx).CreateInBatches(&apis, global.CreateBatchSize).Error
}

func (s *apis) CreateIfNotExist(ctx context.Context, api *model.ApiM) error {
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&api).
		Error
}

func (s *apis) FirstOrCreate(ctx context.Context, where any, api *model.ApiM) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&api).
		FirstOrCreate(&api).
		Error
}

func (s *apis) UpdateOrCreate(ctx context.Context, where any, api *model.ApiM) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.ApiM
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		api.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&api).Error
	})
}

func (s *apis) Upsert(ctx context.Context, api *model.ApiM, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&api).
		Error
}

func (s *apis) All(ctx context.Context) (ret []*model.ApiM, err error) {
	err = s.db.WithContext(ctx).Find(&ret).Error

	return
}

func (s *apis) GetByIDs(ctx context.Context, IDs []uint) (ret []*model.ApiM, err error) {
	err = s.db.WithContext(ctx).Where("id IN ?", IDs).Find(&ret).Error

	return
}

func (s *apis) GetIDsByPathAndMethod(ctx context.Context, pathAndMethod [][]string) (ret []uint, err error) {
	err = s.db.WithContext(ctx).
		Model(&model.ApiM{}).
		Select("id").
		Where("(path, method) IN ?", pathAndMethod).
		Find(&ret).
		Error

	return
}
