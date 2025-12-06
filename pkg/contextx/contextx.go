// ABOUTME: Context utilities for storing and retrieving request-scoped values.
// ABOUTME: Provides type-safe context keys for user info, access tokens, and request IDs.

package contextx

import (
	"context"
)

// 定义用于上下文的键.
type (
	userInfoKey    struct{}
	usernameKey    struct{}
	userIDKey      struct{}
	accessTokenKey struct{}
	requestIDKey   struct{}
	clientIPKey    struct{}
	taskKey        struct{}
	objectKey      struct{}
	instanceKey    struct{}
)

// WithUserInfo 将用户信息存放到上下文中.
func WithUserInfo[T any](ctx context.Context, userInfo T) context.Context {
	return context.WithValue(ctx, userInfoKey{}, userInfo)
}

// UserInfo 从上下文中提取用户信息.
func UserInfo[T any](ctx context.Context) (T, bool) {
	val, ok := ctx.Value(userInfoKey{}).(T)
	return val, ok
}

// WithUserID 将用户 ID 存放到上下文中.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID 从上下文中提取用户 ID.
func UserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey{}).(string)
	return userID
}

// WithUsername 将用户名存放到上下文中.
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey{}, username)
}

// Username User 从上下文中提取用户名.
func Username(ctx context.Context) string {
	username, _ := ctx.Value(usernameKey{}).(string)
	return username
}

// WithAccessToken 将访问令牌存放到上下文中.
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

// AccessToken 从上下文中提取访问令牌.
func AccessToken(ctx context.Context) string {
	accessToken, _ := ctx.Value(accessTokenKey{}).(string)
	return accessToken
}

// WithRequestID 将请求 ID 存放到上下文中.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// RequestID 从上下文中提取请求 ID.
func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}

// WithClientIP 将客户端 IP 存放到上下文中.
func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, clientIP)
}

// ClientIP 从上下文中提取客户端 IP.
func ClientIP(ctx context.Context) string {
	clientIP, _ := ctx.Value(clientIPKey{}).(string)
	return clientIP
}

// WithTask 将任务名存放到上下文中.
func WithTask(ctx context.Context, task string) context.Context {
	return context.WithValue(ctx, taskKey{}, task)
}

// Task 从上下文中提取任务名.
func Task(ctx context.Context) string {
	task, _ := ctx.Value(taskKey{}).(string)
	return task
}

// WithObject 将操作对象存放到上下文中.
func WithObject(ctx context.Context, object string) context.Context {
	return context.WithValue(ctx, objectKey{}, object)
}

// Object 从上下文中提取操作对象.
func Object(ctx context.Context) string {
	object, _ := ctx.Value(objectKey{}).(string)
	return object
}

// WithInstance 将实例标识存放到上下文中.
func WithInstance(ctx context.Context, instance string) context.Context {
	return context.WithValue(ctx, instanceKey{}, instance)
}

// Instance 从上下文中提取实例标识.
func Instance(ctx context.Context) string {
	instance, _ := ctx.Value(instanceKey{}).(string)
	return instance
}
