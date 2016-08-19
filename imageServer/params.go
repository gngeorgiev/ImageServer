package imageServer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/h2non/bimg.v1"
)

type ImageParams struct {
	Width, Height ImageDimension
	Image         Image
	PixelDensity  float64
	Upscale       *bool //using a pointer here allows us to check for nil
}

type ImageDimension struct {
	Value        int
	IsPercentage bool
}

type BatchParams struct {
	URL        string           `json:"url"`
	Operations []BatchOperation `json:"operations"`
}

type BatchOperation struct {
	Filename           string `json:"filename"`
	RawOperationParams string `json:"operation"`
}

type BatchOperationParams struct {
	Operation   string
	ImageParams ImageParams
}

func (p ImageParams) getImageType() bimg.ImageType {
	m := p.Image.GetMimeType()

	if m == "image/jpeg" || m == "image/jpg" {
		return bimg.JPEG
	} else if m == "image/png" {
		return bimg.PNG
	} else if m == "image/webp" {
		return bimg.WEBP
	} else if m == "image/tiff" {
		return bimg.TIFF
	} else if m == "image/magick" {
		return bimg.MAGICK
	}

	return bimg.UNKNOWN
}

func (p ImageParams) getWidth() int {
	var width float64
	if p.Width.IsPercentage {
		percents := float64(p.Width.Value) / 100
		width = float64(p.Image.Metadata.Size.Width) * percents
	} else {
		width = float64(p.Width.Value)
	}

	return int(width * p.PixelDensity)
}

func (p ImageParams) getHeight() int {
	var height float64
	if p.Height.IsPercentage {
		percents := float64(p.Height.Value) / 100
		height = float64(p.Image.Metadata.Size.Height) * percents
	} else {
		height = float64(p.Height.Value)
	}

	return int(height * p.PixelDensity)
}

func (p ImageParams) toBimgOptions() bimg.Options {
	return bimg.Options{
		Width:  p.getWidth(),
		Height: p.getHeight(),
		Type:   p.getImageType(),
	}
}

func parseImageDimension(value, paramName string) (ImageDimension, error) {
	var isPercentage bool
	var valueToParse string

	if strings.Contains(value, "pct") {
		isPercentage = true
		valueToParse = strings.Replace(value, "pct", "", 1)
	} else {
		valueToParse = value
	}

	parsedValue, err := strconv.Atoi(valueToParse)
	if err != nil {
		return ImageDimension{}, errors.New(fmt.Sprintf("Invalid %s value: %s", paramName, value))
	}

	return ImageDimension{parsedValue, isPercentage}, nil
}

//Modded to reuse in batch operation parsing
func parseParams(params []string, img Image) (ImageParams, error) {
	p := &ImageParams{Image: img}
	for _, s := range params {
		options := strings.Split(s, ":")
		key := options[0]
		value := options[1]
		//TODO: validations
		if key == "w" {
			w, err := parseImageDimension(value, "width")
			if err != nil {
				return *p, err
			}

			p.Width = w
		} else if key == "h" {
			h, err := parseImageDimension(value, "height")
			if err != nil {
				return *p, err
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
	//TODO: metadata retention?
	//TODO: allow specifying type to convert to in params

	//TODO: error if not width/height?

	if p.Width.Value == 0 && p.Height.Value == 0 {
		return *p, errors.New("Invalid width and height parameters, specify at least one")
	}

	if p.PixelDensity == 0 {
		p.PixelDensity = 1
	}

	if p.Upscale == nil {
		p.Upscale = new(bool) //pointer of false
	}

	if *p.Upscale == false {
		if p.Width.Value > p.Image.Metadata.Size.Width {
			p.Width = ImageDimension{p.Image.Metadata.Size.Width, false}
		}

		if p.Height.Value > p.Image.Metadata.Size.Height {
			p.Height = ImageDimension{p.Image.Metadata.Size.Height, false}
		}
	}

	return *p, nil
}

func parseBatchOperationParams(rawOperationParams string, img Image) (BatchOperationParams, error) {
	bp := &BatchOperationParams{}

	splitQuery := strings.Split(rawOperationParams, "=")
	bp.Operation = splitQuery[0]
	splitParams := strings.Split(splitQuery[1], ",")
	p, err := parseParams(splitParams, img)
	if err != nil {
		return *bp, err
	}
	bp.ImageParams = p
	return *bp, nil
}

func validateParams(p ImageParams, img Image) error {
	//TODO:
	return nil
}
