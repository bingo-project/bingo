package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AdminInfo struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"`
	RoleName string `json:"roleName"`

	Role  *RoleInfo  `json:"role,omitempty"`
	Roles []RoleInfo `json:"roles"`
}

type ListAdminRequest struct {
	gormutil.ListOptions

	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	RoleName string `form:"roleName"`
}

type ListAdminResponse struct {
	Total int64       `json:"total"`
	Data  []AdminInfo `json:"data"`
}

type CreateAdminRequest struct {
	Username  string   `json:"username" binding:"required,alphanum,min=2,max=255"`
	Password  string   `json:"password" binding:"required,min=6,max=18"`
	Nickname  string   `json:"nickname" binding:"required,alphanum,min=2,max=20"`
	Email     *string  `json:"email" binding:"omitempty,email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	RoleNames []string `json:"roleNames"`
}

type UpdateAdminRequest struct {
	Nickname  *string  `json:"nickname" binding:"omitempty,min=2,max=20"`
	Email     *string  `json:"email" binding:"omitempty,email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	Status    string   `json:"status"`
	RoleNames []string `json:"roleNames"`
}

type ResetAdminPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=18"`
}

type SetRolesRequest struct {
	RoleNames []string `json:"roleNames" binding:"required"`
}

type SwitchRoleRequest struct {
	RoleName string `json:"roleName" binding:"required"`
	TOTPCode string `json:"totpCode,omitempty"` // TOTP 验证码（角色要求时必填）
}
