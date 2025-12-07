// ABOUTME: gRPC authentication interceptor for apiserver.
// ABOUTME: Validates JWT tokens and loads user info into context.

package interceptor

import (
	"context"

	"github.com/bingo-project/component-base/web/token"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
	known "bingo/pkg/auth"
	"bingo/pkg/contextx"
)

// PublicMethods lists gRPC methods that don't require authentication.
var PublicMethods = map[string]bool{
	"/apiserver.v1.ApiServer/Healthz": true,
	"/apiserver.v1.ApiServer/Version": true,
	"/apiserver.v1.ApiServer/Login":   true,
}

// Authn returns a gRPC unary interceptor that authenticates requests.
func Authn(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	// Skip authentication for public methods
	if PublicMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	// Extract token from metadata
	tokenStr, err := auth.AuthFromMD(ctx, "Bearer")
	if err != nil {
		return nil, err
	}

	// Parse JWT Token
	payload, err := token.Parse(tokenStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Get user from store
	userM, _ := store.S.User().GetByUID(ctx, payload.Subject)
	if userM.ID == 0 {
		return nil, status.Error(codes.Unauthenticated, "user not found")
	}

	// Convert to DTO
	var userInfo v1.UserInfo
	_ = copier.Copy(&userInfo, userM)
	userInfo.PayPassword = userM.PayPassword != ""

	// Set user info in context
	ctx = context.WithValue(ctx, known.XUserID, userInfo.UID)
	ctx = context.WithValue(ctx, known.XUsername, userInfo.Username)
	ctx = contextx.WithUserInfo(ctx, &userInfo)
	ctx = contextx.WithUserID(ctx, userInfo.UID)
	ctx = contextx.WithUsername(ctx, userInfo.Username)

	return handler(ctx, req)
}
