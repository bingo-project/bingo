// ABOUTME: Unified authorizer for role-based access control.
// ABOUTME: Provides Casbin-based authorization with pluggable subject resolution.

package auth

import (
	"context"
	"time"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	adapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/known"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
	"github.com/bingo-project/bingo/pkg/errorsx"
)

const (
	aclModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.sub == p.sub && keyMatch2(r.obj, p.obj) && regexMatch(r.act, p.act)`

	AclDefaultMethods = "(GET)|(POST)|(PUT)|(DELETE)"
)

// SubjectResolver resolves the authorization subject from context.
type SubjectResolver interface {
	ResolveSubject(ctx context.Context) (string, error)
}

// Authorizer provides role-based access control.
type Authorizer struct {
	enforcer *casbin.SyncedEnforcer
	resolver SubjectResolver
}

// NewAuthorizer creates a new Authorizer with the given database and subject resolver.
// Pass nil for resolver if only using Enforcer() for policy management.
func NewAuthorizer(db *gorm.DB, resolver SubjectResolver) (*Authorizer, error) {
	enforcer, err := newEnforcer(db)
	if err != nil {
		return nil, err
	}

	return &Authorizer{
		enforcer: enforcer,
		resolver: resolver,
	}, nil
}

func newEnforcer(db *gorm.DB) (*casbin.SyncedEnforcer, error) {
	a, err := adapter.NewAdapterByDBUseTableName(db, "sys", "casbin_rule")
	if err != nil {
		return nil, err
	}

	m, _ := model.NewModelFromString(aclModel)

	enforcer, err := casbin.NewSyncedEnforcer(m, a)
	if err != nil {
		return nil, err
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}
	enforcer.StartAutoLoadPolicy(time.Second * 5)

	return enforcer, nil
}

// Authorize checks if the subject in context has permission to perform the action on the object.
func (a *Authorizer) Authorize(ctx context.Context, obj, act string) error {
	if a.resolver == nil {
		return errorsx.New(500, "InternalError", "authorization resolver not configured")
	}

	sub, err := a.resolver.ResolveSubject(ctx)
	if err != nil {
		return err
	}

	allowed, err := a.enforcer.Enforce(sub, obj, act)
	if err != nil {
		return errorsx.New(500, "InternalError", "authorization check failed: %s", err.Error())
	}

	if !allowed {
		return errorsx.New(403, "Forbidden", "permission denied")
	}

	return nil
}

// Enforcer returns the underlying Casbin enforcer for advanced operations.
func (a *Authorizer) Enforcer() *casbin.SyncedEnforcer {
	return a.enforcer
}

// AuthzMiddleware returns a Gin middleware that checks authorization.
func AuthzMiddleware(a *Authorizer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Root user bypass: skip authorization check if user is root
		admin, ok := contextx.UserInfo[*v1.AdminInfo](c.Request.Context())
		if ok && known.IsRoot(admin.Username, admin.RoleName) {
			c.Next()
			return
		}

		obj := c.Request.URL.Path
		act := c.Request.Method

		if err := a.Authorize(c.Request.Context(), obj, act); err != nil {
			e := errorsx.FromError(err)
			core.Response(c, nil, e)
			c.Abort()

			return
		}

		c.Next()
	}
}
