package server

import (
	"github.com/bingo-project/component-base/version"
	"github.com/bwmarrin/discordgo"
	"github.com/duke-git/lancet/v2/convertor"

	"bingo/internal/bot/biz"
	"bingo/internal/bot/discord/client"
	mw "bingo/internal/bot/discord/middleware"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model/bot"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1/bot"
)

type ServerController struct {
	b biz.IBiz
	*client.Client
}

func New(ds store.IStore, s *discordgo.Session, i *discordgo.InteractionCreate) *ServerController {
	return &ServerController{
		b:      biz.NewBiz(ds),
		Client: client.NewClient(s, i),
	}
}

func (ctrl *ServerController) Pong() {
	log.C(mw.Ctx).Infow("Pong function called")

	ctrl.WriteResponse("pong")
}

func (ctrl *ServerController) Healthz() {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(mw.Ctx)
	if err != nil {
		ctrl.WriteResponse(err.Error())

		return
	}

	ctrl.WriteResponse(status)
}

func (ctrl *ServerController) Version() {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	ctrl.WriteResponse(v)
}

func (ctrl *ServerController) ToggleMaintenance() {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(mw.Ctx)
	if err != nil {
		ctrl.WriteResponse("Operation failed:" + err.Error())

		return
	}

	ctrl.WriteResponse("Operation success")
}

func (ctrl *ServerController) Subscribe() {
	log.C(mw.Ctx).Infow("Subscribe function called")

	user := ctrl.I.User
	if ctrl.I.User == nil {
		user = ctrl.I.Member.User
	}

	req := v1.CreateChannelRequest{
		Source:    string(bot.SourceDiscord),
		ChannelID: ctrl.I.ChannelID,
		Author:    convertor.ToString(user),
	}

	_, err := ctrl.b.Channels().Create(mw.Ctx, &req)
	if err != nil {
		ctrl.WriteResponse(err.Error())

		return
	}

	ctrl.WriteResponse("Successfully subscribe, enjoy it!")
}

func (ctrl *ServerController) UnSubscribe() {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(mw.Ctx, ctrl.I.ChannelID)
	if err != nil {
		ctrl.WriteResponse(err.Error())

		return
	}

	ctrl.WriteResponse("Successfully unsubscribe, thanks for your support!")
}
