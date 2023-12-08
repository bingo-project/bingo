package v1

import (
	"time"
)

type RoleInfo struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListRoleRequest struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type ListRoleResponse struct {
	Total int64       `json:"total"`
	Data  []*RoleInfo `json:"data"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" valid:"required,alphanum,stringlength(1|20)"`
	Description string `json:"description" valid:"required,stringlength(1|255)"`
}

type GetRoleResponse RoleInfo

type UpdateRoleRequest struct {
	Name        *string `json:"name" valid:"alphanum,stringlength(1|20)"`
	Description *string `json:"description" valid:"stringlength(1|255)"`
}

type SetPermissionsRequest struct {
	PermissionIDs []uint `json:"permissionIDs" valid:"required"`
}
