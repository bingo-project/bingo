package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApiKey struct {
	Base

	UID         string                      `gorm:"column:uid;type:varchar(255);not null;index:idx_uid,priority:1" json:"uid"`
	AppID       string                      `gorm:"column:app_id;type:varchar(255);not null;index:idx_app_id,priority:1" json:"appId"`
	Name        string                      `gorm:"column:name;type:varchar(255);not null" json:"name"`
	AccessKey   string                      `gorm:"column:access_key;type:varchar(255);not null" json:"accessKey"`
	SecretKey   string                      `gorm:"column:secret_key;type:varchar(255);not null" json:"secretKey"`
	Status      ApiKeyStatus                `gorm:"column:status;type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled" json:"status"` // Status, 1-enabled, 2-disabled
	ACL         datatypes.JSONSlice[string] `gorm:"column:acl;type:json" json:"acl"`
	Description string                      `gorm:"column:description;type:varchar(1000);not null" json:"description"`
	ExpiredAt   *time.Time                  `gorm:"column:expired_at;type:datetime(3);index:idx_expired_at,priority:1" json:"expiredAt"`
}

func (*ApiKey) TableName() string {
	return "app_api_key"
}

// ApiKeyStatus 1-enabled, 2-disabled.
type ApiKeyStatus int

const (
	ApiKeyStatusEnabled  ApiKeyStatus = 1
	ApiKeyStatusDisabled ApiKeyStatus = 2
)

func (m *ApiKey) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate ak sk
	if m.AccessKey == "" {
		m.AccessKey = uuid.New().String()
	}
	if m.SecretKey == "" {
		m.SecretKey = uuid.New().String()
	}

	return nil
}
