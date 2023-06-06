package v1

import (
	"time"
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
}

type ListAdminRequest struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type ListAdminResponse struct {
	TotalCount int64        `json:"totalCount"`
	Data       []*AdminInfo `json:"data"`
}

type CreateAdminRequest struct {
	Username  string   `json:"username" valid:"required,alphanum,stringlength(1|255)"`
	Password  string   `json:"password" valid:"required,stringlength(6|20)"`
	Nickname  string   `json:"nickname" valid:"required,alphanum,stringlength(1|20)"`
	Email     *string  `json:"email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	RoleNames []string `json:"roleNames" valid:"required"`
}

type GetAdminResponse AdminInfo

type UpdateAdminRequest struct {
	Nickname  *string  `json:"nickname" valid:"required,alphanum,stringlength(1|20)"`
	Email     *string  `json:"email"`
	Phone     *string  `json:"phone"`
	Avatar    *string  `json:"avatar"`
	Status    *int     `json:"status"`
	RoleNames []string `json:"roleNames" valid:"required"`
}

type SetRolesRequest struct {
	RoleNames []string `json:"roleNames" valid:"required"`
}

type SwitchRoleRequest struct {
	RoleName string `json:"roleName" valid:"required"`
}
