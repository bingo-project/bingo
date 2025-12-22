package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
)

type CreateSysAuthRoleTable struct {
	model.Base

	Name        string `gorm:"uniqueIndex:uk_name;type:varchar(255);not null;default:'';comment:名称"`
	Description string `gorm:"type:varchar(255);not null;default:'';comment:描述"`
	Status      string `gorm:"type:varchar(20);not null;default:'enabled';comment:状态(enabled/disabled)"`
	Remark      string `gorm:"type:varchar(255);not null;default:'';comment:备注"`
}

func (CreateSysAuthRoleTable) TableName() string {
	return "sys_auth_role"
}

func (CreateSysAuthRoleTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthRoleTable{})
}

func (CreateSysAuthRoleTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthRoleTable{})
}

func init() {
	migrate.Add("2024_05_18_215212_create_sys_auth_role_table", CreateSysAuthRoleTable{}.Up, CreateSysAuthRoleTable{}.Down)
}
