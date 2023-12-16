package v1

import (
	"time"
)

type MenuInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ParentID  int    `json:"parentID"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Sort      int    `json:"sort"`
	Component string `json:"component"`
	Redirect  string `json:"redirect"`
	Meta      Meta   `json:"meta"`

	Children []MenuInfo `json:"children,omitempty"`
}

type Meta struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Hidden bool   `json:"hideMenu"`
}

type ListMenuRequest struct {
	ListRequest
}

type CreateMenuRequest struct {
	ParentID  int    `json:"parentID" valid:"int"`
	Title     string `json:"title" valid:"stringlength(1|255)"`
	Name      string `json:"name" `
	Path      string `json:"path" valid:"required,stringlength(1|255)"`
	Hidden    int    `json:"hidden" valid:"int"`
	Sort      int    `json:"sort" valid:"required,int"`
	Icon      string `json:"icon" valid:"required,stringlength(1|255)"`
	Component string `json:"component" valid:"required,stringlength(1|255)"`
}

type UpdateMenuRequest struct {
	ParentID  *uint   `json:"parentID" valid:"int"`
	Title     *string `json:"title" valid:"stringlength(1|255)"`
	Name      *string `json:"name"`
	Path      *string `json:"path" valid:"stringlength(1|255)"`
	Hidden    *bool   `json:"hidden" valid:"bool"`
	Sort      *int    `json:"sort" valid:"int"`
	Icon      *string `json:"icon" valid:"stringlength(1|255)"`
	Component *string `json:"component" valid:"stringlength(1|255)"`
}
