package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/pkg/model"
)

type CreateSysBotChannelTable struct {
	model.Base

	Source    string `gorm:"uniqueIndex:uk_source_channel;type:varchar(255);not null;default:''"`
	ChannelID string `gorm:"uniqueIndex:uk_source_channel;type:varchar(255);not null;default:''"`
	Author    string `gorm:"type:json;not null"`
}

func (CreateSysBotChannelTable) TableName() string {
	return "sys_bot_channel"
}

func (CreateSysBotChannelTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysBotChannelTable{})
}

func (CreateSysBotChannelTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysBotChannelTable{})
}

func init() {
	migrate.Add("2024_02_11_214659_create_sys_bot_channel_table", CreateSysBotChannelTable{}.Up, CreateSysBotChannelTable{}.Down)
}
