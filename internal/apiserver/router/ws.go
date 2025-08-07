package router

import (
	"bingo/internal/apiserver/store"
	"bingo/internal/apiserver/ws/v1/auth"
	ws "bingo/pkg/ws/server"
)

func Websocket() {
	authController := auth.NewAuthController(store.S)

	ws.Register("login", authController.Login)
}
