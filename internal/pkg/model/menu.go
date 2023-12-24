package model

type MenuM struct {
	Base

	ParentID  uint   `gorm:"index:idx_parent;type:int;not null;default:0"`
	Name      string `gorm:"type:varchar(255);not null;default:''"`
	Path      string `gorm:"index:idx_path;type:varchar(255);not null;default:''"`
	Sort      int    `gorm:"type:int;not null;default:0"`
	Title     string `gorm:"type:varchar(255);not null;default:''"`
	Icon      string `gorm:"type:varchar(255);not null;default:''"`
	Hidden    bool   `gorm:"type:tinyint;not null;default:0;comment:Is Hidden"`
	Component string `gorm:"type:varchar(255);not null;default:''"`
	Redirect  string `gorm:"type:varchar(255);not null;default:''"`

	// Relations
	Children []*MenuM `gorm:"foreignKey:parent_id"`
}

func (u *MenuM) TableName() string {
	return "sys_auth_menu"
}
