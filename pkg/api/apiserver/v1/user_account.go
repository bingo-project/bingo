package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type UserAccountInfo struct {
	ID        uint64     `json:"id"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`

	UID       string `json:"uid"`
	Provider  string `json:"provider"`
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar"`
}

type ListUserAccountRequest struct {
	gormutil.ListOptions

	UID       *string `json:"uid"`
	Provider  *string `json:"provider"`
	AccountID *string `json:"accountId"`
	Username  *string `json:"username"`
	Nickname  *string `json:"nickname"`
	Email     *string `json:"email"`
	Bio       *string `json:"bio"`
	Avatar    *string `json:"avatar"`
}

type ListUserAccountResponse struct {
	Total int64             `json:"total"`
	Data  []UserAccountInfo `json:"data"`
}

type CreateUserAccountRequest struct {
	UID       string `json:"uid"`
	Provider  string `json:"provider"`
	AccountID string `json:"accountId"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	Avatar    string `json:"avatar"`
}

type UpdateUserAccountRequest struct {
	UID       *string `json:"uid"`
	Provider  *string `json:"provider"`
	AccountID *string `json:"accountId"`
	Username  *string `json:"username"`
	Nickname  *string `json:"nickname"`
	Email     *string `json:"email"`
	Bio       *string `json:"bio"`
	Avatar    *string `json:"avatar"`
}
