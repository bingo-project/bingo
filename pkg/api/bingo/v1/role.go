package v1

import (
	"time"
)

type RoleInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ListRoleRequest struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type ListRoleResponse struct {
	TotalCount int64       `json:"totalCount"`
	Data       []*RoleInfo `json:"data"`
}

type CreateRoleRequest struct {
	Slug string `json:"slug" valid:"required,alphanum,stringlength(1|20)"`
	Name string `json:"name" valid:"required,stringlength(1|255)"`
}

type GetRoleResponse RoleInfo

type UpdateRoleRequest struct {
	Slug *string `json:"slug" valid:"alphanum,stringlength(1|20)"`
	Name *string `json:"name" valid:"stringlength(1|255)"`
}
