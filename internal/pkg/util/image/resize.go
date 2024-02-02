package image

import (
	"errors"
	"math"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/gookit/goutil/fsutil"
)

const (
	OriginalDir  = "original/"
	ThumbnailDir = "thumbnail/"
)

var (
	ErrOpenImage         = "failed to open image: "
	ErrSaveOriginalImage = "failed to save original image: "
	ErrGenerateThumbnail = "failed to generate thumbnail: "
	ErrResizeImage       = "failed to resize image: "
)

func Resize(path string, ratio float64, original, thumbnail bool) error {
	// Get dir & filename
	dir, filename := filepath.Split(path)

	// Create original & thumbnail directory
	originalDir, thumbnailDir, err := createOriginalAndThumbnailDirectory(dir)
	if err != nil {
		return err
	}

	// Open image
	src, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return errors.New(ErrOpenImage + err.Error())
	}

	// Resize ratio
	width := getResizeWidth(float64(src.Bounds().Size().X), ratio)

	// Save original image
	if original {
		err = fsutil.CopyFile(path, originalDir+filename)
		if err != nil {
			return errors.New(ErrSaveOriginalImage + err.Error())
		}
	}

	// Save thumbnail
	if thumbnail {
		thumb := imaging.Resize(src, 200, 0, imaging.Lanczos)
		err = imaging.Save(thumb, thumbnailDir+filename)
		if err != nil {
			return errors.New(ErrGenerateThumbnail + err.Error())
		}
	}

	// Resize
	resize := imaging.Resize(src, int(width), 0, imaging.Lanczos)
	err = imaging.Save(resize, path)
	if err != nil {
		return errors.New(ErrResizeImage + err.Error())
	}

	return nil
}

func getResizeWidth(width, ratio float64) float64 {
	if ratio <= 0 {
		ratio = 1
	}

	width *= ratio

	return math.Max(width, 300)
}

func createOriginalAndThumbnailDirectory(directory string) (original, thumbnail string, err error) {
	original = directory + OriginalDir
	err = os.MkdirAll(original, 0755)
	if err != nil {
		return
	}

	thumbnail = directory + ThumbnailDir
	err = os.MkdirAll(thumbnail, 0755)

	return
}
