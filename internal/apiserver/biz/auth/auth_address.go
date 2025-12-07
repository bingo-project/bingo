package auth

import (
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/eth/sign"
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/known"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/apiserver/v1"
)

func (b *authBiz) Nonce(ctx *gin.Context, req *v1.AddressRequest) (ret *v1.NonceResponse, err error) {
	where := model.UserAccount{AccountID: req.Address}
	account := model.UserAccount{
		UID:       facade.Snowflake.Generate().String(),
		Provider:  model.AuthProviderWallet,
		AccountID: where.AccountID,
		Nonce:     uuid.New().String(),
	}

	err = b.ds.UserAccount().FirstOrCreate(ctx, where, &account)
	if err != nil {
		return nil, err
	}

	ret = &v1.NonceResponse{
		Nonce: account.Nonce,
	}

	return
}

func (b *authBiz) LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (ret *v1.LoginResponse, err error) {
	account, err := b.ds.UserAccount().GetAccount(ctx, model.AuthProviderWallet, req.Address)
	if err != nil {
		return nil, err
	}

	// Check signature
	verified := sign.Verify(req.Address, req.Sign, account.Nonce)
	if !verified {
		return nil, errno.ErrTokenInvalid
	}

	// First or create user.
	user := &model.UserM{
		UID:           account.UID,
		Email:         account.Email,
		Status:        model.UserStatusEnabled,
		Avatar:        account.Avatar,
		LastLoginTime: pointer.Of(time.Now()),
		LastLoginIP:   ctx.ClientIP(),
		LastLoginType: account.Provider,
	}
	err = b.ds.User().FirstOrCreate(ctx, &model.UserM{UID: user.UID}, user)
	if err != nil {
		return
	}

	// Generate token
	t, err := token.Sign(user.UID, known.RoleUser)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	resp := &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}

	return resp, nil
}
