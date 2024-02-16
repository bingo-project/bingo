package server

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/bwmarrin/discordgo"
	"github.com/duke-git/lancet/v2/convertor"

	"bingo/internal/apiserver/biz"
	mw "bingo/internal/apiserver/bot/discord/middleware"
	v1 "bingo/internal/apiserver/http/request/v1/bot"
	"bingo/internal/apiserver/model/bot"
	"bingo/internal/apiserver/store"
)

type ServerController struct {
	b biz.IBiz
}

func New(ds store.IStore) *ServerController {
	return &ServerController{b: biz.NewBiz(ds)}
}

func (ctrl *ServerController) Pong(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("Pong function called")

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "pong",
		},
	})
}

func (ctrl *ServerController) Healthz(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(mw.Ctx)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: status,
		},
	})
}

func (ctrl *ServerController) Version(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: v,
		},
	})
}

func (ctrl *ServerController) ToggleMaintenance(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(context.Background())
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Operation failed:" + err.Error(),
			},
		})

		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Operation success",
		},
	})
}

func (ctrl *ServerController) Subscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("Subscribe function called")

	user := i.User
	if i.User == nil {
		user = i.Member.User
	}

	req := v1.CreateChannelRequest{
		Source:    string(bot.SourceDiscord),
		ChannelID: i.ChannelID,
		Author:    convertor.ToString(user),
	}

	_, err := ctrl.b.Channels().Create(mw.Ctx, &req)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully subscribe, enjoy it!",
		},
	})
}

func (ctrl *ServerController) UnSubscribe(s *discordgo.Session, i *discordgo.InteractionCreate) {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(context.Background(), i.ChannelID)
	if err != nil {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: err.Error(),
			},
		})

		return
	}

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Successfully unsubscribe, thanks for your support!",
		},
	})
}
