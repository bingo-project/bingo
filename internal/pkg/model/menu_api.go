// ABOUTME: Join table model for menu-api associations.
// ABOUTME: Links menus to APIs for permission-based access control.

package model

type MenuApiM struct {
	MenuID uint `gorm:"uniqueIndex:uk_menu_api;not null"`
	ApiID  uint `gorm:"uniqueIndex:uk_menu_api;not null"`
}

func (MenuApiM) TableName() string {
	return "sys_auth_menu_api"
}
