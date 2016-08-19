package imageServer

import "gopkg.in/h2non/bimg.v1"

func resize(buffer []byte, params ImageParams) (Image, error) {
	//todo: validate etc

	opts := params.toBimgOptions()
	b, err := bimg.Resize(buffer, opts)
	if err != nil {
		return Image{}, err
	}

	return NewImage(b)
}
