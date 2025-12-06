package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateSysAuthAdminRoleTable struct {
	Username string `gorm:"type:varchar(255);uniqueIndex:uk_username_role;not null;default:''"`
	RoleName string `gorm:"type:varchar(255);uniqueIndex:uk_username_role;not null;default:''"`
}

func (CreateSysAuthAdminRoleTable) TableName() string {
	return "sys_auth_admin_role"
}

func (CreateSysAuthAdminRoleTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthAdminRoleTable{})
}

func (CreateSysAuthAdminRoleTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthAdminRoleTable{})
}

func init() {
	migrate.Add("2024_05_18_215148_create_sys_auth_admin_role_table", CreateSysAuthAdminRoleTable{}.Up, CreateSysAuthAdminRoleTable{}.Down)
}
