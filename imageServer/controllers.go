package imageServer

import (
	"archive/zip"
	"bytes"
	"net/http"

	"strings"

	"log"

	"github.com/gin-gonic/gin"
)

type resizeController struct {
}

func (r *resizeController) resize() gin.HandlerFunc {
	return func(c *gin.Context) {
		//apiKey := c.Param("apiKey")
		params := c.Param("params")
		url := c.Param("url")[1:]

		image, err := downloadImage(url)
		if err != nil {
			handleError(c, err)
			return
		}

		//Modded to reuse in batch operation parsing
		splitQuery := strings.Split(params, "=")[1]
		splitParams := strings.Split(splitQuery, ",")
		p, err := parseParams(splitParams, image)
		if err != nil {
			handleError(c, err)
			return
		}

		validationError := validateParams(p)
		if validationError != nil {
			handleError(c, validationError)
			return
		}

		ch := make(chan ResizeResult)
		defer close(ch)
		go resize(image.Contents, p, ch)
		result := <-ch

		if result.Error != nil {
			handleError(c, result.Error)
			return
		}

		c.Data(http.StatusOK, result.Image.GetMimeType(), result.Image.Contents)
	}
}

type batchController struct {
}

func (r *batchController) batch() gin.HandlerFunc {
	return func(c *gin.Context) {
		//apiKey := c.Param("apiKey")
		var batchParams BatchParams
		c.BindJSON(&batchParams)

		image, err := downloadImage(batchParams.URL)
		if err != nil {
			handleError(c, err)
			return
		}

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		defer w.Close()

		errors := make([]error, 0)
		resultChannels := make([]chan ResizeResult, 0)
		for _, operation := range batchParams.Operations {
			ch := make(chan ResizeResult)
			defer close(ch)
			resultChannels = append(resultChannels, ch)
			bp, err := parseBatchOperationParams(operation.RawOperationParams, image)
			if err != nil {
				errors = append(errors, err)
				return
			}

			go resize(image.Contents, bp.ImageParams, ch)
		}

		for i, ch := range resultChannels {
			res := <-ch

			if res.Error != nil {
				errors = append(errors, res.Error)
				continue
			}

			operation := batchParams.Operations[i]
			f, err := w.Create(operation.Filename) //TODO: operation and channel might not be from the same index!! bug!
			if err != nil {
				log.Fatal(err)
			}

			_, err = f.Write(res.Image.Contents)
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(errors) > 0 {
			handleErrors(c, errors)
			return
		}

		c.Data(http.StatusOK, "application/zip", buf.Bytes())
	}
}
