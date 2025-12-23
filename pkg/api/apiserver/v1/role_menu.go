package v1

import "github.com/bingo-project/component-base/util/gormutil"

type RoleMenuInfo struct {
	RoleName string `json:"roleName"`
	MenuID   uint   `json:"menuId"`
}

type ListRoleMenuRequest struct {
	gormutil.ListOptions
}

type GetMenuIDsResponse []uint
