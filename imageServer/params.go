package imageServer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
)

type imageParams struct {
	Width, Height int
	PixelDensity  float64
	Upscale       *bool //using a pointer here allows us to check for nil
	Mime          string
}

type batchParams struct {
	URL string `json:"url"`
	Operations []Operation `json:"operations"`
}

type Operation struct {
	Filename string `json:"filename"`
	RawParams string `json:"operation"`
}

func (p imageParams) getImageType() bimg.ImageType {
	if p.Mime == "image/jpeg" || p.Mime == "image/jpg" {
		return bimg.JPEG
	} else if p.Mime == "image/png" {
		return bimg.PNG
	} else if p.Mime == "image/webp" {
		return bimg.WEBP
	} else if p.Mime == "image/tiff" {
		return bimg.TIFF
	} else if p.Mime == "image/magick" {
		return bimg.MAGICK
	}

	return bimg.UNKNOWN
}

func (p imageParams) getWidth() int {
	return int(float64(p.Width) * p.PixelDensity)
}

func (p imageParams) getHeight() int {
	return int(float64(p.Height) * p.PixelDensity)
}

func (p imageParams) toBimgOptions() bimg.Options {
	return bimg.Options{
		Width:  p.getWidth(),
		Height: p.getHeight(),
		Type:   p.getImageType(),
	}
}

func parseParams(params string, img image) (imageParams, error) {
	splitQuery := strings.Split(params, "=")[1]
	splitParams := strings.Split(splitQuery, ",")
	p := &imageParams{}
	for _, s := range splitParams {
		options := strings.Split(s, ":")
		key := options[0]
		value := options[1]
		//TODO: validations
		if key == "w" {
			w, err := strconv.Atoi(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid width value: %s", value)) //TODO: percentages
			}

			p.Width = w
		} else if key == "h" {
			h, err := strconv.Atoi(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid height value: %s", value))
			}

			p.Height = h
		} else if key == "pd" {
			pd, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid pixel density value: %s", value))
			}

			p.PixelDensity = pd
		} else if key == "upscale" {
			upscale, err := strconv.ParseBool(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid upscale value: %s", value))
			}

			p.Upscale = &upscale
		}
	}

	//TODO: I hardly think the png compression param is needed
	//TODO: metadata retention

	//TODO: error if not width/height?

	if p.Mime == "" { //TODO: allow specifying type in params
		p.Mime = img.Mime
	}

	if p.PixelDensity == 0 {
		p.PixelDensity = 1
	}

	if p.Upscale == nil {
		p.Upscale = new(bool) //pointer of false
	}

	if *p.Upscale == false {
		//TODO: reuse size and/or meta?
		s, err := bimg.Size(img.Contents)
		if err != nil {
			return *p, err
		}

		if p.Width > s.Width {
			p.Width = s.Width
		}

		if p.Height > s.Height {
			p.Height = s.Height
		}
	}

	return *p, nil
}

func validateParams(p imageParams, img image) error {
	//TODO:
	return nil
}
