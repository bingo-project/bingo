package middleware

import (
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/store"
)

func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		admin, _ := store.S.BotAdmin().IsAdmin(Ctx, cast.ToString(c.Sender().ID))
		if !admin {
			return c.Send(errno.ErrPermissionDenied.Message)
		}

		return next(c)
	}
}
