package biz

//go:generate mockgen -destination mock_biz.go -package biz bingo/internal/admserver/biz IBiz

import (
	"github.com/bingo-project/bingo/internal/admserver/biz/ai"
	"github.com/bingo-project/bingo/internal/admserver/biz/app"
	"github.com/bingo-project/bingo/internal/admserver/biz/auth"
	"github.com/bingo-project/bingo/internal/admserver/biz/bot"
	"github.com/bingo-project/bingo/internal/admserver/biz/common"
	"github.com/bingo-project/bingo/internal/admserver/biz/file"
	"github.com/bingo-project/bingo/internal/admserver/biz/notification"
	"github.com/bingo-project/bingo/internal/admserver/biz/syscfg"
	"github.com/bingo-project/bingo/internal/admserver/biz/system"
	"github.com/bingo-project/bingo/internal/admserver/biz/user"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

// IBiz 定义了 Biz 层需要实现的方法.
type IBiz interface {
	Auth() auth.AuthBiz
	AuthProviders() auth.AuthProviderBiz
	Users() user.UserBiz

	AiAgents() ai.AiAgentBiz
	AiProviders() ai.AiProviderBiz
	AiModels() ai.AiModelBiz
	AiQuotas() ai.AiQuotaBiz

	Servers() syscfg.ServerBiz
	Email() common.EmailBiz
	Files() file.FileBiz

	Admins() system.AdminBiz
	Roles() system.RoleBiz
	Apis() system.ApiBiz
	Menus() system.MenuBiz

	AppVersions() syscfg.AppVersionBiz
	Configs() syscfg.ConfigBiz

	Bots() bot.BotBiz
	Channels() bot.ChannelBiz
	Apps() app.AppBiz
	ApiKeys() app.ApiKeyBiz

	Announcements() notification.AnnouncementBiz
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

func (b *biz) AiAgents() ai.AiAgentBiz {
	return ai.NewAiAgent(b.ds)
}

func (b *biz) AiProviders() ai.AiProviderBiz {
	return ai.NewAiProvider(b.ds)
}

func (b *biz) AiModels() ai.AiModelBiz {
	return ai.NewAiModel(b.ds)
}

func (b *biz) AiQuotas() ai.AiQuotaBiz {
	return ai.NewAiQuota(b.ds)
}

func (b *biz) Servers() syscfg.ServerBiz {
	return syscfg.NewServer(b.ds)
}

func (b *biz) Email() common.EmailBiz {
	return common.NewEmail(b.ds)
}

func (b *biz) Files() file.FileBiz {
	return file.NewFile(b.ds)
}

// Admins 管理员.
func (b *biz) Admins() system.AdminBiz {
	return system.NewAdmin(b.ds)
}

// Roles 角色管理.
func (b *biz) Roles() system.RoleBiz {
	return system.NewRole(b.ds)
}

func (b *biz) Apis() system.ApiBiz {
	return system.NewApi(b.ds)
}

func (b *biz) Menus() system.MenuBiz {
	return system.NewMenu(b.ds)
}

func (b *biz) AppVersions() syscfg.AppVersionBiz {
	return syscfg.NewAppVersion(b.ds)
}

func (b *biz) Configs() syscfg.ConfigBiz {
	return syscfg.NewConfig(b.ds)
}

func (b *biz) Bots() bot.BotBiz {
	return bot.NewBot(b.ds)
}

func (b *biz) Channels() bot.ChannelBiz {
	return bot.NewChannel(b.ds)
}

func (b *biz) Apps() app.AppBiz {
	return app.NewApp(b.ds)
}

func (b *biz) ApiKeys() app.ApiKeyBiz {
	return app.NewApiKey(b.ds)
}

func (b *biz) Announcements() notification.AnnouncementBiz {
	return notification.NewAnnouncement(b.ds)
}
