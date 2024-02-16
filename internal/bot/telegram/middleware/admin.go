package middleware

import (
	"slices"

	"github.com/spf13/cast"
	"gopkg.in/telebot.v3"

	model "bingo/internal/apiserver/model/bot"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
)

func AdminOnly(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		admin, _ := store.S.BotAdmins().GetByUserID(Ctx, cast.ToString(c.Sender().ID))
		roles := []model.Role{model.RoleRoot, model.RoleAdmin}
		if !slices.Contains(roles, admin.Role) {
			return c.Send(errno.ErrForbidden.Message)
		}

		return next(c)
	}
}
