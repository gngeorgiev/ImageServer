package imageServer

import (
	"io/ioutil"
	"net/http"
)

type image struct {
	Contents []byte
	Mime     string
}

func downloadImage(url string) (image, error) {
	r, err := http.Get(url)
	if err != nil {
		return image{}, err
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return image{}, err
	}

	mime := http.DetectContentType(b)

	return image{b, mime}, nil
}
