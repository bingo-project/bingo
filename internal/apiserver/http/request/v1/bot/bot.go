package bot

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type BotInfo struct {
	ID          uint64     `json:"id"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	Name        string     `json:"name"`
	Source      string     `json:"source"`
	Description string     `json:"description"`
	Token       string     `json:"token"`
	Enabled     *int32     `json:"enabled"` // Is enabled
}

type ListBotRequest struct {
	gormutil.ListOptions

	Name        *string `json:"name"`
	Source      *string `json:"source"`
	Description *string `json:"description"`
	Token       *string `json:"token"`
	Enabled     **int32 `json:"enabled"` // Is enabled
}

type ListBotResponse struct {
	Total int64     `json:"total"`
	Data  []BotInfo `json:"data"`
}

type CreateBotRequest struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Token       string `json:"token"`
	Enabled     *int32 `json:"enabled"` // Is enabled
}

type UpdateBotRequest struct {
	Name        *string `json:"name"`
	Source      *string `json:"source"`
	Description *string `json:"description"`
	Token       *string `json:"token"`
	Enabled     *int32  `json:"enabled"` // Is enabled
}
