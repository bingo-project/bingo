package client

import (
	"github.com/bwmarrin/discordgo"

	"github.com/bingo-project/bingo/internal/bot/discord/middleware"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

type Client struct {
	S *discordgo.Session
	I *discordgo.InteractionCreate
}

func NewClient(s *discordgo.Session, i *discordgo.InteractionCreate) *Client {
	return &Client{s, i}
}

func (r *Client) WriteResponse(content string) {
	err := r.S.InteractionRespond(r.I.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})

	if err != nil {
		log.C(middleware.Ctx).Errorw("Discord response error", "err", err)
	}
}
