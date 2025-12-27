package biz

//go:generate mockgen -destination mock_biz.go -package biz bingo/internal/apiserver/biz IBiz

import (
	"github.com/bingo-project/bingo/internal/apiserver/biz/app"
	"github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/apiserver/biz/file"
	"github.com/bingo-project/bingo/internal/apiserver/biz/syscfg"
	"github.com/bingo-project/bingo/internal/apiserver/biz/user"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// IBiz 定义了 Biz 层需要实现的方法.
type IBiz interface {
	Auth() auth.AuthBiz
	AuthProviders() auth.AuthProviderBiz
	Users() user.UserBiz

	Servers() syscfg.ServerBiz
	Files() file.FileBiz

	AppVersions() syscfg.AppVersionBiz
	Configs() syscfg.ConfigBiz

	Apps() app.AppBiz
	ApiKeys() app.ApiKeyBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	ds store.IStore
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

func (b *biz) Auth() auth.AuthBiz {
	return auth.NewAuth(b.ds)
}

func (b *biz) AuthProviders() auth.AuthProviderBiz {
	return auth.NewAuthProvider(b.ds)
}

// Users 返回一个实现了 UserBiz 接口的实例.
func (b *biz) Users() user.UserBiz {
	return user.New(b.ds)
}

func (b *biz) Servers() syscfg.ServerBiz {
	return syscfg.NewServer(b.ds)
}

func (b *biz) Files() file.FileBiz {
	return file.NewFile(b.ds)
}

func (b *biz) AppVersions() syscfg.AppVersionBiz {
	return syscfg.NewAppVersion(b.ds)
}

func (b *biz) Configs() syscfg.ConfigBiz {
	return syscfg.NewConfig(b.ds)
}

func (b *biz) Apps() app.AppBiz {
	return app.NewApp(b.ds)
}

func (b *biz) ApiKeys() app.ApiKeyBiz {
	return app.NewApiKey(b.ds)
}
