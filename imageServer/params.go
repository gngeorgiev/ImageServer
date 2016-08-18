package imageServer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
)

type ImageParams struct {
	Width, Height int
}

func (p ImageParams) toBimgOptions() bimg.Options {
	return bimg.Options{
		Width:  p.Width,
		Height: p.Height,
	}
}

func parseParams(params string) (ImageParams, error) {
	splitQuery := strings.Split(params, "=")[1]
	splitParams := strings.Split(splitQuery, ",")
	p := &ImageParams{}
	for _, s := range splitParams {
		options := strings.Split(s, ":")
		key := options[0]
		value := options[1]
		if key == "w" {
			w, err := strconv.Atoi(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid width value: %s", value))
			}

			p.Width = w
		} else if key == "h" {
			h, err := strconv.Atoi(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid height value: %s", value))
			}

			p.Height = h
		}
	}

	return *p, nil
}
