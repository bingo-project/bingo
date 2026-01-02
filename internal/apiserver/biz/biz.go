// ABOUTME: Business layer interface aggregator.
// ABOUTME: Defines IBiz interface that exposes all business services.

package biz

//go:generate mockgen -destination mock_biz.go -package biz bingo/internal/apiserver/biz IBiz

import (
	"github.com/bingo-project/bingo/internal/apiserver/biz/app"
	"github.com/bingo-project/bingo/internal/apiserver/biz/auth"
	"github.com/bingo-project/bingo/internal/apiserver/biz/chat"
	"github.com/bingo-project/bingo/internal/apiserver/biz/file"
	"github.com/bingo-project/bingo/internal/apiserver/biz/notification"
	"github.com/bingo-project/bingo/internal/apiserver/biz/syscfg"
	"github.com/bingo-project/bingo/internal/apiserver/biz/user"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
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

	Notifications() notification.NotificationBiz
	NotificationPreferences() notification.PreferenceBiz

	Chat() chat.ChatBiz
	AiAgents() chat.AiAgentBiz
}

// biz 是 IBiz 的一个具体实现.
type biz struct {
	ds       store.IStore
	registry *ai.Registry
}

// 确保 biz 实现了 IBiz 接口.
var _ IBiz = (*biz)(nil)

// NewBiz 创建一个 IBiz 类型的实例.
func NewBiz(ds store.IStore) *biz {
	return &biz{ds: ds}
}

// WithRegistry sets the AI registry for the biz instance.
func (b *biz) WithRegistry(registry *ai.Registry) *biz {
	b.registry = registry

	return b
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

func (b *biz) Notifications() notification.NotificationBiz {
	return notification.New(b.ds)
}

func (b *biz) NotificationPreferences() notification.PreferenceBiz {
	return notification.NewPreference(b.ds)
}

func (b *biz) Chat() chat.ChatBiz {
	return chat.New(b.ds, b.registry)
}

func (b *biz) AiAgents() chat.AiAgentBiz {
	return chat.NewAiAgent(b.ds)
}
