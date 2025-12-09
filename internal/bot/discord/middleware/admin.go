package middleware

import (
	"github.com/bwmarrin/discordgo"

	"github.com/bingo-project/bingo/internal/pkg/store"
)

func IsAdmin(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	user := i.User
	if i.User == nil {
		user = i.Member.User
	}

	admin, _ := store.S.BotAdmin().IsAdmin(Ctx, user.ID)
	if !admin {
		_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Forbidden",
			},
		})
	}

	return admin
}
