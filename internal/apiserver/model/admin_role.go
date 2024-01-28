package model

type AdminRoleM struct {
	Username string `gorm:"type:varchar(255);uniqueIndex:uk_username_role;not null;default:''"`
	RoleName string `gorm:"type:varchar(255);uniqueIndex:uk_username_role;not null;default:''"`
}

func (u *AdminRoleM) TableName() string {
	return "sys_auth_admin_role"
}
