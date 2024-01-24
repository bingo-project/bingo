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
	Roles []RoleInfo `json:"roles,omitempty"`
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
	Username  string   `json:"username" valid:"required,alphanum,stringlength(1|255)"`
	Password  string   `json:"password" valid:"required,stringlength(6|20)"`
	Nickname  string   `json:"nickname" valid:"required,alphanum,stringlength(1|20)"`
	Email     *string  `json:"email" valid:"email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	RoleNames []string `json:"roleNames"`
}

type UpdateAdminRequest struct {
	Nickname  *string  `json:"nickname" valid:"stringlength(1|20)"`
	Password  *string  `json:"password" valid:"stringlength(6|20)"`
	Email     *string  `json:"email" valid:"email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	Status    *int     `json:"status"`
	RoleNames []string `json:"roleNames"`
}

type SetRolesRequest struct {
	RoleNames []string `json:"roleNames" valid:"required"`
}

type SwitchRoleRequest struct {
	RoleName string `json:"roleName" valid:"required"`
}
