# Root 用户权限模型设计

## 背景

当前系统中 root 用户和 root 角色是内置的，通过硬编码给予所有权限和菜单。但前端仍可通过分配角色给其他用户 root 权限，存在安全隐患。

## 设计目标

1. 移除 root 角色概念，仿 Linux UID 0 模型
2. root 用户的特殊权限通过 username 硬编码判断
3. 隐藏 root 的存在，外部无法感知

## 核心设计

### 权限模型

- root 用户的超级权限通过 `username == "root"` 硬编码判断，不依赖角色
- root 角色从角色表中移除，"root" 只作为 RoleName 的标识值
- 权限检查第一步判断：`if IsRoot(username, roleName)` → 直接放行

### 角色扮演机制

- root 用户关联普通角色（如 super-admin、admin）
- 切换到普通角色时，走正常权限检查，体验受限权限
- 切回 "root" 时，恢复超级权限

### 权限检查流程

```
请求 → IsRoot(username, roleName)?
       ├─ 是 → 直接放行（超级权限）
       └─ 否 → ResolveSubject → casbin 检查
```

## Helper 函数

位置：`internal/pkg/known/known.go`

```go
const UserRoot = "root"

// IsRoot 判断是否为 root 用户且当前处于 root 权限状态
func IsRoot(username, roleName string) bool {
    return username == UserRoot && roleName == UserRoot
}
```

## 边界情况处理

| 场景 | 处理方式 |
|------|---------|
| 创建角色 name="root" | 拒绝，返回 "该名称不可用" |
| 创建用户 username="root" | 拒绝，返回 "该用户名不可用" |
| 给用户分配角色时传入 "root" | 正常过滤掉（查不到），不会生效 |
| root 用户切换到 "root" | 允许，将 RoleName 设为 "root" |
| root 用户切换到不存在的角色 | 拒绝，返回 404 |
| 删除/查询 root 用户 | 返回 404（隐藏存在） |

## 变更文件清单

| 文件 | 变更内容 |
|------|---------|
| `internal/pkg/known/known.go` | 新增 `UserRoot` 常量和 `IsRoot()` 函数 |
| `internal/pkg/database/seeder/role_seeder.go` | 移除 root 角色 |
| `internal/pkg/database/seeder/admin_seeder.go` | root 用户关联 super-admin 角色；移除 casbin policy |
| `internal/pkg/auth/middleware.go` | AuthzMiddleware 添加 `IsRoot()` 短路判断 |
| `internal/admserver/biz/system/role.go` | GetMenuTree 改用 `IsRoot()`；Create 禁止 name="root"；移除原有 RoleRoot 判断 |
| `internal/admserver/handler/http/system/auth.go` | Menus 接口传 username 和 roleName |
| `internal/admserver/biz/system/admin.go` | SwitchRole 对 root 用户允许切换到 "root"；Create 禁止 username="root"；Delete 对 root 返回 404 |
| `internal/pkg/store/sys_role.go` | ListWithRequest 移除 root 过滤（不再需要） |

## admin seeder 变更

```go
admin := model.AdminM{
    Username: "root",
    Password: "123456",
    Nickname: "Root",
    RoleName: "root",  // 保持 "root" 作为标识
}

// 关联 super-admin 角色
roles, _ := store.S.SysRole().GetByNames(ctx, []string{"super-admin"})
admin.Roles = roles

// 移除 casbin AddNamedPolicy，改用中间件判断
```
