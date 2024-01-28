package model

import (
	"gorm.io/gorm"

	"bingo/pkg/auth"
)

type UserM struct {
	Base

	UID          string  `gorm:"column:uid;type:varchar(255);uniqueIndex:uk_uid,priority:1" json:"uid"`
	CountryCode  string  `gorm:"column:country_code;type:varchar(255);not null" json:"countryCode"`
	Nickname     string  `gorm:"column:nickname;type:varchar(255);not null" json:"nickname"`
	Username     string  `gorm:"column:username;type:varchar(255);uniqueIndex:uk_username,priority:1" json:"username"`
	Email        string  `gorm:"column:email;type:varchar(255);uniqueIndex:uk_email,priority:1" json:"email"`
	Phone        string  `gorm:"column:phone;type:varchar(255);uniqueIndex:uk_phone,priority:1" json:"phone"`
	Status       int32   `gorm:"column:status;type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled" json:"status"`                          // Status, 1-enabled, 2-disabled
	KycStatus    int32   `gorm:"column:kyc_status;type:tinyint;not null;comment:KYC status, 0-not verify, 1-pending, 2-verified, 3-failed" json:"kycStatus"` // KYC status, 0-not verify, 1-pending, 2-verified, 3-failed
	GoogleKey    string  `gorm:"column:google_key;type:varchar(255);not null" json:"googleKey"`
	GoogleStatus *string `gorm:"column:google_status;type:enum('unbind','disabled','enabled');not null;default:unbind" json:"googleStatus"`
	Pid          int64   `gorm:"column:pid;type:bigint;not null;index:idx_pid,priority:1" json:"pid"`
	InviteCount  int64   `gorm:"column:invite_count;type:bigint;not null" json:"inviteCount"`
	Depth        int64   `gorm:"column:depth;type:bigint;not null" json:"depth"`
	Age          int32   `gorm:"column:age;type:tinyint;not null" json:"age"`
	Gender       *string `gorm:"column:gender;type:enum('secret','male','female');not null;default:secret" json:"gender"`
	Avatar       string  `gorm:"column:avatar;type:varchar(255);not null" json:"avatar"`
	Password     string  `gorm:"column:password;type:varchar(255);not null" json:"password"`
	PayPassword  string  `gorm:"column:pay_password;type:varchar(255);not null" json:"payPassword"`
}

func (*UserM) TableName() string {
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
