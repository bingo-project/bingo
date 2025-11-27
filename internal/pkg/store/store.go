package store

import (
	"context"
	"sync"

	"gorm.io/gorm"

	"bingo/pkg/store/where"
)

//go:generate mockgen -destination mock_store.go -package store stx/internal/apiserver/store IStore

var (
	once sync.Once
	S    *datastore
)

// IStore defines the interface for the shared store layer.
type IStore interface {
	// DB returns a database instance.
	DB(ctx context.Context, wheres ...where.Where) *gorm.DB
	// TX executes a function in a transaction context.
	TX(ctx context.Context, fn func(ctx context.Context) error) error

	// Admin returns the system admin store.
	Admin() AdminStore
	// Schedule returns the system schedule store.
	Schedule() ScheduleStore
	// SysConfig returns the system config store.
	SysConfig() ConfigStore

	// Bot returns the bot store.
	Bot() BotStore
	// BotChannel returns the bot channel store.
	BotChannel() BotChannelStore
	// BotAdmin returns the bot admin store.
	BotAdmin() BotAdminStore

	// User returns the user store.
	User() UserStore
}

// transactionKey used for context
type transactionKey struct{}

type datastore struct {
	core *gorm.DB

	// 可以根据需要添加其他数据库实例
	// fake *gorm.DB
}

var _ IStore = (*datastore)(nil)

func NewStore(db *gorm.DB) *datastore {
	once.Do(func() {
		S = &datastore{core: db}
	})
	return S
}

// DB 根据传入的条件（wheres）对数据库实例进行筛选.
// 如果未传入任何条件，则返回上下文中的数据库实例（事务实例或核心数据库实例）.
func (ds *datastore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	db := ds.core
	// 从上下文中提取事务实例
	if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
		db = tx
	}

	// 遍历所有传入的条件并逐一叠加到数据库查询对象上
	for _, whr := range wheres {
		db = whr.Where(db)
	}
	return db
}

// TX 返回一个新的事务实例.
// nolint: fatcontext
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return ds.core.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			ctx = context.WithValue(ctx, transactionKey{}, tx)
			return fn(ctx)
		},
	)
}

// Admin returns the system admin store.
func (ds *datastore) Admin() AdminStore {
	return NewAdminStore(ds)
}

// Schedule returns the system schedule store.
func (ds *datastore) Schedule() ScheduleStore {
	return NewScheduleStore(ds)
}

// SysConfig returns the system config store.
func (ds *datastore) SysConfig() ConfigStore {
	return NewConfigStore(ds)
}

// Bot returns the bot store.
func (ds *datastore) Bot() BotStore {
	return NewBotStore(ds)
}

// BotChannel returns the bot channel store.
func (ds *datastore) BotChannel() BotChannelStore {
	return NewBotChannelStore(ds)
}

// BotAdmin returns the bot admin store.
func (ds *datastore) BotAdmin() BotAdminStore {
	return NewBotAdminStore(ds)
}

// User returns the user store.
func (ds *datastore) User() UserStore {
	return NewUserStore(ds)
}
