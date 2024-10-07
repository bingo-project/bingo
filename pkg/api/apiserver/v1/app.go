package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AppInfo struct {
	ID          uint64     `json:"id"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	UID         string     `json:"uid"`
	AppID       string     `json:"appId"`
	Name        string     `json:"name"`
	Status      int32      `json:"status"` // Status, 1-enabled, 2-disabled
	Description string     `json:"description"`
	Logo        string     `json:"logo"`
}

type ListAppRequest struct {
	gormutil.ListOptions

	UID         *string `json:"uid"`
	AppID       *string `json:"appId"`
	Name        *string `json:"name"`
	Status      *int32  `json:"status"` // Status, 1-enabled, 2-disabled
	Description *string `json:"description"`
	Logo        *string `json:"logo"`
}

type ListAppResponse struct {
	Total int64     `json:"total"`
	Data  []AppInfo `json:"data"`
}

type CreateAppRequest struct {
	UID         string `json:"uid"`
	Name        string `json:"name" binding:"required"`
	Status      int32  `json:"status" binding:"required,oneof=1 2"` // Status, 1-enabled, 2-disabled
	Description string `json:"description"`
	Logo        string `json:"logo"`
}

type UpdateAppRequest struct {
	Name        *string `json:"name"`
	Status      *int32  `json:"status" binding:"omitempty,oneof=1 2"` // Status, 1-enabled, 2-disabled
	Description *string `json:"description"`
	Logo        *string `json:"logo"`
}
