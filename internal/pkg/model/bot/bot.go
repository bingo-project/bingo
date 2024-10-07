package bot

import (
	"bingo/internal/pkg/model"
)

type Bot struct {
	model.Base

	Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Source      Source `gorm:"column:source;type:varchar(255);not null" json:"source"`
	Description string `gorm:"column:description;type:varchar(1024);not null" json:"description"`
	Token       string `gorm:"column:token;type:varchar(255);not null" json:"token"`
	Enabled     int32  `gorm:"column:enabled;type:tinyint;not null;default:1;comment:Is enabled" json:"enabled"` // Is enabled
}

func (*Bot) TableName() string {
	return "sys_bot"
}

type Source string

const (
	SourceTelegram Source = "telegram"
	SourceDiscord  Source = "discord"
)
