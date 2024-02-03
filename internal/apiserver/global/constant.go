package global

const (
	RolePrefix = "role::" // Role prefix, only for casbin rule.
	RoleRoot   = "root"   // Root has all permissions.
	AuthAdmin  = "system" // Auth guard: system admin.

	CreateBatchSize = 1000

	CacheKeyVerifyCodeTtl     = "verify_code_ttl:"
	CacheKeyVerifyCodeWaiting = "verify_code_waiting:"
)
