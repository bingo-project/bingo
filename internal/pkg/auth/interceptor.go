// ABOUTME: gRPC authentication interceptor using unified authenticator.
// ABOUTME: Provides unary and stream interceptors for token-based authentication.

package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor returns a gRPC unary interceptor that authenticates requests.
// publicMethods is a map of method names that don't require authentication.
func UnaryInterceptor(a *Authenticator, publicMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// Skip authentication for public methods
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract token from metadata
		tokenStr, err := extractTokenFromMetadata(ctx)
		if err != nil {
			return nil, err
		}

		// Verify token
		ctx, err = a.Verify(ctx, tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(ctx, req)
	}
}

// StreamInterceptor returns a gRPC stream interceptor that authenticates requests.
// publicMethods is a map of method names that don't require authentication.
func StreamInterceptor(a *Authenticator, publicMethods map[string]bool) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for public methods
		if publicMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx := ss.Context()

		// Extract token from metadata
		tokenStr, err := extractTokenFromMetadata(ctx)
		if err != nil {
			return err
		}

		// Verify token
		ctx, err = a.Verify(ctx, tokenStr)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		// Wrap the stream with authenticated context
		wrapped := &wrappedStream{ServerStream: ss, ctx: ctx}

		return handler(srv, wrapped)
	}
}

// extractTokenFromMetadata extracts the bearer token from gRPC metadata.
func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}

	return ExtractBearerToken(values[0]), nil
}

// wrappedStream wraps a gRPC server stream with an authenticated context.
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the authenticated context.
func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
