package handler

import (
	"github.com/bingo-project/component-base/version"
	"github.com/duke-git/lancet/v2/convertor"
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	"github.com/bingo-project/bingo/internal/bot/biz"
	mw "github.com/bingo-project/bingo/internal/bot/telegram/middleware"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model/bot"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1/bot"
)

type ServerHandler struct {
	b biz.IBiz
}

func NewServerHandler(ds store.IStore) *ServerHandler {
	return &ServerHandler{b: biz.NewBiz(ds)}
}

func (ctrl *ServerHandler) Pong(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Pong function called")

	return c.Send("pong")
}

func (ctrl *ServerHandler) Healthz(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(mw.Ctx)
	if err != nil {
		return err
	}

	return c.Send(status)
}

func (ctrl *ServerHandler) Version(c telebot.Context) error {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	return c.Send(v)
}

func (ctrl *ServerHandler) ToggleMaintenance(c telebot.Context) error {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(mw.Ctx)
	if err != nil {
		return c.Send("Operation failed:" + err.Error())
	}

	return c.Send("Operation success")
}

func (ctrl *ServerHandler) Subscribe(c telebot.Context) error {
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

func (ctrl *ServerHandler) UnSubscribe(c telebot.Context) error {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(mw.Ctx, cast.ToString(c.Chat().ID))
	if err != nil {
		return c.Send(err.Error())
	}

	return c.Send("Successfully unsubscribe, thanks for your support!")
}
