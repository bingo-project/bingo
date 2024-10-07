package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type ApiInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Method      string `json:"method"`
	Path        string `json:"path"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type ListApiRequest struct {
	gormutil.ListOptions

	Method string `form:"method"`
	Path   string `form:"path"`
	Group  string `form:"group"`
}

type ListApiResponse struct {
	Total int64     `json:"total"`
	Data  []ApiInfo `json:"data"`
}

type CreateApiRequest struct {
	Method      string `json:"method" binding:"required,alphanum,min=1,max=255"`
	Path        string `json:"path" binding:"required,min=1,max=255"`
	Group       string `json:"group" binding:"required,min=1,max=255"`
	Description string `json:"description" binding:"required,min=1,max=255"`
}

type UpdateApiRequest struct {
	Method      *string `json:"method" binding:"alphanum,min=1,max=255"`
	Path        *string `json:"path" binding:"min=1,max=255"`
	Group       *string `json:"group" binding:"min=1,max=255"`
	Description *string `json:"description" binding:"min=1,max=255"`
}

type GetApiIDsResponse []uint

type GroupApiResponse struct {
	Key   string    `json:"key"`
	Group []ApiInfo `json:"children"`
}
