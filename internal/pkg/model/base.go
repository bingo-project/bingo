package model

import (
	"time"
)

type Base struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	CreatedAt time.Time  `gorm:"type:DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"type:DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index;type:DATETIME NULL" json:"deletedAt"`
}
