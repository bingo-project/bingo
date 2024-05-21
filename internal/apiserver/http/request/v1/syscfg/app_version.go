package syscfg

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AppVersionInfo struct {
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

type ListAppVersionRequest struct {
	gormutil.ListOptions

	Name        *string `json:"name"`
	Version     *string `json:"version"`
	Description *string `json:"description"`
	AboutUs     *string `json:"aboutUs"`
	Logo        *string `json:"logo"`
	Enabled     *int32  `json:"enabled"` // Is enabled
}

type ListAppVersionResponse struct {
	Total int64            `json:"total"`
	Data  []AppVersionInfo `json:"data"`
}

type CreateAppVersionRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Version     string `json:"version" binding:"required,max=255"`
	Description string `json:"description" binding:"required,max=1000"`
	AboutUs     string `json:"aboutUs" binding:"required,max=2000"`
	Logo        string `json:"logo" binding:"required,alphanum,max=255"`
	Enabled     int32  `json:"enabled"` // Is enabled
}

type UpdateAppVersionRequest struct {
	Name        *string `json:"name" binding:"omitempty,max=255"`
	Version     *string `json:"version"`
	Description *string `json:"description"`
	AboutUs     *string `json:"aboutUs"`
	Logo        *string `json:"logo"`
	Enabled     *int32  `json:"enabled"` // Is enabled
}
