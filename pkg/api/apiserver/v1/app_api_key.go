package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type ApiKeyInfo struct {
	ID          uint64     `json:"id"`
	CreatedAt   *time.Time `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt"`
	UID         string     `json:"uid"`
	AppID       string     `json:"appId"`
	Name        string     `json:"name"`
	AccessKey   string     `json:"accessKey"`
	SecretKey   string     `json:"secretKey"`
	Status      int32      `json:"status"` // Status, 1-enabled, 2-disabled
	ACL         []string   `json:"acl"`
	Description string     `json:"description"`
	ExpiredAt   *time.Time `json:"expiredAt"`
}

type ListApiKeyRequest struct {
	gormutil.ListOptions

	UID       *string `json:"uid"`
	AppID     *string `json:"appId"`
	Name      *string `json:"name"`
	AccessKey *string `json:"accessKey"`
	Status    *int32  `json:"status"` // Status, 1-enabled, 2-disabled
}

type ListApiKeyResponse struct {
	Total int64        `json:"total"`
	Data  []ApiKeyInfo `json:"data"`
}

type CreateApiKeyRequest struct {
	AppID       string   `json:"appId"`
	Name        string   `json:"name" binding:"required,alphanum,min=2,max=255"`
	Status      int32    `json:"status" binding:"required,oneof=1 2"` // Status, 1-enabled, 2-disabled
	ACL         []string `json:"acl" binding:"omitempty,dive,ip|cidr"`
	Description string   `json:"description"`
	ExpiredAt   string   `json:"expiredAt" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}

type UpdateApiKeyRequest struct {
	Name        *string  `json:"name"`
	Status      *int32   `json:"status" binding:"omitempty,oneof=1 2"` // Status, 1-enabled, 2-disabled
	ACL         []string `json:"acl" binding:"omitempty,dive,ip|cidr"`
	Description *string  `json:"description"`
	ExpiredAt   *string  `json:"expiredAt" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}
