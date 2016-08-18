package imageServer

import "gopkg.in/h2non/bimg.v1"

func resize(buffer []byte, params ImageParams) ([]byte, error) {
	//todo: validate etc

	return bimg.Resize(buffer, params.toBimgOptions())
}
