package middleware

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/bingo-project/bingo/pkg/contextx"
)

var Ctx = context.Background()

func Context(s *discordgo.Session, i *discordgo.InteractionCreate) {
	user := i.User
	if i.User == nil {
		user = i.Member.User
	}

	Ctx = contextx.WithRequestID(Ctx, uuid.New().String())
	Ctx = contextx.WithUserID(Ctx, fmt.Sprintf("%s#%s", user.Username, user.Discriminator))
	Ctx = contextx.WithObject(Ctx, i.ApplicationCommandData().Name)
	Ctx = contextx.WithInstance(Ctx, i.ChannelID)
}
