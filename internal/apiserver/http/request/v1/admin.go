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
	Status   int    `json:"status"`
	RoleName string `json:"roleName"`

	Role  *RoleInfo  `json:"role,omitempty"`
	Roles []RoleInfo `json:"roles"`
}

type ListAdminRequest struct {
	gormutil.ListOptions

	Username string `form:"username"`
	Nickname string `form:"nickname"`
	Status   *int   `form:"status"`
	RoleName string `form:"roleName"`
	Email    string `form:"email"`
	Phone    string `form:"phone"`
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
	Nickname  *string  `json:"nickname" binding:"min=2,max=20"`
	Password  *string  `json:"password" binding:"omitempty,min=6,max=18"`
	Email     *string  `json:"email" binding:"omitempty,email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	Status    *int     `json:"status"`
	RoleNames []string `json:"roleNames"`
}

type SetRolesRequest struct {
	RoleNames []string `json:"roleNames" binding:"required"`
}

type SwitchRoleRequest struct {
	RoleName string `json:"roleName" binding:"required"`
}
