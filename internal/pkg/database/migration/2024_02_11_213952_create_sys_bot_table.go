package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
)

type CreateSysBotTable struct {
	model.Base

	Name        string `gorm:"type:varchar(255);not null;default:''"`
	Source      string `gorm:"type:varchar(255);not null;default:''"`
	Description string `gorm:"type:varchar(1024);not null;default:''"`
	Token       string `gorm:"type:varchar(255);not null;default:''"`
	Enabled     int    `gorm:"type:tinyint;not null;default:1;comment:Is enabled"`
}

func (CreateSysBotTable) TableName() string {
	return "sys_bot"
}

func (CreateSysBotTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysBotTable{})
}

func (CreateSysBotTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysBotTable{})
}

func init() {
	migrate.Add("2024_02_11_213952_create_sys_bot_table", CreateSysBotTable{}.Up, CreateSysBotTable{}.Down)
}
