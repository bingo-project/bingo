package server

import (
	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/bwmarrin/discordgo"
	"github.com/duke-git/lancet/v2/convertor"

	"bingo/internal/apiserver/biz"
	v1 "bingo/internal/apiserver/http/request/v1/bot"
	"bingo/internal/apiserver/model/bot"
	"bingo/internal/apiserver/store"
	mw "bingo/internal/bot/discord/middleware"
)

type ServerController struct {
	b biz.IBiz
	s *discordgo.Session
	i *discordgo.InteractionCreate
}

func New(ds store.IStore, s *discordgo.Session, i *discordgo.InteractionCreate) *ServerController {
	return &ServerController{b: biz.NewBiz(ds), s: s, i: i}
}

func (ctrl *ServerController) Pong() {
	log.C(mw.Ctx).Infow("Pong function called")

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",
		},
	})
}

func (ctrl *ServerController) Healthz() {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(mw.Ctx)
	if err != nil {
		_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: status,
		},
	})
}

func (ctrl *ServerController) Version() {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: v,
		},
	})
}

func (ctrl *ServerController) ToggleMaintenance() {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(mw.Ctx)
	if err != nil {
		_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Operation failed:" + err.Error(),
			},
		})

		return
	}

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Operation success",
		},
	})
}

func (ctrl *ServerController) Subscribe() {
	log.C(mw.Ctx).Infow("Subscribe function called")

	user := ctrl.i.User
	if ctrl.i.User == nil {
		user = ctrl.i.Member.User
	}

	req := v1.CreateChannelRequest{
		Source:    string(bot.SourceDiscord),
		ChannelID: ctrl.i.ChannelID,
		Author:    convertor.ToString(user),
	}

	_, err := ctrl.b.Channels().Create(mw.Ctx, &req)
	if err != nil {
		_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully subscribe, enjoy it!",
		},
	})
}

func (ctrl *ServerController) UnSubscribe() {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(mw.Ctx, ctrl.i.ChannelID)
	if err != nil {
		_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = ctrl.s.InteractionRespond(ctrl.i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully unsubscribe, thanks for your support!",
		},
	})
}
