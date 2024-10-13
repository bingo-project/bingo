package middleware

import (
	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	"bingo/internal/admserver/store"
	"bingo/internal/pkg/errno"
)

func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		admin, _ := store.S.BotAdmins().IsAdmin(Ctx, cast.ToString(c.Sender().ID))
		if !admin {
			return c.Send(errno.ErrForbidden.Message)
		}

		return next(c)
	}
}
