package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type MenuInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ParentID  int    `json:"parentId"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Sort      int    `json:"sort"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	Hidden    bool   `json:"hidden"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	Type      string `json:"type"`
	AuthCode  string `json:"authCode"`
	Status    string `json:"status"`

	Children []MenuInfo `json:"children,omitempty"`
	ApiIDs   []uint     `json:"apiIds,omitempty"`
}

type ListMenuRequest struct {
	gormutil.ListOptions
}

type ListMenuResponse struct {
	Total int64      `json:"total"`
	Data  []MenuInfo `json:"data"`
}

type CreateMenuRequest struct {
	ParentID  int    `json:"parentId"`
	Title     string `json:"title" binding:"min=1,max=255"`
	Name      string `json:"name"`
	Path      string `json:"path" binding:"required,min=1,max=255"`
	Hidden    bool   `json:"hidden"`
	Sort      int    `json:"sort"`
	Icon      string `json:"icon" binding:"omitempty,min=1,max=255"`
	Component string `json:"component" binding:"required,min=1,max=255"`
	Type      string `json:"type" binding:"omitempty,oneof=catalog menu button embedded link"`
	AuthCode  string `json:"authCode" binding:"omitempty,max=100"`
	Status    string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	ApiIDs    []uint `json:"apiIds"`
}

type UpdateMenuRequest struct {
	ParentID  *uint   `json:"parentId" binding:"omitempty,number"`
	Title     *string `json:"title" binding:"omitempty,min=1,max=255"`
	Name      *string `json:"name"`
	Path      *string `json:"path" binding:"omitempty,min=1,max=255"`
	Hidden    *bool   `json:"hidden"`
	Sort      *int    `json:"sort" binding:"omitempty,number"`
	Icon      *string `json:"icon" binding:"omitempty,min=1,max=255"`
	Component *string `json:"component" binding:"omitempty,min=1,max=255"`
	Redirect  *string `json:"redirect" binding:"omitempty,min=1,max=255"`
	Type      *string `json:"type" binding:"omitempty,oneof=catalog menu button embedded link"`
	AuthCode  *string `json:"authCode" binding:"omitempty,max=100"`
	Status    *string `json:"status" binding:"omitempty,oneof=enabled disabled"`
	ApiIDs    []uint  `json:"apiIds"`
}
