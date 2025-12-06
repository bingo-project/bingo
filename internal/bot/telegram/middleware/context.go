package middleware

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"gopkg.in/telebot.v3"

	"bingo/pkg/contextx"
)

var Ctx = context.Background()

func Context(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		Ctx = contextx.WithRequestID(Ctx, uuid.New().String())
		Ctx = contextx.WithUserID(Ctx, strconv.FormatInt(c.Sender().ID, 10))
		Ctx = contextx.WithObject(Ctx, c.Text())

		return next(c)
	}
}
