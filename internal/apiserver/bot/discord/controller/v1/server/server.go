package server

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"github.com/bwmarrin/discordgo"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
)

type ServerController struct {
	b biz.IBiz
}

func New(ds store.IStore) *ServerController {
	return &ServerController{b: biz.NewBiz(ds)}
}

func (ctrl *ServerController) Pong(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Infow("Pong function called")

	_, err := s.ChannelMessageSend(m.ChannelID, "pong")
	if err != nil {
		log.Errorw(err.Error())

		return
	}

	return
}

func (ctrl *ServerController) Healthz(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Infow("Healthz function called")

	status, err := ctrl.b.Servers().Status(context.Background())
	if err != nil {
		return
	}

	_, err = s.ChannelMessageSend(m.ChannelID, status)
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())

		return
	}
}

func (ctrl *ServerController) Version(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Infow("Version function called")

	v := version.Get().GitVersion

	_, err := s.ChannelMessageSend(m.ChannelID, v)
	if err != nil {
		log.Errorw("send message error", log.KeyResult, err.Error())

		return
	}
}

func (ctrl *ServerController) ToggleMaintenance(s *discordgo.Session, m *discordgo.MessageCreate) {
	log.Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(context.Background())
	if err != nil {
		// return c.Send("Operation failed:" + err.Error())
	}
}
