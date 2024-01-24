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
	ParentID  int    `json:"parentID" valid:"int"`
	Title     string `json:"title" valid:"stringlength(1|255)"`
	Name      string `json:"name"`
	Path      string `json:"path" valid:"required,stringlength(1|255)"`
	Hidden    bool   `json:"hidden"`
	Sort      int    `json:"sort" valid:"required,int"`
	Icon      string `json:"icon" valid:"stringlength(1|255)"`
	Component string `json:"component" valid:"required,stringlength(1|255)"`
}

type UpdateMenuRequest struct {
	ParentID  uint    `json:"parentID" valid:"int"`
	Title     *string `json:"title" valid:"stringlength(1|255)"`
	Name      *string `json:"name"`
	Path      *string `json:"path" valid:"stringlength(1|255)"`
	Hidden    *bool   `json:"hidden"`
	Sort      *int    `json:"sort" valid:"int"`
	Icon      *string `json:"icon" valid:"stringlength(1|255)"`
	Component *string `json:"component" valid:"stringlength(1|255)"`
	Redirect  string  `json:"redirect" valid:"stringlength(1|255)"`
}
