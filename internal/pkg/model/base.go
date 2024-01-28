package model

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);index:idx_updated_at;autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"index:idx_deleted_at" json:"deletedAt"`
}

type Model struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3);index;autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}
