package middleware

import (
	"context"
	"fmt"

	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

var Ctx = context.Background()

func Context(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.User
	if i.User == nil {
		user = i.Member.User
	}

	Ctx = context.WithValue(Ctx, log.KeyTrace, uuid.New().String())
	Ctx = context.WithValue(Ctx, log.KeySubject, user.ID)
	Ctx = context.WithValue(Ctx, log.KeyObject, i.ApplicationCommandData().Name)
	Ctx = context.WithValue(Ctx, log.KeyInstance, i.ChannelID)
	Ctx = context.WithValue(Ctx, log.KeyInfo, fmt.Sprintf("%s#%s", user.Username, user.Discriminator))
}
