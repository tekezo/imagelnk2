package core

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/h2non/bimg"
)

func SaveImageCache(data []byte) (string, error) {
	img := bimg.NewImage(data)

	//
	// Create thumbnail
	//

	imgSize, err := img.Size()
	if err != nil {
		return "", err
	}
	if imgSize.Width == 0 || imgSize.Height == 0 {
		return "", fmt.Errorf("image width or height is zero")
	}

	thumbnailWidth := float64(imgSize.Width)
	thumbnailHeight := float64(imgSize.Height)
	ratio := float64(imgSize.Width) / float64(imgSize.Height)
	// Ensure width < 160
	if thumbnailWidth > 160 {
		thumbnailWidth = 160
		thumbnailHeight = thumbnailWidth / ratio
	}
	// Ensure height < 160*3
	if thumbnailHeight > 160*3 {
		// If thumbnailWidth == 160, thumbnailHeight == 920, newWidth = 80, newHeight = 480
		thumbnailWidth = thumbnailWidth * thumbnailHeight / (160 * 3)
		thumbnailHeight = 160 * 3
	}

	thumbnail, err := img.ForceResize(int(thumbnailWidth), int(thumbnailHeight))
	if err != nil {
		return "", err
	}

	//
	// Save image
	//

	timeHex := fmt.Sprintf("%x", time.Now().Unix())
	randomString := fmt.Sprintf("%x", rand.New(rand.NewSource(time.Now().UnixNano())).Int())

	outputPath := filepath.Join(
		Config.ImageCacheDirectory,
		"original",
		randomString[0:2],
		fmt.Sprintf("%s-%s.%s", timeHex, randomString, img.Type()),
	)

	err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		return "", err
	}

	os.WriteFile(outputPath, data, os.ModePerm)

	//
	// Save thumbnail
	//

	thumbnailPath := filepath.Join(
		Config.ImageCacheDirectory,
		"thumbnail",
		randomString[0:2],
		fmt.Sprintf("%s-%s.jpg", timeHex, randomString),
	)

	err = os.MkdirAll(filepath.Dir(thumbnailPath), os.ModePerm)
	if err != nil {
		return "", err
	}

	bimg.Write(thumbnailPath, thumbnail)

	return fmt.Sprintf("%s/original/%s/%s",
		Config.ImageCacheURL,
		randomString[0:2],
		filepath.Base(outputPath),
	), nil
}
