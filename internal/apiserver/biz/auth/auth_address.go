package auth

import (
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"bingo/internal/apiserver/global"
	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/eth/sign"
	"bingo/internal/pkg/facade"
	model2 "bingo/internal/pkg/model"
)

func (b *authBiz) Nonce(ctx *gin.Context, req *v1.AddressRequest) (ret *v1.NonceResponse, err error) {
	where := model2.UserAccount{AccountID: req.Address}
	account := model2.UserAccount{
		UID:       facade.Snowflake.Generate().String(),
		Provider:  model2.AuthProviderWallet,
		AccountID: where.AccountID,
		Nonce:     uuid.New().String(),
	}

	err = b.ds.UserAccounts().FirstOrCreate(ctx, where, &account)
	if err != nil {
		return nil, err
	}

	ret = &v1.NonceResponse{
		Nonce: account.Nonce,
	}

	return
}

func (b *authBiz) LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (ret *v1.LoginResponse, err error) {
	account, err := b.ds.UserAccounts().GetAccount(ctx, model2.AuthProviderWallet, req.Address)
	if err != nil {
		return nil, err
	}

	// Check signature
	verified := sign.Verify(req.Address, req.Sign, account.Nonce)
	if !verified {
		return nil, errno.ErrTokenInvalid
	}

	// First or create user.
	user := &model2.UserM{
		UID:           account.UID,
		Email:         account.Email,
		Status:        model2.UserStatusEnabled,
		Avatar:        account.Avatar,
		LastLoginTime: pointer.Of(time.Now()),
		LastLoginIP:   ctx.ClientIP(),
		LastLoginType: account.Provider,
	}
	err = b.ds.Users().FirstOrCreate(ctx, &model2.UserM{UID: user.UID}, user)
	if err != nil {
		return
	}

	// Generate token
	t, err := token.Sign(user.UID, global.AuthUser)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	resp := &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}

	return resp, nil
}
