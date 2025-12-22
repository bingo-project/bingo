package model

type RoleM struct {
	Base

	Name        string `gorm:"uniqueIndex:uk_name;type:varchar(255);not null;default:'';comment:名称"`
	Description string `gorm:"type:varchar(255);not null;default:'';comment:描述"`
	Status      string `gorm:"type:varchar(20);not null;default:'enabled';comment:状态(enabled/disabled)"`
	Remark      string `gorm:"type:varchar(255);not null;default:'';comment:备注"`

	// Relation
	Menus []*MenuM `gorm:"many2many:sys_auth_role_menu;foreignKey:name;joinForeignKey:role_name;joinReferences:menu_id"`
}

func (u *RoleM) TableName() string {
	return "sys_auth_role"
}
