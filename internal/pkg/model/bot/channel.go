package bot

import (
	"bingo/internal/pkg/model"
)

type Channel struct {
	model.Base

	Source    Source `gorm:"column:source;type:varchar(255);not null;uniqueIndex:uk_source_channel,priority:1" json:"source"`
	ChannelID string `gorm:"column:channel_id;type:varchar(255);not null;uniqueIndex:uk_source_channel,priority:2" json:"channelId"`
	Author    string `gorm:"column:author;type:json;not null;default:'{}'" json:"author"`
}

func (*Channel) TableName() string {
	return "sys_bot_channel"
}
