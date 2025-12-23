// ABOUTME: Migration to create sys_auth_menu_api join table.
// ABOUTME: Links menus to APIs for permission-based access control.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateSysAuthMenuApiTable struct {
	MenuID uint `gorm:"uniqueIndex:uk_menu_api;not null"`
	ApiID  uint `gorm:"uniqueIndex:uk_menu_api;not null"`
}

func (CreateSysAuthMenuApiTable) TableName() string {
	return "sys_auth_menu_api"
}

func (CreateSysAuthMenuApiTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthMenuApiTable{})
}

func (CreateSysAuthMenuApiTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthMenuApiTable{})
}

func init() {
	migrate.Add("2025_12_23_191335_create_sys_auth_menu_api_table", CreateSysAuthMenuApiTable{}.Up, CreateSysAuthMenuApiTable{}.Down)
}
