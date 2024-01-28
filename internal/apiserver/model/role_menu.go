package model

type RoleMenuM struct {
	RoleName string `gorm:"type:varchar(255);uniqueIndex:uk_role_menu;not null;default:''"`
	MenuID   uint   `gorm:"type:int;uniqueIndex:uk_role_menu;not null;default:0"`
}

func (u *RoleMenuM) TableName() string {
	return "sys_auth_role_menu"
}
