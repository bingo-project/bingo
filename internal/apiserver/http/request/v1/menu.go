package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type MenuInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ParentID  int    `json:"parentID"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Sort      int    `json:"sort"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	Hidden    bool   `json:"hidden"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`

	Children []MenuInfo `json:"children,omitempty"`
}

type ListMenuRequest struct {
	gormutil.ListOptions
}

type ListMenuResponse struct {
	Total int64      `json:"total"`
	Data  []MenuInfo `json:"data"`
}

type CreateMenuRequest struct {
	ParentID  int    `json:"parentID" binding:"int"`
	Title     string `json:"title" binding:"min=1,max=255"`
	Name      string `json:"name"`
	Path      string `json:"path" binding:"required,min=1,max=255"`
	Hidden    bool   `json:"hidden"`
	Sort      int    `json:"sort" binding:"required,int"`
	Icon      string `json:"icon" binding:"min=1,max=255"`
	Component string `json:"component" binding:"required,min=1,max=255"`
}

type UpdateMenuRequest struct {
	ParentID  uint    `json:"parentID" binding:"int"`
	Title     *string `json:"title" binding:"min=1,max=255"`
	Name      *string `json:"name"`
	Path      *string `json:"path" binding:"min=1,max=255"`
	Hidden    *bool   `json:"hidden"`
	Sort      *int    `json:"sort" binding:"int"`
	Icon      *string `json:"icon" binding:"min=1,max=255"`
	Component *string `json:"component" binding:"min=1,max=255"`
	Redirect  string  `json:"redirect" binding:"min=1,max=255"`
}
