package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type UserInfo struct {
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	UID          string    `json:"uid"`
	CountryCode  string    `json:"countryCode"`
	Nickname     string    `json:"nickname"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Status       int32     `json:"status"`    // Status, 1-enabled, 2-disabled
	KycStatus    int32     `json:"kycStatus"` // KYC status, 0-not verify, 1-pending, 2-verified, 3-failed
	GoogleStatus string    `json:"googleStatus"`
	Pid          int64     `json:"pid"`
	InviteCount  int64     `json:"inviteCount"`
	Age          int32     `json:"age"`
	Gender       string    `json:"gender"`
	Avatar       string    `json:"avatar"`
	PayPassword  bool      `json:"payPassword"`
}

type ListUserRequest struct {
	gormutil.ListOptions
	Keyword     string `form:"keyword"`
	Status      int32  `form:"status"`
	CountryCode string `form:"countryCode"`
}

type ListUserResponse struct {
	Total int64      `json:"total"`
	Data  []UserInfo `json:"data"`
}

type CreateUserRequest struct {
	CountryCode string  `json:"countryCode" binding:"required" example:"us"`
	Nickname    string  `json:"nickname" binding:"required,alphanumunicode" example:"Peter"`
	Username    string  `json:"username" binding:"required,alphanum" example:"peter"`
	Email       *string `json:"email" binding:"omitempty,email" example:"peter@gmail.com"`
	Phone       *string `json:"phone" example:"9999999999"`
	Status      int32   `json:"status" binding:"oneof=1 2" default:"1"` // Status, 1-enabled, 2-disabled
	Pid         string  `json:"pid" example:"88888888"`
	Age         int32   `json:"age" binding:"gte=0,lte=130" example:"0"`
	Gender      string  `json:"gender" binding:"oneof=male female secret" example:"male"` // Gender, male female secret
	Avatar      string  `json:"avatar"`
	Password    string  `json:"password" binding:"required,min=6" example:"123456"`
}

type UpdateUserRequest struct {
	Nickname *string `json:"nickname"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Status   *int32  `json:"status"` // Status, 1-enabled, 2-disabled
	Age      *int32  `json:"age"`
	Gender   *string `json:"gender"`
	Avatar   *string `json:"avatar"`
}

type ResetUserPasswordRequest struct {
	Password string `json:"password" binding:"required,min=6,max=18"`
}
