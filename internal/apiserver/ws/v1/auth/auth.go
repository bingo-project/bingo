package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/bingo-project/component-base/web/token"

	"bingo/internal/apiserver/biz"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/pkg/ws/cache"
	"bingo/pkg/ws/common"
	"bingo/pkg/ws/model"
	ws "bingo/pkg/ws/server"
)

type AuthController struct {
	b biz.IBiz
}

func NewAuthController(ds store.IStore) *AuthController {
	return &AuthController{b: biz.NewBiz(ds)}
}

func (ctrl *AuthController) Login(client *ws.Client, seq string, message []byte) (code uint32, msg string, data any) {
	log.Infow("Login function called")

	code = common.OK
	currentTime := uint64(time.Now().Unix())
	request := &model.LoginRequest{}
	if err := json.Unmarshal(message, request); err != nil {
		code = common.ParameterIllegal
		log.Debugw("解析数据失败", "seq", seq, "err", err)

		return
	}

	payload, err := token.Parse(request.ServiceToken)
	if err != nil {
		code = common.UnauthorizedUserID
		msg = errno.ErrTokenInvalid.Message
		log.Errorw("login error", "err", err)

		return
	}

	// User
	c := context.Background()
	userInfo, _ := store.S.Users().GetByUID(c, payload.Subject)
	if userInfo.ID == 0 {
		code = common.UnauthorizedUserID
		msg = errno.ErrUserNotFound.Message
		log.Errorw("user not found", "err", err)

		return
	}

	log.Debugw("webSocket_request 用户登录", "seq", seq, "ServiceToken", request.ServiceToken, "userInfo", userInfo)

	client.Login(request.AppID, request.UserID, currentTime)

	// 存储数据
	userOnline := model.UserLogin(ws.ServerIp, ws.ServerPort, request.AppID, request.UserID, client.Addr, currentTime)
	err = cache.SetUserOnlineInfo(client.GetKey(), userOnline)
	if err != nil {
		code = common.ServerError
		log.Debugw("用户登录 SetUserOnlineInfo", "seq", seq, "err", err)

		return
	}

	// 用户登录
	login := &ws.Login{
		AppID:  request.AppID,
		UserID: userInfo.UID,
		Client: client,
	}

	ws.ClientManager.Login <- login

	log.Debugw("用户登录成功", "seq", seq, "addr", client.Addr, "userID", request.UserID)

	return
}
