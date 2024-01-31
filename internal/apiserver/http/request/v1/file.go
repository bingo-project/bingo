package v1

import "mime/multipart"

type UploadFileRequest struct {
	File *multipart.FileHeader `form:"file" binding:"required"` // File
}

// Image struct for image information and storage location.
type Image struct {
	Mime string `validate:"required,oneof=image/png image/jpg image/jpeg"`
	Size int64  `validate:"required,gt=0,lte=5242880"`
}
