package bot

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AdminInfo struct {
	ID        uint64     `json:"id"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt"`
	Source    string     `json:"source"`
	UserID    string     `json:"userId"`
}

type ListAdminRequest struct {
	gormutil.ListOptions

	Source *string `json:"source"`
	UserID *string `json:"userId"`
}

type ListAdminResponse struct {
	Total int64       `json:"total"`
	Data  []AdminInfo `json:"data"`
}

type CreateAdminRequest struct {
	Source string `json:"source"`
	UserID string `json:"userId"`
}

type UpdateAdminRequest struct {
	Source *string `json:"source"`
	UserID *string `json:"userId"`
}
