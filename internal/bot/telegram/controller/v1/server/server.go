package server

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	"bingo/internal/admserver/biz"
	"bingo/internal/admserver/store"
	mw "bingo/internal/bot/telegram/middleware"
	"bingo/internal/pkg/model/bot"
	v1 "bingo/pkg/api/apiserver/v1/bot"
)

type ServerController struct {
	b biz.IBiz
}

func New(ds store.IStore) *ServerController {
	return &ServerController{b: biz.NewBiz(ds)}
}

func (ctrl *ServerController) Pong(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Pong function called")

	return c.Send("pong")
}

func (ctrl *ServerController) Healthz(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(mw.Ctx)
	if err != nil {
		return err
	}

	return c.Send(status)
}

func (ctrl *ServerController) Version(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	return c.Send(v)
}

func (ctrl *ServerController) ToggleMaintenance(c telebot.Context) error {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(mw.Ctx)
	if err != nil {
		return c.Send("Operation failed:" + err.Error())
	}

	return c.Send("Operation success")
}

func (ctrl *ServerController) Subscribe(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Subscribe function called")

	req := v1.CreateChannelRequest{
		Source:    string(bot.SourceTelegram),
		ChannelID: cast.ToString(c.Chat().ID),
		Author:    convertor.ToString(c.Sender()),
	}

	_, err := ctrl.b.Channels().Create(mw.Ctx, &req)
	if err != nil {
		return c.Send(err.Error())
	}

	return c.Send("Successfully subscribe, enjoy it!")
}

func (ctrl *ServerController) UnSubscribe(c telebot.Context) error {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(mw.Ctx, cast.ToString(c.Chat().ID))
	if err != nil {
		return c.Send(err.Error())
	}

	return c.Send("Successfully unsubscribe, thanks for your support!")
}
