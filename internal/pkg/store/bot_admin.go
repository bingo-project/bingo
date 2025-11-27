package store

import (
	"context"
	"slices"

	"github.com/bingo-project/component-base/util/gormutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"bingo/internal/pkg/global"
	model "bingo/internal/pkg/model/bot"
	v1 "bingo/pkg/api/apiserver/v1/bot"
	genericstore "bingo/pkg/store"
	"bingo/pkg/store/where"
)

// BotAdminStore 定义了 Bot Admin 相关操作的接口.
type BotAdminStore interface {
	Create(ctx context.Context, obj *model.Admin) error
	Update(ctx context.Context, obj *model.Admin, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.Admin, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.Admin, error)

	BotAdminExpansion
}

// BotAdminExpansion 定义了 Bot Admin 操作的扩展方法.
type BotAdminExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.Admin, error)
	CreateInBatch(ctx context.Context, admins []*model.Admin) error
	CreateIfNotExist(ctx context.Context, admin *model.Admin) error
	FirstOrCreate(ctx context.Context, where any, admin *model.Admin) error
	UpdateOrCreate(ctx context.Context, where any, admin *model.Admin) error
	Upsert(ctx context.Context, admin *model.Admin, fields ...string) error
	DeleteInBatch(ctx context.Context, ids []uint) error

	GetByUserID(ctx context.Context, userID string) (*model.Admin, error)
	IsAdmin(ctx context.Context, userID string) (bool, error)
}

type botAdminStore struct {
	*genericstore.Store[model.Admin]
}

var _ BotAdminStore = (*botAdminStore)(nil)

func NewBotAdminStore(store *datastore) *botAdminStore {
	return &botAdminStore{
		Store: genericstore.NewStore[model.Admin](store, NewLogger()),
	}
}

// ListWithRequest 根据请求参数列表查询.
func (s *botAdminStore) ListWithRequest(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.Admin, error) {
	// 构建查询条件
	opts := where.NewWhere()

	if req.Source != nil {
		opts = opts.F("source", *req.Source)
	}
	if req.UserID != nil {
		opts = opts.F("user_id", *req.UserID)
	}

	// 处理分页
	db := s.DB(ctx, opts)
	var ret []*model.Admin
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// CreateInBatch 批量创建.
func (s *botAdminStore) CreateInBatch(ctx context.Context, admins []*model.Admin) error {
	return s.DB(ctx).CreateInBatches(admins, global.CreateBatchSize).Error
}

// CreateIfNotExist 如果不存在则创建.
func (s *botAdminStore) CreateIfNotExist(ctx context.Context, admin *model.Admin) error {
	return s.DB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(admin).
		Error
}

// FirstOrCreate 首先查找，不存在则创建.
func (s *botAdminStore) FirstOrCreate(ctx context.Context, where any, admin *model.Admin) error {
	return s.DB(ctx).
		Where(where).
		Attrs(admin).
		FirstOrCreate(admin).
		Error
}

// UpdateOrCreate 更新或创建.
func (s *botAdminStore) UpdateOrCreate(ctx context.Context, where any, admin *model.Admin) error {
	return s.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.Admin
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(where).
			First(&exist).
			Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		admin.ID = exist.ID
		return tx.Omit("CreatedAt").Save(admin).Error
	})
}

// Upsert 创建或更新.
func (s *botAdminStore) Upsert(ctx context.Context, admin *model.Admin, fields ...string) error {
	do := clause.OnConflict{UpdateAll: true}
	if len(fields) > 0 {
		do.UpdateAll = false
		do.DoUpdates = clause.AssignmentColumns(fields)
	}

	return s.DB(ctx).Clauses(do).Create(admin).Error
}

// DeleteInBatch 批量删除.
func (s *botAdminStore) DeleteInBatch(ctx context.Context, ids []uint) error {
	return s.DB(ctx).
		Where("id IN (?)", ids).
		Delete(&model.Admin{}).
		Error
}

// GetByUserID 根据 user_id 获取.
func (s *botAdminStore) GetByUserID(ctx context.Context, userID string) (*model.Admin, error) {
	return s.Get(ctx, where.F("user_id", userID))
}

// IsAdmin 检查用户是否是管理员.
func (s *botAdminStore) IsAdmin(ctx context.Context, userID string) (bool, error) {
	admin, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return false, err
	}

	roles := []model.Role{model.RoleRoot, model.RoleAdmin}
	return slices.Contains(roles, admin.Role), nil
}
