# 菜单-API 权限关联设计

## 目标

将菜单扩展到按钮级别，每个按钮/菜单可关联 API。设置角色菜单权限时，自动提取关联的 API 同步到 Casbin，实现菜单权限驱动 API 权限。

## 设计决策

| 决策点 | 选择 | 原因 |
|--------|------|------|
| APIs 存储格式 | ID 列表 | 与现有模式一致，API 通过 FirstOrCreate 初始化 ID 稳定 |
| SetMenus 同步策略 | 菜单驱动 API 权限 | 单一入口，权限模型清晰 |
| SetApis 接口 | 暂时保留 | 调试阶段可能需要 |
| 事务处理 | 不使用 | SetMenus 幂等，失败可重试 |
| 敏感 API 过滤 | API 表加 internal 字段 | 灵活控制可绑定的 API |

## 数据模型变更

### MenuM 新增字段

```go
type MenuM struct {
    // ... 原有字段 ...

    Type     string `gorm:"type:varchar(20);not null;default:'menu'"`   // catalog/menu/button/embedded/link
    AuthCode string `gorm:"type:varchar(100);not null;default:''"`      // 权限标识，如 System:User:Create
    ApiIDs   []uint `gorm:"type:json;serializer:json"`                  // 关联的 API ID 列表
    Status   string `gorm:"type:varchar(20);not null;default:'enabled'"` // enabled/disabled
}
```

### ApiM 新增字段

```go
type ApiM struct {
    // ... 原有字段 ...

    Internal bool `gorm:"type:tinyint;not null;default:0"` // 内部 API，不可被菜单关联
}
```

## 请求/响应结构变更

### CreateMenuRequest 扩展

```go
type CreateMenuRequest struct {
    ParentID  int    `json:"parentID"`
    Title     string `json:"title" binding:"required,min=1,max=255"`
    Name      string `json:"name"`
    Path      string `json:"path" binding:"required,min=1,max=255"`
    Hidden    bool   `json:"hidden"`
    Sort      int    `json:"sort"`
    Icon      string `json:"icon"`
    Component string `json:"component"`

    // 新增
    Type     string `json:"type" binding:"omitempty,oneof=catalog menu button embedded link"`
    AuthCode string `json:"authCode" binding:"omitempty,max=100"`
    ApiIDs   []uint `json:"apiIDs"`
    Status   string `json:"status" binding:"omitempty,oneof=enabled disabled"`
}
```

### UpdateMenuRequest 扩展

同样新增对应指针字段。

### MenuInfo 扩展

同样新增对应字段。

### ApiInfo 扩展

```go
type ApiInfo struct {
    // ... 原有字段 ...

    Internal bool `json:"internal"`
}
```

## 核心逻辑：SetMenus

```go
func (b *roleBiz) SetMenus(ctx context.Context, a *auth.Authorizer, roleName string, menuIDs []uint) error {
    if roleName == known.RoleRoot {
        return errno.ErrPermissionDenied
    }

    roleM, err := b.ds.SysRole().GetByName(ctx, roleName)
    if err != nil {
        return errno.ErrNotFound
    }

    // 1. 获取菜单列表
    menus, err := b.ds.SysMenu().GetByIDs(ctx, menuIDs)
    if err != nil {
        return err
    }

    // 2. 保存菜单权限（原有逻辑）
    roleM.Menus = menus
    if err := b.ds.SysRole().UpdateWithMenus(ctx, roleM); err != nil {
        return err
    }

    // 3. 提取所有关联的 API IDs（去重）
    apiIDs := extractApiIDsFromMenus(menus)

    // 4. 同步到 Casbin
    if err := b.syncRoleApis(ctx, a, roleM.Name, apiIDs); err != nil {
        return err
    }

    return nil
}

func extractApiIDsFromMenus(menus []*model.MenuM) []uint {
    idSet := make(map[uint]struct{})
    for _, menu := range menus {
        for _, apiID := range menu.ApiIDs {
            idSet[apiID] = struct{}{}
        }
    }

    ids := make([]uint, 0, len(idSet))
    for id := range idSet {
        ids = append(ids, id)
    }
    return ids
}

func (b *roleBiz) syncRoleApis(ctx context.Context, a *auth.Authorizer, roleName string, apiIDs []uint) error {
    // 清除角色原有 API 权限
    _, err := a.Enforcer().RemoveFilteredPolicy(0, known.RolePrefix+roleName)
    if err != nil {
        return err
    }

    if len(apiIDs) == 0 {
        return nil
    }

    // 查询 API 详情
    apis, err := b.ds.SysApi().GetByIDs(ctx, apiIDs)
    if err != nil {
        return err
    }

    // 添加新的 API 权限
    rules := make([][]string, 0, len(apis))
    for _, api := range apis {
        rules = append(rules, []string{known.RolePrefix + roleName, api.Path, api.Method})
    }

    _, err = a.Enforcer().AddPolicies(rules)
    return err
}
```

## 接口变更

### 现有接口保留

```
GET  /v1/apis            # API 列表（含 internal）
GET  /v1/apis/tree       # API 树形结构（只返回 internal=false，菜单关联用）
POST /v1/apis            # 创建 API
GET  /v1/apis/:id        # 获取 API 详情
PUT  /v1/apis/:id        # 更新 API（可设置 internal）
DELETE /v1/apis/:id      # 删除 API

PUT  /v1/roles/:name/apis   # 暂时保留，调试用
GET  /v1/roles/:name/apis   # 保留，查看角色 API 权限
```

### 敏感 API 标记

以下 API 应在初始化时标记为 `internal=true`：

- `PUT /v1/roles/:name/apis` - 直接设置角色 API 权限

## 实施步骤

### 1. Migration（新增 2 个文件）

```go
// 2024_12_23_xxx_add_menu_permission_fields.go
// sys_auth_menu 表新增：type, auth_code, api_ids, status

// 2024_12_23_xxx_add_api_internal_field.go
// sys_auth_api 表新增：internal
```

### 2. 模型层

- `MenuM` 新增字段：`Type`, `AuthCode`, `ApiIDs`, `Status`
- `ApiM` 新增字段：`Internal`

### 3. API 层（pkg/api/apiserver/v1）

- `MenuInfo` / `CreateMenuRequest` / `UpdateMenuRequest` 扩展新字段
- `ApiInfo` 扩展 `Internal` 字段

### 4. Store 层

- `SysApiStore.Tree` 改造，过滤 `internal=false`

### 5. Biz 层

- `SetMenus` 增加 Casbin 同步逻辑
- 方法签名增加 `*auth.Authorizer` 参数

### 6. Handler 层

- `SetMenus` handler 传递 authorizer 实例

### 7. 数据初始化（Seed）

- 菜单 seed 包含 `type`, `authCode`, `apiIDs`, `status`
- API 初始化时标记敏感接口 `internal=true`
