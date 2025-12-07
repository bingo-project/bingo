package file

import (
	"github.com/gin-gonic/gin"

	"bingo/internal/admserver/biz"
	"bingo/internal/pkg/auth"
	"bingo/internal/pkg/core"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
)

type FileHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewFileHandler(ds store.IStore, a *auth.Authorizer) *FileHandler {
	return &FileHandler{a: a, b: biz.NewBiz(ds)}
}

// Upload
// @Summary    Upload file
// @Security   Bearer
// @Tags       File
// @Accept     multipart/form-data
// @Produce    json
// @Param      file     formData    file    true  "File"
// @Success	   200		{object}	string
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/file/upload [POST].
func (ctrl *FileHandler) Upload(c *gin.Context) {
	log.C(c).Infow("Upload file function called")

	var req v1.UploadFileRequest
	if err := c.ShouldBind(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument)

		return
	}

	// Create file
	resp, err := ctrl.b.Files().Upload(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}
