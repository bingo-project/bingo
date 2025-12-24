// ABOUTME: 定义全局业务常量.
// ABOUTME: 包括 HTTP Header、角色、并发参数、缓存键等.

package known

// HTTP/gRPC Header 常量.
// gRPC 底层使用 HTTP/2，规范要求 Header 键必须小写.
// 为兼容性统一使用小写，以 x- 开头表示自定义 Header.
const (
	XRequestID    = "x-request-id"
	XUserID       = "x-user-id"
	XUsername     = "x-username"
	XForwardedFor = "x-forwarded-for"
)

// 用户常量.
const (
	// UserRoot is the reserved root username
	UserRoot = "root"
)

// 角色常量.
const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RolePrefix = "role::" // Casbin 规则中的角色前缀.
)

// IsRoot checks if the user is root and currently in root privilege mode.
func IsRoot(username, roleName string) bool {
	return username == UserRoot && roleName == UserRoot
}

// 并发与批量处理参数.
const (
	CreateBatchSize        = 100
	MaxErrGroupConcurrency = 1000
)

// 缓存键前缀.
const (
	CacheKeyVerifyCodeTTL     = "verify_code_ttl:"
	CacheKeyVerifyCodeWaiting = "verify_code_waiting:"
)
