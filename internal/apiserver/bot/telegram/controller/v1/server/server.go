package server

import (
	"context"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/version"
	"gopkg.in/telebot.v3"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
)

type ServerController struct {
	b biz.IBiz
}

func New(ds store.IStore) *ServerController {
	return &ServerController{b: biz.NewBiz(ds)}
}

func (ctrl *ServerController) Healthz(c telebot.Context) error {
	log.Infow("Healthz function called")

	return c.Send("ok")
}

func (ctrl *ServerController) Version(c telebot.Context) error {
	log.Infow("Version function called")

	v := version.Get().GitVersion

	return c.Send(v)
}

func (ctrl *ServerController) ToggleMaintenance(c telebot.Context) error {
	log.Infow("ToggleMaintenance function called")

	err := ctrl.b.Servers().ToggleMaintenance(context.Background())
	if err != nil {
		return c.Send("Operation failed:" + err.Error())
	}

	return c.Send("Operation success")
}
