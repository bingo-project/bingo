# 管理员和用户管理接口优化设计

## 背景

当前管理员和用户管理接口存在以下问题需要优化：

1. **密码管理**：Update 接口包含 Password 字段，但实际上应该有独立的重置密码接口
2. **搜索功能**：用户列表只有 keyword 模糊搜索，管理员列表缺少 keyword 字段，两者不一致
3. **代码 Bug**：User 创建逻辑中的错误处理有 bug
4. **硬编码权限**：ResetTOTP 接口硬编码了 root 检查，应使用 RBAC 系统

## 优化目标

1. 新增独立的密码重置接口，移除 Update 接口中的 Password 字段
2. 统一管理员和用户列表的搜索模式（精准筛选 + 模糊搜索）
3. 修复 User 创建的错误处理 bug
4. 统一 Create 接口返回值
5. 移除硬编码的权限检查，依赖 RBAC 系统

## API 变更

### 1. 新增密码重置接口

#### 管理员密码重置
```http
PUT /v1/admins/{username}/password
```

**请求体：**
```json
{
  "password": "newPassword123"  // 6-18 位字符
}
```

**响应：** 成功返回 200，无数据

**权限：** 通过 RBAC 控制

---

#### 用户密码重置
```http
PUT /v1/users/{uid}/password
```

**请求体：**
```json
{
  "password": "newPassword123"  // 6-18 位字符
}
```

**响应：** 成功返回 200，无数据

**权限：** 通过 RBAC 控制

---

### 2. 修改列表搜索接口

#### 管理员列表
```http
GET /v1/admins
```

**查询参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| keyword | string | 模糊搜索 Username/Nickname/Email/Phone |
| status | string | 精准筛选：enabled/disabled |
| roleName | string | 精准筛选角色名称 |
| page | int | 页码 |
| pageSize | int | 每页数量 |

**变更：**
- ✅ 新增 `keyword` 字段（模糊搜索）
- ❌ 移除 `username`, `nickname`, `email`, `phone` 独立字段

---

#### 用户列表
```http
GET /v1/users
```

**查询参数：**
| 参数 | 类型 | 说明 |
|------|------|------|
| keyword | string | 模糊搜索 Username/Nickname/Email/Phone |
| status | int32 | 精准筛选：1-enabled, 2-disabled |
| countryCode | string | 精准筛选国家代码 |
| page | int | 页码 |
| pageSize | int | 每页数量 |

**变更：**
- ✅ 保持 `keyword` 字段
- ✅ 新增 `status`, `countryCode` 精准筛选字段

---

### 3. 修改 Update 接口

#### UpdateAdmin
```http
PUT /v1/admins/{username}
```

**请求体变更：**
```json
{
  "nickname": "string",
  "email": "string",
  "phone": "string",
  "avatar": "string",
  "status": "string",
  "roleNames": ["string"]
}
```

**变更：** ❌ 移除 `password` 字段

---

#### UpdateUser
```http
PUT /v1/users/{uid}
```

**请求体变更：**
```json
{
  "nickname": "string",
  "email": "string",
  "phone": "string",
  "status": 1,
  "age": 0,
  "gender": "male",
  "avatar": "string"
}
```

**变更：** ❌ 移除 `password` 字段（如果存在）

---

### 4. 修改 Create 接口返回值

#### CreateUser
```http
POST /v1/users
```

**响应变更：**
```json
{
  "uid": "xxx",
  "username": "xxx",
  "nickname": "xxx",
  "email": "xxx",
  "phone": "xxx",
  "status": 1,
  "createdAt": "2025-01-11T00:00:00Z",
  "updatedAt": "2025-01-11T00:00:00Z"
  // ... 其他字段
}
```

**变更：** ✅ 返回完整的 `UserInfo`，而不是 `nil`

---

## 实现清单

### Phase 1: API 定义更新

**文件：** `pkg/api/apiserver/v1/admin.go`
- [ ] 新增 `ResetAdminPasswordRequest` 结构体
- [ ] 移除 `UpdateAdminRequest.Password` 字段

**文件：** `pkg/api/apiserver/v1/user.go`
- [ ] 新增 `ResetUserPasswordRequest` 结构体
- [ ] 修改 `ListUserRequest`：新增 `status`, `countryCode` 字段
- [ ] 移除 `UpdateUserRequest.Password` 字段（如果存在）

---

### Phase 2: Handler 层实现

**文件：** `internal/admserver/handler/http/system/admin.go`
- [ ] 新增 `ResetPassword` 方法
- [ ] 更新 Swagger 注释

**文件：** `internal/admserver/handler/http/user/user.go`
- [ ] 新增 `ResetPassword` 方法
- [ ] 更新 Swagger 注释

---

### Phase 3: Biz 层实现

**文件：** `internal/admserver/biz/system/admin.go`
- [ ] 新增 `ResetPassword(ctx, username, password)` 方法
- [ ] 移除 `Update` 方法中的密码处理逻辑
- [ ] 移除 `ResetTOTP` 中的硬编码 root 检查

**文件：** `internal/admserver/biz/user/user.go`
- [ ] 修复 `Create` 方法的错误处理 bug
- [ ] 修改 `Create` 返回值为 `(*v1.UserInfo, error)`
- [ ] 新增 `ResetPassword(ctx, uid, password)` 方法

---

### Phase 4: Store 层实现

**文件：** `internal/pkg/store/sys_admin.go`
- [ ] 修改 `ListWithRequest`：使用 `keyword` 替代独立的 username/email/phone 字段

**文件：** `internal/pkg/store/user.go`
- [ ] 修改 `ListWithRequest`：新增 `status`, `countryCode` 精准筛选

---

### Phase 5: 路由注册

**文件：** `internal/admserver/router/api.go`
- [ ] 注册 `PUT /v1/admins/:username/password` 路由
- [ ] 注册 `PUT /v1/users/:uid/password` 路由

---

### Phase 6: 测试

- [ ] 测试管理员密码重置接口
- [ ] 测试用户密码重置接口
- [ ] 测试管理员列表 keyword 搜索
- [ ] 测试用户列表精准筛选
- [ ] 测试 RBAC 权限控制
- [ ] 测试 User 创建返回值

---

## 数据库迁移

无需要数据库迁移。

## 向后兼容性

### 破坏性变更
1. ❌ `PUT /v1/admins/{username}` 不再接受 `password` 字段
2. ❌ `GET /v1/admins` 不再接受 `username`, `nickname`, `email`, `phone` 独立查询参数

### 新增功能
1. ✅ `PUT /v1/admins/{username}/password`
2. ✅ `PUT /v1/users/{uid}/password`
3. ✅ `GET /v1/admins` 新增 `keyword` 参数
4. ✅ `GET /v1/users` 新增 `status`, `countryCode` 参数

### 兼容建议
- 前端需要更新调用方式
- 移除 Update 接口中的 password 字段
- 使用新的 ResetPassword 接口

## 注意事项

1. **密码加密**：密码在存储前必须使用 `auth.Encrypt()` 加密
2. **RBAC 权限**：确保新接口在路由中注册了 RBAC 中间件
3. **Swagger 文档**：修改 API 后记得运行 `make swag` 更新文档
4. **ACL 清理**：用户删除时需要清理 ACL 策略（保持现有逻辑）
5. **Root 用户保护**：删除管理员时仍需检查是否为 root 用户
