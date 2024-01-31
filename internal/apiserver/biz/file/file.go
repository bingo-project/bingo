package file

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gookit/goutil/fsutil"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/store"
)

type FileBiz interface {
	Upload(ctx *gin.Context, req *v1.UploadFileRequest) (string, error)
}

type fileBiz struct {
	ds store.IStore
}

var _ FileBiz = (*fileBiz)(nil)

func NewFile(ds store.IStore) *fileBiz {
	return &fileBiz{ds: ds}
}

func (b *fileBiz) Upload(ctx *gin.Context, req *v1.UploadFileRequest) (path string, err error) {
	file, err := req.File.Open()
	if err != nil {
		return
	}

	image := v1.Image{
		Mime: fsutil.ReaderMimeType(file),
		Size: req.File.Size,
	}

	// Validate image
	validate := validator.New()
	err = validate.Struct(image)
	if err != nil {
		return
	}

	// Random filename
	fileName := random.RandString(40) + filepath.Ext(req.File.Filename)

	// Save file
	path = filepath.Join("storage/public/upload", time.Now().Format("2006/01/02"), fileName)
	err = ctx.SaveUploadedFile(req.File, path)
	if err != nil {
		return "", err
	}

	// TODO:: Resize image.

	path = strings.Replace(path, "storage/public/", "storage/", 1)

	return
}
