package model

import (
	"time"
)

type Base struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"type:DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"type:DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index:idx_deleted;type:DATETIME NULL" json:"deletedAt"`
}
