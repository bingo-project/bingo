// ABOUTME: SIWE (Sign-In with Ethereum) wallet authentication for admin server.
// ABOUTME: Implements EIP-4361 standard for secure wallet login.

package auth

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/bingo-project/component-base/web/token"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/gin-gonic/gin"
	siwe "github.com/spruceid/siwe-go"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/known"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

const siweNoncePrefix = "siwe:nonce:"

func (b *authBiz) Nonce(ctx *gin.Context, req *v1.AddressRequest) (*v1.NonceResponse, error) {
	cfg := facade.Config.Auth
	if cfg == nil || !cfg.SIWE.Enabled {
		return nil, errno.ErrNotFound
	}

	// 1. Validate Origin against whitelist
	origin := ctx.GetHeader("Origin")
	domain, err := auth.ValidateOriginAndExtractDomain(origin, cfg.SIWE.Domains)
	if err != nil {
		log.C(ctx).Warnw("SIWE invalid origin", "origin", origin, "allowed", cfg.SIWE.Domains)
		return nil, errno.ErrInvalidOrigin
	}

	// 2. Generate random nonce
	nonce := siwe.GenerateNonce()

	// 3. Build SIWE message
	uri, _ := url.Parse(origin)
	options := map[string]interface{}{
		"statement":      cfg.SIWE.Statement,
		"chainId":        cfg.SIWE.ChainID,
		"issuedAt":       time.Now().UTC().Format(time.RFC3339),
		"expirationTime": time.Now().UTC().Add(cfg.SIWE.NonceExpiration).Format(time.RFC3339),
	}

	msg, err := siwe.InitMessage(domain, req.Address, uri.String(), nonce, options)
	if err != nil {
		log.C(ctx).Errorw("SIWE init message failed", "err", err)
		return nil, errno.ErrInternal
	}

	// 4. Store nonce in Redis with TTL
	key := siweNoncePrefix + nonce
	if err := facade.Redis.Set(ctx, key, req.Address, cfg.SIWE.NonceExpiration).Err(); err != nil {
		log.C(ctx).Errorw("SIWE save nonce failed", "err", err)
		return nil, errno.ErrInternal
	}

	return &v1.NonceResponse{
		Message: msg.String(),
		Nonce:   nonce,
	}, nil
}

func (b *authBiz) LoginByAddress(ctx *gin.Context, req *v1.LoginByAddressRequest) (*v1.LoginResponse, error) {
	cfg := facade.Config.Auth
	if cfg == nil || !cfg.SIWE.Enabled {
		return nil, errno.ErrNotFound
	}

	// 1. Parse SIWE message
	msg, err := siwe.ParseMessage(req.Message)
	if err != nil {
		log.C(ctx).Warnw("SIWE parse message failed", "err", err)
		return nil, errno.ErrInvalidSIWEMessage
	}

	// 2. Validate domain against whitelist
	if !auth.IsDomainAllowed(msg.GetDomain(), cfg.SIWE.Domains) {
		log.C(ctx).Warnw("SIWE domain not allowed", "domain", msg.GetDomain())
		return nil, errno.ErrInvalidDomain
	}

	// 3. Validate nonce (get and delete to ensure one-time use)
	key := siweNoncePrefix + msg.GetNonce()
	storedAddress, err := facade.Redis.GetDel(ctx, key).Result()
	if err != nil || !strings.EqualFold(storedAddress, msg.GetAddress().Hex()) {
		log.C(ctx).Warnw("SIWE invalid nonce", "nonce", msg.GetNonce(), "stored", storedAddress, "msg_addr", msg.GetAddress().Hex())
		return nil, errno.ErrInvalidNonce
	}

	// 4. Verify signature (includes expiration check)
	_, err = msg.Verify(req.Signature, nil, nil, nil)
	if err != nil {
		log.C(ctx).Warnw("SIWE signature verification failed", "err", err)
		return nil, errno.ErrSignatureInvalid
	}

	// 5. Get or create user
	address := msg.GetAddress().Hex()
	account, user, err := b.getOrCreateWalletUser(ctx, address)
	if err != nil {
		return nil, err
	}

	// 6. Update login info
	user.LastLoginTime = pointer.Of(time.Now())
	user.LastLoginIP = ctx.ClientIP()
	user.LastLoginType = account.Provider
	_ = b.ds.User().Update(ctx, user, "last_login_time", "last_login_ip", "last_login_type")

	// 7. Generate JWT
	t, err := token.Sign(user.UID, known.RoleUser)
	if err != nil {
		return nil, errno.ErrSignToken
	}

	return &v1.LoginResponse{
		AccessToken: t.AccessToken,
		ExpiresAt:   t.ExpiresAt,
	}, nil
}

// getOrCreateWalletUser finds or creates user by wallet address.
func (b *authBiz) getOrCreateWalletUser(ctx context.Context, address string) (*model.UserAccount, *model.UserM, error) {
	// Find existing account
	account, err := b.ds.UserAccount().GetAccount(ctx, model.AuthProviderWallet, address)
	if err == nil && account != nil {
		// Account exists, get user
		user, err := b.ds.User().GetByUID(ctx, account.UID)
		if err != nil {
			return nil, nil, errno.ErrUserNotFound
		}
		return account, user, nil
	}

	// Create new user and account
	uid := facade.Snowflake.Generate().String()
	user := &model.UserM{
		UID:    uid,
		Status: model.UserStatusEnabled,
	}

	account = &model.UserAccount{
		UID:       uid,
		Provider:  model.AuthProviderWallet,
		AccountID: address,
	}

	if err := b.ds.User().CreateWithAccount(ctx, user, account); err != nil {
		return nil, nil, errno.ErrDBWrite.WithMessage("create wallet user: %v", err)
	}

	return account, user, nil
}
