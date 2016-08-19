package imageServer

import (
	"io/ioutil"
	"net/http"

	"fmt"

	"gopkg.in/h2non/bimg.v1"
)

type Image struct {
	Contents []byte
	Metadata bimg.ImageMetadata
}

func (i Image) GetMimeType() string {
	t := i.Metadata.Type

	return fmt.Sprintf("image/%s", t)
}

func NewImage(buf []byte) (Image, error) {
	meta, err := bimg.Metadata(buf)
	if err != nil {
		return Image{}, err
	}

	return Image{buf, meta}, nil
}

func downloadImage(url string) (Image, error) {
	r, err := http.Get(url)
	if err != nil {
		return Image{}, err
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return Image{}, err
	}

	return NewImage(b)
}
