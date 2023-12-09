package v1

import (
	"time"
)

type MenuInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	ParentID  int    `json:"parentID"`
	Title     string `json:"title"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	Hidden    int    `json:"hidden"`
	Sort      int    `json:"sort"`
	Icon      string `json:"icon"`
	Component string `json:"component"`
}

type ListMenuRequest struct {
	ListRequest
}

type CreateMenuRequest struct {
	ParentID  int    `json:"parentID" valid:"alphanum,stringlength(1|255)"`
	Title     string `json:"title" valid:"stringlength(1|255)"`
	Name      string `json:"name" `
	Path      string `json:"path" valid:"required,stringlength(1|255)"`
	Hidden    int    `json:"hidden" valid:"int"`
	Sort      int    `json:"sort" valid:"required,int"`
	Icon      string `json:"icon" valid:"required,stringlength(1|255)"`
	Component string `json:"component" valid:"required,stringlength(1|255)"`
}

type UpdateMenuRequest struct {
	ParentID  *int    `json:"parentID" valid:"int,stringlength(1|255)"`
	Title     *string `json:"title" valid:"stringlength(1|255)"`
	Name      *string `json:"name"`
	Path      *string `json:"path" valid:"stringlength(1|255)"`
	Hidden    *int    `json:"hidden" valid:"int"`
	Sort      *int    `json:"sort" valid:"int"`
	Icon      *string `json:"icon" valid:"stringlength(1|255)"`
	Component *string `json:"component" valid:"stringlength(1|255)"`
}
