package middleware

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/google/uuid"
	"gopkg.in/telebot.v3"
)

var Ctx = context.Background()

func Context(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		Ctx = context.WithValue(Ctx, log.KeyTrace, uuid.New().String())
		Ctx = context.WithValue(Ctx, log.KeySubject, c.Sender().ID)
		Ctx = context.WithValue(Ctx, log.KeyInfo, c.Text())

		return next(c)
	}
}
