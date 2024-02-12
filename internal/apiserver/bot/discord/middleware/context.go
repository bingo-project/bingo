package middleware

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

var Ctx = context.Background()

func Context(s *discordgo.Session, m *discordgo.MessageCreate) {
	Ctx = context.WithValue(Ctx, log.KeyTrace, uuid.New().String())
	Ctx = context.WithValue(Ctx, log.KeySubject, m.Author.ID)
	Ctx = context.WithValue(Ctx, log.KeyObject, m.Content)
}
