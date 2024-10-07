package file

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/bingo-project/component-base/log"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gookit/goutil/fsutil"

	"bingo/internal/admserver/store"
	imageutil "bingo/internal/pkg/util/image"
	"bingo/pkg/api/apiserver/v1"
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
		log.C(ctx).Errorw("upload validator error", "err", err)

		return
	}

	// Random filename
	fileName := random.RandString(40) + filepath.Ext(req.File.Filename)

	// Save file
	path = filepath.Join("storage/public/upload", time.Now().Format("2006/01/02"), fileName)
	err = ctx.SaveUploadedFile(req.File, path)
	if err != nil {
		log.C(ctx).Errorw("SaveUploadedFile error", "err", err)

		return "", err
	}

	// Resize image
	if fsutil.IsImageFile(path) {
		err = imageutil.Resize(path, getResizeRatio(req.File.Size), true, true)
		if err != nil {
			log.C(ctx).Errorw("Resize image error", "err", err)

			return
		}
	}

	path = strings.Replace(path, "storage/public/", "storage/", 1)

	return
}

func getResizeRatio(size int64) float64 {
	// <= 300k
	if size < 1024*100 {
		return 1
	}

	// 300k - 500k
	if size <= 1024*500 {
		return 0.9
	}

	// 500k - 1M
	if size <= 1024*1024 {
		return 0.8
	}

	// 1M - 5M
	if size <= 1024*1024*5 {
		return 0.6
	}

	// 5M - 10M
	if size <= 1024*1024*10 {
		return 0.4
	}

	// 10M - 20M
	if size <= 1024*1024*10 {
		return 0.3
	}

	return 0.2
}
