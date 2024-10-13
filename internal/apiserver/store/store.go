package store

import (
	"sync"

	"gorm.io/gorm"

	"bingo/internal/apiserver/store/syscfg"
)

//go:generate mockgen -destination mock_store.go -package store bingo/internal/apiserver/store IStore

var (
	once sync.Once
	S    *datastore // 全局变量，方便其它包直接调用已初始化好的 S 实例.
)

// IStore 定义了 Store 层需要实现的方法.
type IStore interface {
	DB() *gorm.DB

	Users() UserStore
	UserAccounts() UserAccountStore
	AuthProviders() AuthProviderStore

	AppVersions() syscfg.AppVersionStore
	Configs() syscfg.ConfigStore

	Apps() AppStore
	ApiKeys() ApiKeyStore
}

// datastore 是 IStore 的一个具体实现.
type datastore struct {
	db *gorm.DB
}

// 确保 datastore 实现了 IStore 接口.
var _ IStore = (*datastore)(nil)

// NewStore 创建一个 IStore 类型的实例.
func NewStore(db *gorm.DB) *datastore {
	// 确保 S 只被初始化一次
	once.Do(func() {
		S = &datastore{db}
	})

	return S
}

// DB 返回存储在 datastore 中的 *gorm.DB.
func (ds *datastore) DB() *gorm.DB {
	return ds.db
}

// Users 返回一个实现了 UserStore 接口的实例.
func (ds *datastore) Users() UserStore {
	return newUsers(ds.db)
}

func (ds *datastore) AuthProviders() AuthProviderStore {
	return NewAuthProviders(ds.db)
}

func (ds *datastore) UserAccounts() UserAccountStore {
	return NewUserAccounts(ds.db)
}

func (ds *datastore) AppVersions() syscfg.AppVersionStore {
	return syscfg.NewAppVersions(ds.db)
}

func (ds *datastore) Configs() syscfg.ConfigStore {
	return syscfg.NewConfigs(ds.db)
}

func (ds *datastore) Apps() AppStore {
	return NewApps(ds.db)
}

func (ds *datastore) ApiKeys() ApiKeyStore {
	return NewApiKeys(ds.db)
}
