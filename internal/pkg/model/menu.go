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
	Type      string `gorm:"type:varchar(20);not null;default:'menu'"`
	AuthCode  string `gorm:"type:varchar(100);not null;default:''"`
	Status    string `gorm:"type:varchar(20);not null;default:'enabled'"`

	// Relations
	Children []*MenuM `gorm:"foreignKey:parent_id"`
	Apis     []*ApiM  `gorm:"many2many:sys_auth_menu_api;joinForeignKey:menu_id;joinReferences:api_id"`
}

func (u *MenuM) TableName() string {
	return "sys_auth_menu"
}
