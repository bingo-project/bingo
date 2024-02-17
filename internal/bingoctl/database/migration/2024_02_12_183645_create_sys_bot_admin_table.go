package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/apiserver/model"
)

type CreateSysBotAdminTable struct {
	model.Base

	Source string `gorm:"uniqueIndex:uk_source_user;type:varchar(255);not null;default:''"`
	UserID string `gorm:"uniqueIndex:uk_source_user;type:varchar(255);not null;default:''"`
	Role   string `gorm:"type:varchar(255);not null;default:''"`
}

func (CreateSysBotAdminTable) TableName() string {
	return "sys_bot_admin"
}

func (CreateSysBotAdminTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysBotAdminTable{})
}

func (CreateSysBotAdminTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysBotAdminTable{})
}

func init() {
	migrate.Add("2024_02_12_183645_create_sys_bot_admin_table", CreateSysBotAdminTable{}.Up, CreateSysBotAdminTable{}.Down)
}
