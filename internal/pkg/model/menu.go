package model

type MenuM struct {
	Base

	ParentID  int    `gorm:"index:idx_parent;type:int;not null;default:0"`
	Title     string `gorm:"type:varchar(255);not null;default:''"`
	Name      string `gorm:"type:varchar(255);not null;default:''"`
	Path      string `gorm:"index:idx_path;type:varchar(255);not null;default:''"`
	Hidden    int    `gorm:"type:tinyint;not null;default:0;comment:Is Hidden"`
	Sort      int    `gorm:"type:int;not null;default:0"`
	Icon      string `gorm:"type:varchar(255);not null;default:''"`
	Component string `gorm:"type:varchar(255);not null;default:''"`
}

func (u *MenuM) TableName() string {
	return "sys_auth_menu"
}
