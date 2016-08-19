package imageServer

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"math"

	"gopkg.in/h2non/bimg.v1"
)

const (
	FillContain = FillType("contain")
	FillCover   = FillType("cover")
	FillNone    = FillType("")
)

type FillType string

type ImageParams struct {
	Width, Height ImageDimension
	Image         Image
	PixelDensity  float64
	Upscale       *bool //using a pointer here allows us to check for nil
	Fill          FillType
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

func (p ImageParams) getFillResizedDimensions() (w, h float64) {
	if p.Fill == FillContain {
		if p.Width.Value != 0 {
			ratio := float64(p.Width.Value) / float64(p.Image.Metadata.Size.Width)
			scaledWidth := float64(p.Image.Metadata.Size.Width) * ratio
			scaledHeight := float64(p.Image.Metadata.Size.Height) * ratio

			if p.Height.Value == 0 || scaledHeight <= float64(p.Image.Metadata.Size.Height) {
				w = scaledWidth
				h = scaledHeight
				return
			}
		}

		ratio := float64(p.Height.Value) / float64(p.Image.Metadata.Size.Height)
		scaledWidth := float64(p.Image.Metadata.Size.Width) * ratio
		scaledHeight := float64(p.Image.Metadata.Size.Height) * ratio

		w = scaledWidth
		h = scaledHeight
		return
	} else if p.Fill == FillCover {
		if p.Width.Value == 0 || p.Height.Value == 0 {
			w = float64(p.Width.Value)
			h = float64(p.Height.Value)
			return
		}

		widthCoef := float64(p.Width.Value) / float64(p.Image.Metadata.Size.Width)
		heightCoef := float64(p.Height.Value) / float64(p.Image.Metadata.Size.Height)
		maxCoef := math.Max(widthCoef, heightCoef)

		resizedWidth := math.Ceil(maxCoef * float64(p.Image.Metadata.Size.Width))
		resizedHeight := math.Ceil(maxCoef * float64(p.Image.Metadata.Size.Height))

		w = resizedWidth
		h = resizedHeight
		return
	}

	//no fill
	w = float64(p.Image.Metadata.Size.Width)
	h = float64(p.Image.Metadata.Size.Height)
	return
}

func (p ImageParams) getResizedDimensions() (w, h int) {
	width, height := p.getFillResizedDimensions()

	if *p.Upscale == false && p.Fill == FillNone { //apply upscale restriction only if we have no fill option
		if p.Width.Value > p.Image.Metadata.Size.Width {
			width = float64(p.Image.Metadata.Size.Width)
		}

		if p.Height.Value > p.Image.Metadata.Size.Height {
			height = float64(p.Image.Metadata.Size.Height)
		}
	}

	if p.Width.Value != 0 {
		if p.Width.IsPercentage {
			percents := float64(p.Width.Value) / 100
			width = float64(width) * percents
		} else {
			width = float64(p.Width.Value)
		}

		if p.Height.Value == 0 {
			coefficient := width / float64(p.Image.Metadata.Size.Width)
			height = float64(p.Image.Metadata.Size.Height) * coefficient
		}
	}
	width *= p.PixelDensity

	if p.Height.Value != 0 {
		if p.Height.IsPercentage {
			percents := float64(p.Height.Value) / 100
			height = float64(height) * percents
		} else {
			height = float64(p.Height.Value)
		}

		if p.Width.Value == 0 {
			coefficient := height / float64(p.Image.Metadata.Size.Height)
			width = float64(p.Image.Metadata.Size.Width) * coefficient
		}
	}
	height *= p.PixelDensity

	w = int(width)
	h = int(height)
	return
}

func (p ImageParams) toBimgOptions() bimg.Options {
	width, height := p.getResizedDimensions()

	return bimg.Options{
		Width:  width,
		Height: height,
		Type:   p.getImageType(),
		Crop:   p.Fill == FillCover,
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
		switch key {
		case "w":
			w, err := parseImageDimension(value, "width")
			if err != nil {
				return *p, err
			}

			p.Width = w
		case "h":
			h, err := parseImageDimension(value, "height")
			if err != nil {
				return *p, err
			}

			p.Height = h
		case "pd":
			pd, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid pixel density value: %s", value))
			}

			p.PixelDensity = pd
		case "upscale":
			upscale, err := strconv.ParseBool(value)
			if err != nil {
				return *p, errors.New(fmt.Sprintf("Invalid upscale value: %s", value))
			}

			p.Upscale = &upscale
		case "fill":
			t := FillType(value)
			if t == FillContain {
				p.Fill = FillContain
			} else if t == FillCover {
				p.Fill = FillCover
			} else {
				return *p, errors.New(fmt.Sprintf("Invalid fill value: %s", value))
			}
		}
	}

	//TODO: I hardly think the png compression param is needed
	//TODO: metadata retention?
	//TODO: allow specifying type to convert to in params

	//TODO: error if not width/height? it seems not, investigate how the java server behaves
	//if p.Width.Value == 0 && p.Height.Value == 0 {
	//	return *p, errors.New("Invalid width and height parameters, specify at least one")
	//}

	if p.PixelDensity == 0 {
		p.PixelDensity = 1
	}

	if p.Upscale == nil {
		p.Upscale = new(bool) //pointer of false
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

func validateParams(p ImageParams) error {
	if p.Width.IsPercentage && p.Fill != FillNone {
		return errors.New(fmt.Sprintf("Cannot use percentage width: %d with fill: %s", p.Width.Value, p.Fill))
	}

	if p.Height.IsPercentage && p.Fill != FillNone {
		return errors.New(fmt.Sprintf("Cannot use percentage height: %d with fill: %s", p.Height.Value, p.Fill))
	}

	return nil
}
