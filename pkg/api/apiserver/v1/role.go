package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type RoleInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Remark      string `json:"remark"`
	RequireTOTP bool   `json:"requireTotp"` // 是否强制要求 TOTP
}

type ListRoleRequest struct {
	gormutil.ListOptions

	Name          string     `form:"name"`
	Description   string     `form:"description"`
	Status        string     `form:"status"`
	CreatedAtFrom *time.Time `form:"createdAtFrom" time_format:"2006-01-02T15:04:05Z07:00"`
	CreatedAtTo   *time.Time `form:"createdAtTo" time_format:"2006-01-02T15:04:05Z07:00"`
}

type ListRoleResponse struct {
	Total int64      `json:"total"`
	Data  []RoleInfo `json:"data"`
}

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,alphanumdash,min=2,max=20"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Status      string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Remark      string `json:"remark" binding:"omitempty,max=255"`
	RequireTOTP *bool  `json:"requireTotp"` // 是否强制要求 TOTP
}

type UpdateRoleRequest struct {
	Description *string `json:"description" binding:"omitempty,min=1,max=255"`
	Status      *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Remark      *string `json:"remark" binding:"omitempty,max=255"`
	RequireTOTP *bool   `json:"requireTotp"` // 是否强制要求 TOTP
}

type SetApisRequest struct {
	ApiIDs []uint `json:"apiIds"`
}

type SetMenusRequest struct {
	MenuIDs []uint `json:"menuIds"`
}
