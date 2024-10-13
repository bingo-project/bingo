package biz

//go:generate mockgen -destination mock_biz.go -package biz bingo/internal/bot/biz IBiz

import (
	"bingo/internal/bot/biz/bot"
	"bingo/internal/bot/biz/syscfg"
	"bingo/internal/bot/store"
)

// IBiz 定义了 Biz 层需要实现的方法.
type IBiz interface {
	Servers() syscfg.ServerBiz
	Configs() syscfg.ConfigBiz

	Bots() bot.BotBiz
	Channels() bot.ChannelBiz
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

func (b *biz) Servers() syscfg.ServerBiz {
	return syscfg.NewServer(b.ds)
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
