package syscfg

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type ConfigInfo struct {
	ID          uint64     `json:"id"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	DeletedAt   *time.Time `json:"deletedAt"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Key         string     `json:"key"`
	Value       string     `json:"value"`
	OperatorID  int32      `json:"operatorId"`
}

type ListConfigRequest struct {
	gormutil.ListOptions

	Name        *string `json:"name"`
	Description *string `json:"description"`
	Key         *string `json:"key"`
}

type ListConfigResponse struct {
	Total int64        `json:"total"`
	Data  []ConfigInfo `json:"data"`
}

type CreateConfigRequest struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description"`
	Key         string `json:"key" binding:"required,max=255"`
	Value       string `json:"value"`
	OperatorID  int32  `json:"operatorId"`
}

type UpdateConfigRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Key         *string `json:"key"`
	Value       *string `json:"value"`
	OperatorID  *int32  `json:"operatorId"`
}
