package v1

type RoleMenuInfo struct {
	RoleName string `json:"roleName"`
	MenuID   uint   `json:"menuID"`
}

type ListRoleMenuRequest struct {
	ListRequest
}

type GetMenuIDsResponse []uint
