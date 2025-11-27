package store

import (
	"context"
	"slices"

	"github.com/bingo-project/component-base/util/gormutil"

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
