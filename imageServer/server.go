package imageServer

import (
	"net/http"

	"io/ioutil"
	"strings"

	"strconv"

	"fmt"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-errors/errors"
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

func handleError(c *gin.Context, err error, status ...int) {
	var statusCode int
	if len(status) == 0 {
		statusCode = http.StatusBadRequest
	} else {
		statusCode = status[0]
	}

	log.Println(err)

	c.String(statusCode, err.Error())
}

func downloadImage(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
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

func Run() {
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/:apiKey/:params/*url", func(c *gin.Context) {
			//apiKey := c.Param("apiKey")
			params := c.Param("params")
			url := c.Param("url")[1:]

			b, err := downloadImage(url)
			if err != nil {
				handleError(c, err)
				return
			}

			p, err := parseParams(params)
			if err != nil {
				handleError(c, err)
				return
			}

			img, err := bimg.Resize(b, p.toBimgOptions())
			if err != nil {
				handleError(c, err)
				return
			}

			c.Data(http.StatusOK, "image/png", img)
		})
	}

	r.Run(":9001")
}
