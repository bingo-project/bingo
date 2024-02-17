package bot

import (
	"context"
	"errors"
	"slices"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1/bot"
	model "bingo/internal/apiserver/model/bot"
)

type AdminStore interface {
	List(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.Admin, error)
	Create(ctx context.Context, admin *model.Admin) error
	Get(ctx context.Context, ID uint) (*model.Admin, error)
	Update(ctx context.Context, admin *model.Admin, fields ...string) error
	Delete(ctx context.Context, ID uint) error

	CreateInBatch(ctx context.Context, admins []*model.Admin) error
	CreateIfNotExist(ctx context.Context, admin *model.Admin) error
	FirstOrCreate(ctx context.Context, where any, admin *model.Admin) error
	UpdateOrCreate(ctx context.Context, where any, admin *model.Admin) error
	Upsert(ctx context.Context, admin *model.Admin, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	GetByUserID(ctx context.Context, userID string) (admin *model.Admin, err error)
	IsAdmin(ctx context.Context, userID string) (ret bool, err error)
}

type admins struct {
	db *gorm.DB
}

var _ AdminStore = (*admins)(nil)

func NewAdmins(db *gorm.DB) *admins {
	return &admins{db: db}
}

func SearchAdmin(req *v1.ListAdminRequest) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if req.Source != nil {
			db.Where("source = ?", req.Source)
		}
		if req.UserID != nil {
			db.Where("user_id = ?", req.UserID)
		}

		return db
	}
}

func (s *admins) List(ctx context.Context, req *v1.ListAdminRequest) (count int64, ret []*model.Admin, err error) {
	db := s.db.WithContext(ctx).Scopes(SearchAdmin(req))
	count, err = gormutil.Paginate(db, &req.ListOptions, &ret)

	return
}

func (s *admins) Create(ctx context.Context, admin *model.Admin) error {
	return s.db.WithContext(ctx).Create(&admin).Error
}

func (s *admins) Get(ctx context.Context, ID uint) (admin *model.Admin, err error) {
	err = s.db.WithContext(ctx).Where("id = ?", ID).First(&admin).Error

	return
}

func (s *admins) Update(ctx context.Context, admin *model.Admin, fields ...string) error {
	return s.db.WithContext(ctx).Select(fields).Save(&admin).Error
}

func (s *admins) Delete(ctx context.Context, ID uint) error {
	return s.db.WithContext(ctx).Where("id = ?", ID).Delete(&model.Admin{}).Error
}

func (s *admins) CreateInBatch(ctx context.Context, admins []*model.Admin) error {
	return s.db.WithContext(ctx).CreateInBatches(&admins, global.CreateBatchSize).Error
}

func (s *admins) CreateIfNotExist(ctx context.Context, admin *model.Admin) error {
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&admin).
		Error
}

func (s *admins) FirstOrCreate(ctx context.Context, where any, admin *model.Admin) error {
	return s.db.WithContext(ctx).
		Where(where).
		Attrs(&admin).
		FirstOrCreate(&admin).
		Error
}

func (s *admins) UpdateOrCreate(ctx context.Context, where any, admin *model.Admin) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var exist model.Admin
		err := tx.WithContext(ctx).
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		admin.ID = exist.ID

		return tx.Omit("CreatedAt").Save(&admin).Error
	})
}

func (s *admins) Upsert(ctx context.Context, admin *model.Admin, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.db.WithContext(ctx).
		Clauses(do).
		Create(&admin).
		Error
}

func (s *admins) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.db.WithContext(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Admin{}).
		Error
}

func (s *admins) GetByUserID(ctx context.Context, userID string) (admin *model.Admin, err error) {
	err = s.db.WithContext(ctx).Where("user_id = ?", userID).First(&admin).Error

	return
}

func (s *admins) IsAdmin(ctx context.Context, userID string) (ret bool, err error) {
	admin, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return ret, err
	}

	roles := []model.Role{model.RoleRoot, model.RoleAdmin}
	if !slices.Contains(roles, admin.Role) {
		return
	}

	return true, nil
}
