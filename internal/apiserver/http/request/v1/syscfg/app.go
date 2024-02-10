package syscfg

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AppInfo struct {
	ID          uint64     `json:"id"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	AboutUs     string     `json:"aboutUs"`
	Logo        string     `json:"logo"`
	Enabled     int32      `json:"enabled"` // Is enabled
}

type ListAppRequest struct {
	gormutil.ListOptions

	Name        *string `json:"name"`
	Version     *string `json:"version"`
	Description *string `json:"description"`
	AboutUs     *string `json:"aboutUs"`
	Logo        *string `json:"logo"`
	Enabled     *int32  `json:"enabled"` // Is enabled
}

type ListAppResponse struct {
	Total int64     `json:"total"`
	Data  []AppInfo `json:"data"`
}

type CreateAppRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Version     string `json:"version" binding:"required,max=255"`
	Description string `json:"description" binding:"required,max=1000"`
	AboutUs     string `json:"aboutUs" binding:"required,max=2000"`
	Logo        string `json:"logo" binding:"required,alphanum,max=255"`
	Enabled     int32  `json:"enabled"` // Is enabled
}

type UpdateAppRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Version     *string `json:"version"`
	Description *string `json:"description"`
	AboutUs     *string `json:"aboutUs"`
	Logo        *string `json:"logo"`
	Enabled     *int32  `json:"enabled"` // Is enabled
}
