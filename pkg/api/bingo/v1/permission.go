package v1

import (
	"time"
)

type PermissionInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Method      string `json:"method"`
	Path        string `json:"path"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type ListPermissionRequest struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

type ListPermissionResponse struct {
	TotalCount int64             `json:"totalCount"`
	Data       []*PermissionInfo `json:"data"`
}

type CreatePermissionRequest struct {
	Method      string `json:"method" valid:"required,alphanum,stringlength(1|255)"`
	Path        string `json:"path" valid:"required,stringlength(1|255)"`
	Group       string `json:"group" valid:"required,stringlength(1|255)"`
	Description string `json:"description" valid:"required,stringlength(1|255)"`
}

type GetPermissionResponse PermissionInfo

type UpdatePermissionRequest struct {
	Method      *string `json:"method" valid:"alphanum,stringlength(1|255)"`
	Path        *string `json:"path" valid:"stringlength(1|255)"`
	Group       *string `json:"group" valid:"stringlength(1|255)"`
	Description *string `json:"description" valid:"stringlength(1|255)"`
}

type GetPermissionIDsResponse []uint
