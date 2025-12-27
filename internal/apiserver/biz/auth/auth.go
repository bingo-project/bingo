package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"golang.org/x/oauth2"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AuthBiz 定义了 user 模块在 biz 层所实现的方法.
type AuthBiz interface {
	Register(ctx context.Context, r *v1.RegisterRequest) (*v1.LoginResponse, error)
	Login(ctx context.Context, r *v1.LoginRequest) (*v1.LoginResponse, error)

	Nonce(ctx *gin.Context, req *v1.AddressRequest) (ret *v1.NonceResponse, err error)
	LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (ret *v1.LoginResponse, err error)

	GetAuthCode(ctx *gin.Context, provider string) (*v1.GetAuthCodeResponse, error)
	LoginByProvider(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest) (*v1.LoginResponse, error)
	Bind(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest, user *v1.UserInfo) (ret *v1.UserAccountInfo, err error)

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
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return nil, err
	}

	// 构建用户
	user := &model.UserM{
		Nickname: req.Nickname,
		Password: req.Password,
		Status:   model.UserStatusEnabled,
	}

	// 根据类型设置 email 或 phone
	switch accountType {
	case AccountTypeEmail:
		user.Email = req.Account
	case AccountTypePhone:
		user.Phone = req.Account
	}

	// 检查用户是否存在
	exist, err := b.ds.User().IsExist(ctx, user)
	if err != nil {
		return nil, err
	}
	if exist {
		return nil, errno.ErrUserAlreadyExist
	}

	// 创建用户
	err = b.ds.User().Create(ctx, user)
	if err != nil {
		if match, _ := regexp.MatchString("Duplicate entry '.*'", err.Error()); match {
			return nil, errno.ErrUserAlreadyExist
		}
		return nil, err
	}

	// 生成 token
	t, err := token.Sign(user.UID, known.RoleUser)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}, nil
}

func (b *authBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	// 检测账号类型
	accountType, err := DetectAccountType(req.Account)
	if err != nil {
		return nil, err
	}

	// 查找用户
	var user *model.UserM
	switch accountType {
	case AccountTypeEmail:
		user, err = b.ds.User().FindByEmail(ctx, req.Account)
	case AccountTypePhone:
		user, err = b.ds.User().FindByPhone(ctx, req.Account)
	}
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 验证密码
	if err := auth.Compare(user.Password, req.Password); err != nil {
		return nil, errno.ErrPasswordInvalid
	}

	// 更新登录信息
	user.LastLoginTime = pointer.Of(time.Now())
	user.LastLoginType = string(accountType)
	_ = b.ds.User().Update(ctx, user, "last_login_time", "last_login_type")

	// 生成 token
	t, err := token.Sign(user.UID, known.RoleUser)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}, nil
}

func (b *authBiz) GetAuthCode(ctx *gin.Context, providerName string) (*v1.GetAuthCodeResponse, error) {
	providerName = strings.ToLower(providerName)
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, providerName)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Parse scopes
	scopes := strings.Split(oauthProvider.Scopes, " ")
	if len(scopes) == 0 || scopes[0] == "" {
		scopes = []string{"user"}
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: oauthProvider.ClientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: scopes,
	}

	// Generate state
	state, err := auth.GenerateState()
	if err != nil {
		return nil, err
	}

	// Save state to Redis
	if err := auth.SaveState(ctx, facade.Redis, state); err != nil {
		return nil, err
	}

	resp := &v1.GetAuthCodeResponse{
		State: state,
	}

	// Build auth URL options
	opts := []oauth2.AuthCodeOption{oauth2.SetAuthURLParam("state", state)}

	// PKCE support
	if oauthProvider.PKCEEnabled {
		codeVerifier, err := auth.GenerateCodeVerifier()
		if err != nil {
			return nil, err
		}
		codeChallenge := auth.GenerateCodeChallenge(codeVerifier)
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
		resp.CodeVerifier = codeVerifier
	}

	resp.AuthURL = conf.AuthCodeURL(state, opts...)

	return resp, nil
}

func (b *authBiz) LoginByProvider(ctx *gin.Context, provider string, req *v1.LoginByProviderRequest) (*v1.LoginResponse, error) {
	// Validate state
	if req.State != "" {
		if err := auth.ValidateAndDeleteState(ctx, facade.Redis, req.State); err != nil {
			return nil, errno.ErrInvalidState
		}
	}

	// Get provider
	provider = strings.ToLower(provider)
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, provider)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Parse scopes
	scopes := strings.Split(oauthProvider.Scopes, " ")
	if len(scopes) == 0 || scopes[0] == "" {
		scopes = []string{"user"}
	}

	clientSecret := oauthProvider.ClientSecret

	// Apple: Generate JWT client_secret dynamically
	if provider == model.AuthProviderApple && oauthProvider.Info != "" {
		appleConfig, err := auth.ParseAppleConfig(oauthProvider.Info)
		if err == nil && appleConfig.PrivateKey != "" {
			generatedSecret, err := auth.GenerateAppleClientSecret(oauthProvider.ClientID, appleConfig)
			if err != nil {
				return nil, err
			}
			clientSecret = generatedSecret
		}
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: clientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: scopes,
	}

	// Exchange options
	var opts []oauth2.AuthCodeOption
	if oauthProvider.PKCEEnabled && req.CodeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", req.CodeVerifier))
	}

	// Get Access Token
	oauthToken, err := conf.Exchange(ctx, req.Code, opts...)
	if err != nil {
		return nil, err
	}

	account, err := b.GetUserInfo(ctx, oauthProvider, oauthToken.AccessToken)
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

func (b *authBiz) Bind(ctx *gin.Context, providerName string, req *v1.LoginByProviderRequest, user *v1.UserInfo) (ret *v1.UserAccountInfo, err error) {
	oauthProvider, err := b.ds.AuthProvider().FirstEnabled(ctx, providerName)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	// Parse scopes
	scopes := strings.Split(oauthProvider.Scopes, " ")
	if len(scopes) == 0 || scopes[0] == "" {
		scopes = []string{"user"}
	}

	clientSecret := oauthProvider.ClientSecret

	// Apple: Generate JWT client_secret dynamically
	if providerName == model.AuthProviderApple && oauthProvider.Info != "" {
		appleConfig, err := auth.ParseAppleConfig(oauthProvider.Info)
		if err == nil && appleConfig.PrivateKey != "" {
			generatedSecret, err := auth.GenerateAppleClientSecret(oauthProvider.ClientID, appleConfig)
			if err != nil {
				return nil, err
			}
			clientSecret = generatedSecret
		}
	}

	conf := oauth2.Config{
		ClientID:     oauthProvider.ClientID,
		ClientSecret: clientSecret,
		RedirectURL:  oauthProvider.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  oauthProvider.AuthURL,
			TokenURL: oauthProvider.TokenURL,
		},
		Scopes: scopes,
	}

	// Get Access Token
	oauthToken, err := conf.Exchange(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	account, err := b.GetUserInfo(ctx, oauthProvider, oauthToken.AccessToken)
	if err != nil {
		return nil, err
	}

	// Check exist
	exist := b.ds.UserAccount().CheckExist(ctx, providerName, account.AccountID)
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

// GetUserInfo fetches user info from OAuth provider using configurable field mapping.
func (b *authBiz) GetUserInfo(ctx context.Context, provider *model.AuthProvider, accessToken string) (ret *model.UserAccount, err error) {
	url := provider.UserInfoURL

	// Facebook: token 放在 query parameter
	if provider.TokenInQuery {
		if strings.Contains(url, "?") {
			url += "&access_token=" + accessToken
		} else {
			url += "?access_token=" + accessToken
		}
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	// 标准 Bearer token
	if !provider.TokenInQuery {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	// 额外 headers
	if provider.ExtraHeaders != "" {
		var headers map[string]string
		_ = json.Unmarshal([]byte(provider.ExtraHeaders), &headers)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&data)

	var mapping map[string]string
	_ = json.Unmarshal([]byte(provider.FieldMapping), &mapping)

	account := &model.UserAccount{
		Provider:  provider.Name,
		AccountID: getNestedString(data, mapping["account_id"]),
		Username:  getNestedString(data, mapping["username"]),
		Nickname:  getNestedString(data, mapping["nickname"]),
		Email:     getNestedString(data, mapping["email"]),
		Avatar:    getNestedString(data, mapping["avatar"]),
		Bio:       getNestedString(data, mapping["bio"]),
	}

	return account, nil
}

// getNestedString extracts a string value from nested map using dot notation path (e.g., "data.id").
func getNestedString(data map[string]any, path string) string {
	if path == "" {
		return ""
	}
	parts := strings.Split(path, ".")
	current := data
	for i, part := range parts {
		val, ok := current[part]
		if !ok {
			return ""
		}
		if i == len(parts)-1 {
			return cast.ToString(val)
		}
		if next, ok := val.(map[string]any); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}

func (b *authBiz) ChangePassword(ctx context.Context, uid string, req *v1.ChangePasswordRequest) error {
	userM, err := b.ds.User().GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Check password
	if err := auth.Compare(userM.Password, req.PasswordOld); err != nil {
		return errno.ErrPasswordInvalid
	}

	// Update password
	userM.Password, _ = auth.Encrypt(req.PasswordNew)

	return b.ds.User().Update(ctx, userM, "password")
}
