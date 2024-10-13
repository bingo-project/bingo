package auth

import (
	"time"

	casbin "github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	adapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

const (
	// casbin 访问控制模型.
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

// Authz 定义了一个授权器，提供授权功能.
type Authz struct {
	*casbin.SyncedEnforcer
}

// NewAuthz 创建一个使用 casbin 完成授权的授权器.
func NewAuthz(db *gorm.DB) (*Authz, error) {
	// Initialize a Gorm adapter and use it in a Casbin enforcer
	adapter, err := adapter.NewAdapterByDBUseTableName(db, "sys", "casbin_rule")
	if err != nil {
		return nil, err
	}

	m, _ := model.NewModelFromString(aclModel)

	// Initialize the enforcer.
	enforcer, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// Load the policy from DB.
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}
	enforcer.StartAutoLoadPolicy(time.Second * 5)

	a := &Authz{enforcer}

	return a, nil
}

// Authorize 用来进行授权.
func (a *Authz) Authorize(sub, obj, act string) (bool, error) {
	return a.Enforce(sub, obj, act)
}
