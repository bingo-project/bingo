package auth

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v62/github"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"

	"bingo/internal/pkg/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/facade"
	"bingo/internal/pkg/global"
	"bingo/internal/pkg/model"
	"bingo/pkg/api/apiserver/v1"
	"bingo/pkg/auth"
)

// AuthBiz 定义了 user 模块在 biz 层所实现的方法.
type AuthBiz interface {
	Register(ctx context.Context, r *v1.RegisterRequest) (*v1.LoginResponse, error)
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)

	Nonce(ctx *gin.Context, req *v1.AddressRequest) (ret *v1.NonceResponse, err error)
	LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (ret *v1.LoginResponse, err error)

	LoginByProvider(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest) (*v1.LoginResponse, error)
	Bind(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest, user *model.UserM) (ret *v1.UserAccountInfo, err error)

	ChangePassword(ctx context.Context, uid string, r *v1.ChangePasswordRequest) error
}

type authBiz struct {
	ds store.IStore
}

var _ AuthBiz = (*authBiz)(nil)

func NewAuth(ds store.IStore) *authBiz {
	return &authBiz{ds: ds}
}

func (b *authBiz) Register(ctx context.Context, req *v1.RegisterRequest) (*v1.LoginResponse, error) {
	user := &model.UserM{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: req.Password,
	}

	// Check exist
	exist, err := b.ds.User().IsExist(ctx, user)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errno.ErrUserAlreadyExist
	}

	// Create user
	err = b.ds.User().Create(ctx, user)
	if err != nil {
		// User exists
		if match, _ := regexp.MatchString("Duplicate entry '.*'", err.Error()); match {
			return nil, errno.ErrUserAlreadyExist
		}

		return nil, err
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

func (b *authBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// Get user
	user, err := b.ds.User().GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// Check password
	err = auth.Compare(user.Password, req.Password)
	if err != nil {
		return nil, errno.ErrPasswordIncorrect
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

func (b *authBiz) LoginByProvider(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest) (*v1.LoginResponse, error) {
	// Get provider
	provider = strings.ToLower(provider)
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, provider)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: []string{"user"},
	}

	// Get Access Token
	oauthToken, err := conf.Exchange(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	account, err := b.GetUserInfo(ctx, provider, oauthToken.AccessToken)
	if err != nil {
		return nil, err
	}

	account.UID = facade.Snowflake.Generate().String()
	user := &model.UserM{
		UID:           account.UID,
		Email:         account.Email,
		Status:        model.UserStatusEnabled,
		Avatar:        account.Avatar,
		LastLoginTime: pointer.Of(time.Now()),
		LastLoginIP:   ctx.ClientIP(),
		LastLoginType: provider,
	}

	err = b.ds.User().CreateWithAccount(ctx, user, account)

	// Get user
	user, err = b.ds.User().GetByUID(ctx, account.UID)
	if err != nil {
		return nil, errno.ErrUserNotFound
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

func (b *authBiz) Bind(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest, user *model.UserM) (ret *v1.UserAccountInfo, err error) {
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, provider)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: []string{"user"},
	}

	// Get Access Token
	oauthToken, err := conf.Exchange(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	account, err := b.GetUserInfo(ctx, provider, oauthToken.AccessToken)
	if err != nil {
		return nil, err
	}

	// Check exist
	exist := b.ds.UserAccount().CheckExist(ctx, provider, account.AccountID)
	if exist {
		return nil, errno.ErrUserAccountAlreadyExist
	}

	// Create account
	account.UID = user.UID
	err = b.ds.UserAccount().Create(ctx, account)
	if err != nil {
		return
	}

	var resp v1.UserAccountInfo
	_ = copier.Copy(&resp, account)

	return &resp, err
}

// GetUserInfo todo::other provider
func (b *authBiz) GetUserInfo(ctx context.Context, provider, token string) (ret *model.UserAccount, err error) {
	// Get User info
	client := github.NewClient(nil).WithAuthToken(token)
	data, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	ret = &model.UserAccount{
		Provider:  provider,
		AccountID: cast.ToString(&data.ID),
		Username:  cast.ToString(data.Login),
		Nickname:  cast.ToString(data.Name),
		Email:     cast.ToString(data.Email),
		Bio:       cast.ToString(data.Bio),
		Avatar:    cast.ToString(data.AvatarURL),
	}

	return
}

func (b *authBiz) ChangePassword(ctx context.Context, uid string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordIncorrect
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)
	if err := b.ds.User().Update(ctx, userM); err != nil {
		return err
	}

	return nil
}
