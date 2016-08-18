package imageServer

import (
	"fmt"

	"gopkg.in/h2non/bimg.v1"
)

func resize(buffer []byte, params imageParams) (image, error) {
	//todo: validate etc

	img := &image{}

	b, err := bimg.Resize(buffer, params.toBimgOptions())
	if err != nil {
		return *img, err
	}

	t := bimg.DetermineImageTypeName(b)
	img.Contents = b
	img.Mime = fmt.Sprintf("image/%s", t)

	return *img, nil
}
