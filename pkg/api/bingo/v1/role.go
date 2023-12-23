package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type RoleInfo struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name        string `json:"name"`
	Description string `json:"description"`
}

type ListRoleRequest struct {
	gormutil.ListOptions
}

type CreateRoleRequest struct {
	Name        string `json:"name" valid:"required,alphanum,stringlength(1|20)"`
	Description string `json:"description" valid:"required,stringlength(1|255)"`
}

type UpdateRoleRequest struct {
	Name        *string `json:"name" valid:"alphanum,stringlength(1|20)"`
	Description *string `json:"description" valid:"stringlength(1|255)"`
}

type SetApisRequest struct {
	ApiIDs []uint `json:"apiIDs" valid:"required"`
}

type SetMenusRequest struct {
	MenuIDs []uint `json:"menuIDs" valid:"required"`
}
