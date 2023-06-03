package system

import (
	"bingo/internal/pkg/model"
)

type RoleM struct {
	model.Base

	Slug string `gorm:"type:varchar(255);unique;not null;default:'';comment:标识"`
	Name string `gorm:"type:varchar(255);not null;default:'';comment:显示名称"`
}

func (u *RoleM) TableName() string {
	return "sys_auth_role"
}
