package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/apiserver/model"
)

type CreateSysAuthApiTable struct {
	model.Base

	Method      string `gorm:"uniqueIndex:uk_method_path;type:varchar(255);not null;default:''"`
	Path        string `gorm:"uniqueIndex:uk_method_path;type:varchar(255);not null;default:''"`
	Group       string `gorm:"type:varchar(255);not null;default:''"`
	Description string `gorm:"type:varchar(255);not null;default:''"`
}

func (CreateSysAuthApiTable) TableName() string {
	return "sys_auth_api"
}

func (CreateSysAuthApiTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthApiTable{})
}

func (CreateSysAuthApiTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthApiTable{})
}

func init() {
	migrate.Add("2024_05_18_215155_create_sys_auth_api_table", CreateSysAuthApiTable{}.Up, CreateSysAuthApiTable{}.Down)
}
