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

func (ctrl *ServerController) Pong(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("Pong function called")

	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.Errorw(err.Error())
	}
}

func (ctrl *ServerController) Healthz(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(context.Background())
	if err != nil {
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, status)
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())
	}
}

func (ctrl *ServerController) Version(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("Version function called")

	v := version.Get().GitVersion

	_, err := s.ChannelMessageSend(m.ChannelID, v)
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())
	}
}

func (ctrl *ServerController) ToggleMaintenance(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(context.Background())
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, "Operation failed:"+err.Error())
		if err != nil {
			log.Errorw("send message error", log.KeyResult, err.Error())

			return
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Operation success")
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())
	}
}

func (ctrl *ServerController) Subscribe(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("Subscribe function called")

	req := v1.CreateChannelRequest{
		Source:    string(bot.SourceDiscord),
		ChannelID: m.ChannelID,
		Author:    convertor.ToString(m.Message.Author),
	}

	_, err := ctrl.b.Channels().Create(context.Background(), &req)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
		if err != nil {
			log.Errorw("send message error", log.KeyResult, err.Error())

			return
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Successfully subscribe, enjoy it!")
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())
	}
}

func (ctrl *ServerController) UnSubscribe(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.C(mw.Ctx).Infow("UnSubscribe function called")

	err := ctrl.b.Channels().DeleteChannel(context.Background(), m.ChannelID)
	if err != nil {
		_, err := s.ChannelMessageSend(m.ChannelID, err.Error())
		if err != nil {
			log.Errorw("send message error", log.KeyResult, err.Error())

			return
		}
	}

	_, err = s.ChannelMessageSend(m.ChannelID, "Successfully unsubscribe, thanks for your support!")
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())
	}
}
