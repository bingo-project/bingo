package model

import (
	"gorm.io/gorm"

	"bingo/pkg/auth"
)

// UserM 是数据库中 user 记录 struct 格式的映射.
type UserM struct {
	Base

	Username string `gorm:"column:username;not null"`
	Password string `gorm:"column:password;not null"`
	Nickname string `gorm:"column:nickname"`
	Email    string `gorm:"column:email"`
	Phone    string `gorm:"column:phone"`
}

func (u *UserM) TableName() string {
	return "user"
}

// BeforeCreate 在创建数据库记录之前加密明文密码.
func (u *UserM) BeforeCreate(tx *gorm.DB) (err error) {
	// Encrypt the user password.
	u.Password, err = auth.Encrypt(u.Password)
	if err != nil {
		return err
	}

	return nil
}
