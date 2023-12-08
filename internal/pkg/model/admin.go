package model

import (
	"gorm.io/gorm"

	"bingo/pkg/auth"
)

type AdminM struct {
	Base

	Username string      `gorm:"type:varchar(255);unique;not null"`
	Password string      `gorm:"type:varchar(255);not null;default:''"`
	Nickname string      `gorm:"type:varchar(255);not null;default:''"`
	Email    string      `gorm:"type:varchar(255);unique"`
	Phone    string      `gorm:"type:varchar(255);unique"`
	Avatar   string      `gorm:"type:varchar(255);not null;default:''"`
	Status   AdminStatus `gorm:"type:tinyint;index;default:1;comment:状态：1正常，2冻结"`
	RoleName string      `gorm:"type:varchar(255);index;not null;default:'';comment:当前角色"`

	// Relation
	Role  *RoleM  `gorm:"foreignKey:role_name;references:name"`
	Roles []RoleM `gorm:"many2many:sys_auth_admin_role;foreignKey:username;joinForeignKey:username;References:name;joinReferences:role_name"`
}

func (u *AdminM) TableName() string {
	return "sys_auth_admin"
}

type AdminStatus uint

const (
	AdminStatusEnabled  AdminStatus = 1
	AdminStatusDisabled AdminStatus = 2
)

func (u *AdminM) BeforeCreate(tx *gorm.DB) (err error) {
	u.Password, err = auth.Encrypt(u.Password)
	if err != nil {
		return
	}

	return nil
}
