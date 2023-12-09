package model

type RoleM struct {
	Base

	Name        string `gorm:"uniqueIndex:uk_name;type:varchar(255);not null;default:'';comment:名称"`
	Description string `gorm:"type:varchar(255);not null;default:'';comment:描述"`
}

func (u *RoleM) TableName() string {
	return "sys_auth_role"
}
