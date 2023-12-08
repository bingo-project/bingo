package v1

import (
	"time"
)

type ApiInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Method      string `json:"method"`
	Path        string `json:"path"`
	Group       string `json:"group"`
	Description string `json:"description"`
}

type ListApiRequest struct {
	ListRequest
}

type CreateApiRequest struct {
	Method      string `json:"method" valid:"required,alphanum,stringlength(1|255)"`
	Path        string `json:"path" valid:"required,stringlength(1|255)"`
	Group       string `json:"group" valid:"required,stringlength(1|255)"`
	Description string `json:"description" valid:"required,stringlength(1|255)"`
}

type UpdateApiRequest struct {
	Method      *string `json:"method" valid:"alphanum,stringlength(1|255)"`
	Path        *string `json:"path" valid:"stringlength(1|255)"`
	Group       *string `json:"group" valid:"stringlength(1|255)"`
	Description *string `json:"description" valid:"stringlength(1|255)"`
}

type GetApiIDsResponse []uint
