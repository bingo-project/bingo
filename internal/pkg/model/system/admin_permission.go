package system

import "bingo/internal/pkg/model"

type PermissionM struct {
	model.Base

	Method      string `gorm:"type:varchar(255);not null;default:''"`
	Path        string `gorm:"type:varchar(255);not null;default:''"`
	Group       string `gorm:"type:varchar(255);not null;default:''"`
	Description string `gorm:"type:varchar(255);not null;default:''"`
}

func (u *PermissionM) TableName() string {
	return "sys_auth_permission"
}
