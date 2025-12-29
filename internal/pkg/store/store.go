// ABOUTME: Central data store interface and implementation.
// ABOUTME: Provides unified access to all domain-specific stores.

package store

import (
	"context"
	"sync"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/pkg/store/where"
)

//go:generate mockgen -destination mock_store.go -package store bingo/internal/pkg/store IStore

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
	// AppVersion returns the app version store.
	AppVersion() AppVersionStore
	// SysRole returns the system role store.
	SysRole() SysRoleStore
	// SysApi returns the system API store.
	SysApi() SysApiStore
	// SysMenu returns the system menu store.
	SysMenu() SysMenuStore
	// SysRoleMenu returns the system role-menu store.
	SysRoleMenu() SysRoleMenuStore

	// Bot returns the bot store.
	Bot() BotStore
	// BotChannel returns the bot channel store.
	BotChannel() BotChannelStore
	// BotAdmin returns the bot admin store.
	BotAdmin() BotAdminStore

	// User returns the user store.
	User() UserStore
	// UserAccount returns the user account store.
	UserAccount() UserAccountStore
	// AuthProvider returns the auth provider store.
	AuthProvider() AuthProviderStore

	// App returns the app store.
	App() AppStore
	// ApiKey returns the API key store.
	ApiKey() ApiKeyStore

	// NtfMessage returns the notification message store.
	NtfMessage() NtfMessageStore
	// NtfAnnouncement returns the notification announcement store.
	NtfAnnouncement() NtfAnnouncementStore
	// NtfPreference returns the notification preference store.
	NtfPreference() NtfPreferenceStore

	// AiProvider returns the AI provider store.
	AiProvider() AiProviderStore
	// AiModel returns the AI model store.
	AiModel() AiModelStore
	// AiQuotaTier returns the AI quota tier store.
	AiQuotaTier() AiQuotaTierStore
	// AiUserQuota returns the AI user quota store.
	AiUserQuota() AiUserQuotaStore
	// AiSession returns the AI session store.
	AiSession() AiSessionStore
	// AiMessage returns the AI message store.
	AiMessage() AiMessageStore
}

// transactionKey used for context.
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

// UserAccount returns the user account store.
func (ds *datastore) UserAccount() UserAccountStore {
	return NewUserAccountStore(ds)
}

// AuthProvider returns the auth provider store.
func (ds *datastore) AuthProvider() AuthProviderStore {
	return NewAuthProviderStore(ds)
}

// App returns the app store.
func (ds *datastore) App() AppStore {
	return NewAppStore(ds)
}

// ApiKey returns the API key store.
func (ds *datastore) ApiKey() ApiKeyStore {
	return NewApiKeyStore(ds)
}

// AppVersion returns the app version store.
func (ds *datastore) AppVersion() AppVersionStore {
	return NewAppVersionStore(ds)
}

// SysRole returns the system role store.
func (ds *datastore) SysRole() SysRoleStore {
	return NewSysRoleStore(ds)
}

// SysApi returns the system API store.
func (ds *datastore) SysApi() SysApiStore {
	return NewSysApiStore(ds)
}

// SysMenu returns the system menu store.
func (ds *datastore) SysMenu() SysMenuStore {
	return NewSysMenuStore(ds)
}

// SysRoleMenu returns the system role-menu store.
func (ds *datastore) SysRoleMenu() SysRoleMenuStore {
	return NewSysRoleMenuStore(ds)
}

// NtfMessage returns the notification message store.
func (ds *datastore) NtfMessage() NtfMessageStore {
	return NewNtfMessageStore(ds)
}

// NtfAnnouncement returns the notification announcement store.
func (ds *datastore) NtfAnnouncement() NtfAnnouncementStore {
	return NewNtfAnnouncementStore(ds)
}

// NtfPreference returns the notification preference store.
func (ds *datastore) NtfPreference() NtfPreferenceStore {
	return NewNtfPreferenceStore(ds)
}

// AiProvider returns the AI provider store.
func (ds *datastore) AiProvider() AiProviderStore {
	return NewAiProviderStore(ds)
}

// AiModel returns the AI model store.
func (ds *datastore) AiModel() AiModelStore {
	return NewAiModelStore(ds)
}

// AiQuotaTier returns the AI quota tier store.
func (ds *datastore) AiQuotaTier() AiQuotaTierStore {
	return NewAiQuotaTierStore(ds)
}

// AiUserQuota returns the AI user quota store.
func (ds *datastore) AiUserQuota() AiUserQuotaStore {
	return NewAiUserQuotaStore(ds)
}

// AiSession returns the AI session store.
func (ds *datastore) AiSession() AiSessionStore {
	return NewAiSessionStore(ds)
}

// AiMessage returns the AI message store.
func (ds *datastore) AiMessage() AiMessageStore {
	return NewAiMessageStore(ds)
}
