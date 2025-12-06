package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateSysAuthRoleMenuTable struct {
	RoleName string `gorm:"type:varchar(255);uniqueIndex:uk_role_menu;not null;default:''"`
	MenuID   uint   `gorm:"type:int;uniqueIndex:uk_role_menu;not null;default:0"`
}

func (CreateSysAuthRoleMenuTable) TableName() string {
	return "sys_auth_role_menu"
}

func (CreateSysAuthRoleMenuTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthRoleMenuTable{})
}

func (CreateSysAuthRoleMenuTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthRoleMenuTable{})
}

func init() {
	migrate.Add("2024_05_18_215215_create_sys_auth_role_menu_table", CreateSysAuthRoleMenuTable{}.Up, CreateSysAuthRoleMenuTable{}.Down)
}
