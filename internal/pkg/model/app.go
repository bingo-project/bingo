package model

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

type App struct {
	Base

	UID         string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid,priority:1" json:"uid"`
	AppID       string    `gorm:"column:app_id;type:varchar(255);not null;uniqueIndex:uk_app_id,priority:1" json:"appId"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Status      AppStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled" json:"status"` // Status, 1-enabled, 2-disabled
	Description string    `gorm:"column:description;type:varchar(1000);not null" json:"description"`
	Logo        string    `gorm:"column:logo;type:varchar(1000);not null" json:"logo"`
}

func (*App) TableName() string {
	return "app"
}

// AppStatus 1-enabled, 2-disabled.
type AppStatus int

const (
	AppStatusEnabled  AppStatus = 1
	AppStatusDisabled AppStatus = 2
)

func (u *App) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate App id
	if u.AppID == "" {
		u.AppID = facade.Snowflake.Generate().String()
	}

	return nil
}
