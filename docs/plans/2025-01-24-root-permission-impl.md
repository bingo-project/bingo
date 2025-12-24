# Root 用户权限模型实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 重构 root 权限模型，仿 Linux UID 0 模式，通过 username 硬编码判断而非角色。

**Architecture:** 移除 root 角色概念，新增 `IsRoot(username, roleName)` helper 函数，在授权中间件和菜单获取处添加短路判断。隐藏 root 存在，禁止创建 root 用户/角色。

**Tech Stack:** Go, Casbin, GORM, Gin

**Design Doc:** `docs/plans/2025-01-24-root-user-permission-design.md`

---

### Task 1: 添加 IsRoot Helper 函数

**Files:**
- Modify: `internal/pkg/known/known.go`

**Step 1: 添加 UserRoot 常量和 IsRoot 函数**

在 `known.go` 中添加：

```go
const (
    // UserRoot is the reserved root username
    UserRoot = "root"
)

// IsRoot checks if the user is root and currently in root privilege mode.
func IsRoot(username, roleName string) bool {
    return username == UserRoot && roleName == UserRoot
}
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/pkg/known/known.go
git commit -m "feat: add IsRoot helper function for root user detection"
```

---

### Task 2: 修改授权中间件添加 root 短路判断

**Files:**
- Modify: `internal/pkg/auth/middleware.go`

**Step 1: 查看当前中间件实现**

阅读 `internal/pkg/auth/middleware.go` 中的 `AuthzMiddleware` 函数，理解当前授权流程。

**Step 2: 添加 root 短路判断**

在 casbin 检查之前添加：

```go
import (
    "github.com/bingo-project/bingo/internal/pkg/known"
    v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
    "github.com/bingo-project/bingo/pkg/contextx"
)

// 在 AuthzMiddleware 函数中，casbin 检查之前添加：
admin, ok := contextx.UserInfo[*v1.AdminInfo](c.Request.Context())
if ok && known.IsRoot(admin.Username, admin.RoleName) {
    c.Next()
    return
}
```

**Step 3: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 4: Commit**

```bash
git add internal/pkg/auth/middleware.go
git commit -m "feat: add root user bypass in authorization middleware"
```

---

### Task 3: 修改 GetMenuTree 使用 IsRoot

**Files:**
- Modify: `internal/admserver/biz/system/role.go`
- Modify: `internal/admserver/handler/http/system/auth.go`

**Step 1: 修改 GetMenuTree 签名**

将 `GetMenuTree(ctx, roleName)` 改为 `GetMenuTree(ctx, username, roleName)`：

```go
// 接口定义
GetMenuTree(ctx context.Context, username, roleName string) ([]*v1.MenuInfo, error)

// 实现
func (b *roleBiz) GetMenuTree(ctx context.Context, username, roleName string) (ret []*v1.MenuInfo, err error) {
    var menus []*model.MenuM
    if known.IsRoot(username, roleName) {
        menus, _ = b.ds.SysMenu().AllEnabled(ctx)
    } else {
        menuIDs, _ := b.ds.SysRoleMenu().GetMenuIDsByRoleNameWithParent(ctx, roleName)
        menus, _ = b.ds.SysMenu().GetByIDs(ctx, menuIDs)
    }
    // ... 后续代码不变
}
```

**Step 2: 修改 Handler 调用**

修改 `internal/admserver/handler/http/system/auth.go`：

```go
func (ctrl *AuthHandler) Menus(c *gin.Context) {
    log.C(c).Infow("Menus function called")

    admin, _ := contextx.UserInfo[*v1.AdminInfo](c.Request.Context())

    resp, err := ctrl.b.Roles().GetMenuTree(c, admin.Username, admin.RoleName)
    // ... 后续代码不变
}
```

**Step 3: 移除 role.go 中的 RoleRoot 判断**

删除所有 `roleName == known.RoleRoot` 的判断（SetApis, SetMenus, Delete 等方法中的保护代码）。

**Step 4: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 5: Commit**

```bash
git add internal/admserver/biz/system/role.go internal/admserver/handler/http/system/auth.go
git commit -m "refactor: use IsRoot for menu tree and remove RoleRoot checks"
```

---

### Task 4: 禁止创建 root 角色

**Files:**
- Modify: `internal/admserver/biz/system/role.go`

**Step 1: 在 Create 方法中添加检查**

```go
func (b *roleBiz) Create(ctx context.Context, req *v1.CreateRoleRequest) (*v1.RoleInfo, error) {
    // 禁止创建 root 角色
    if req.Name == known.UserRoot {
        return nil, errno.ErrInvalidArgument.WithMessage("该名称不可用")
    }

    // ... 原有代码
}
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/admserver/biz/system/role.go
git commit -m "feat: prevent creating role with reserved name"
```

---

### Task 5: 禁止创建 root 用户，删除返回 404

**Files:**
- Modify: `internal/admserver/biz/system/admin.go`

**Step 1: 在 Create 方法中添加检查**

```go
func (b *adminBiz) Create(ctx context.Context, req *v1.CreateAdminRequest) (*v1.AdminInfo, error) {
    // 禁止创建 root 用户
    if req.Username == known.UserRoot {
        return nil, errno.ErrInvalidArgument.WithMessage("该用户名不可用")
    }

    // ... 原有代码
}
```

**Step 2: 在 Delete 方法中对 root 返回 404**

```go
func (b *adminBiz) Delete(ctx context.Context, username string) error {
    // root 用户返回 404（隐藏存在）
    if username == known.UserRoot {
        return errno.ErrNotFound
    }

    // ... 原有代码
}
```

**Step 3: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 4: Commit**

```bash
git add internal/admserver/biz/system/admin.go
git commit -m "feat: prevent creating root user and hide root existence"
```

---

### Task 6: 修改 SwitchRole 支持 root 用户切回 root

**Files:**
- Modify: `internal/admserver/biz/system/admin.go`

**Step 1: 修改 SwitchRole 方法**

```go
func (b *adminBiz) SwitchRole(ctx context.Context, username string, req *v1.SwitchRoleRequest) (*v1.AdminInfo, error) {
    adminM, err := b.ds.Admin().GetByUsername(ctx, username)
    if err != nil {
        return nil, errno.ErrNotFound
    }

    // root 用户可以切回 root（特殊处理）
    if username == known.UserRoot && req.RoleName == known.UserRoot {
        adminM.RoleName = known.UserRoot
        err = b.ds.Admin().Update(ctx, adminM, "role_name")
        if err != nil {
            return nil, err
        }

        var resp v1.AdminInfo
        _ = copier.Copy(&resp, adminM)
        return &resp, nil
    }

    // 检查用户是否拥有目标角色
    hasRole := b.ds.Admin().HasRole(ctx, adminM, req.RoleName)
    if !hasRole {
        return nil, errno.ErrNotFound
    }

    // ... 后续代码不变
}
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/admserver/biz/system/admin.go
git commit -m "feat: allow root user to switch back to root role"
```

---

### Task 7: 移除 role seeder 中的 root 角色

**Files:**
- Modify: `internal/pkg/database/seeder/role_seeder.go`

**Step 1: 移除 root 角色定义**

```go
var defaultRoles = []model.RoleM{
    // 移除: {Name: "root", Description: "Root", Status: "enabled"},
    {Name: "super-admin", Description: "Super admin", Status: "enabled"},
    {Name: "admin", Description: "System administrator", Status: "enabled"},
}
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/pkg/database/seeder/role_seeder.go
git commit -m "refactor: remove root role from seeder"
```

---

### Task 8: 修改 admin seeder

**Files:**
- Modify: `internal/pkg/database/seeder/admin_seeder.go`

**Step 1: 移除 casbin policy，关联 super-admin 角色**

```go
func (AdminSeeder) Run() error {
    ctx := context.Background()

    admin := model.AdminM{
        Username: "root",
        Password: "123456",
        Nickname: "Root",
        Email:    nil,
        Phone:    nil,
        RoleName: "root",  // 保持 "root" 作为标识
    }

    // Init admin account
    where := &model.AdminM{Username: admin.Username}
    if err := store.S.Admin().FirstOrCreate(ctx, where, &admin); err != nil {
        return err
    }

    // 关联 super-admin 角色（用于切换）
    roles, _ := store.S.SysRole().GetByNames(ctx, []string{"super-admin"})
    if len(roles) > 0 {
        adminM, _ := store.S.Admin().GetByUsername(ctx, admin.Username)
        if adminM != nil {
            adminM.Roles = []model.RoleM{roles[0]}
            _ = store.S.Admin().UpdateWithRoles(ctx, adminM)
        }
    }

    // 移除 casbin AddNamedPolicy，改用中间件判断

    return nil
}
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/pkg/database/seeder/admin_seeder.go
git commit -m "refactor: update admin seeder to use middleware-based root auth"
```

---

### Task 9: 移除 sys_role.go 中的 root 过滤

**Files:**
- Modify: `internal/pkg/store/sys_role.go`

**Step 1: 移除 ListWithRequest 中的 root 过滤**

删除这行：
```go
db := s.DB(ctx, opts).Where("name != ?", known.RoleRoot)
```

改为：
```go
db := s.DB(ctx, opts)
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误

**Step 3: Commit**

```bash
git add internal/pkg/store/sys_role.go
git commit -m "refactor: remove root role filter from ListWithRequest"
```

---

### Task 10: 清理 RoleRoot 常量引用

**Files:**
- Modify: `internal/pkg/known/known.go`

**Step 1: 移除或标记废弃 RoleRoot 常量**

删除或注释：
```go
// RoleRoot = "root"  // 已废弃，使用 UserRoot
```

**Step 2: 验证构建通过**

Run: `go build ./...`
Expected: 成功，无错误（如有编译错误，检查遗漏的引用）

**Step 3: Commit**

```bash
git add internal/pkg/known/known.go
git commit -m "chore: remove deprecated RoleRoot constant"
```

---

### Task 11: 最终验证

**Step 1: 完整构建验证**

Run: `go build ./...`
Expected: 成功，无错误

**Step 2: 运行测试（如有）**

Run: `go test ./...`
Expected: 所有测试通过

**Step 3: 推送到远程**

```bash
git push -u origin feature/root-permission-model
```
