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
}

type UpdateRoleRequest struct {
	Description *string `json:"description" binding:"omitempty,min=1,max=255"`
	Status      *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	Remark      *string `json:"remark" binding:"omitempty,max=255"`
}

type SetApisRequest struct {
	ApiIDs []uint `json:"apiIds"`
}

type SetMenusRequest struct {
	MenuIDs []uint `json:"menuIds"`
}
