package bot

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type ChannelInfo struct {
	ID        uint64     `json:"id"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Source    string     `json:"source"`
	ChannelID string     `json:"channelId"`
	Author    string     `json:"author"`
}

type ListChannelRequest struct {
	gormutil.ListOptions

	Source    *string `json:"source"`
	ChannelID *string `json:"channelId"`
	Author    *string `json:"author"`
}

type ListChannelResponse struct {
	Total int64         `json:"total"`
	Data  []ChannelInfo `json:"data"`
}

type CreateChannelRequest struct {
	Source    string `json:"source"`
	ChannelID string `json:"channelId"`
	Author    string `json:"author"`
}

type UpdateChannelRequest struct {
	Source    *string `json:"source"`
	ChannelID *string `json:"channelId"`
	Author    *string `json:"author"`
}
