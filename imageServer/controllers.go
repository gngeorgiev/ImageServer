package imageServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"strings"
	"bytes"
	"archive/zip"
	"log"
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

		validationError := validateParams(p, image)
		if validationError != nil {
			handleError(c, validationError)
			return
		}

		img, err := resize(image.Contents, p)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Data(http.StatusOK, img.Mime, img.Contents)
	}
}

type batchController struct {
}

func (r *batchController) batch() gin.HandlerFunc {
	return func(c *gin.Context) {
		//apiKey := c.Param("apiKey")
		var batchParams batchParams
		c.BindJSON(&batchParams)

		image, err := downloadImage(batchParams.URL)
		if err != nil {
			handleError(c, err)
			return
		}

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		for _, operation := range batchParams.Operations {
			bp, err := parseBatchOperationParams(operation.RawOperationParams, image)

			img, err := resize(image.Contents, bp.ImageParams)
			if err != nil {
				handleError(c, err)
				return
			}

			f, err := w.Create(operation.Filename)
			if err != nil {
				log.Fatal(err)
			}
			_, err = f.Write(img.Contents)
			if err != nil {
				log.Fatal(err)
			}

		}
		err = w.Close()
		if err != nil {
			handleError(c, err)
		}
		//TODO create and return zip file
		c.Data(http.StatusOK, "application/zip", buf.Bytes())
	}
}