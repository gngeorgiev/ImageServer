package imageServer

import (
	"net/http"

	"archive/zip"
	"bytes"
	"log"
	"strings"
	"sync"

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

		c.Data(http.StatusOK, img.GetMimeType(), img.Contents)
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

		wg := sync.WaitGroup{}

		mu := &sync.Mutex{}

		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		defer w.Close()

		errors := make([]error, 0)
		for _, operation := range batchParams.Operations {
			wg.Add(1)
			go func(operation BatchOperation) {
				defer wg.Done()

				bp, err := parseBatchOperationParams(operation.RawOperationParams, image)
				if err != nil {
					errors = append(errors, err)
					return
				}

				img, err := resize(image.Contents, bp.ImageParams)
				if err != nil {
					errors = append(errors, err)
					return
				}

				mu.Lock()
				defer mu.Unlock()
				f, err := w.Create(operation.Filename)
				if err != nil {
					log.Fatal(err)
				}

				_, err = f.Write(img.Contents)
				if err != nil {
					log.Fatal(err)
				}
			}(operation)
		}

		wg.Wait()

		if len(errors) > 0 {
			handleErrors(c, errors)
			return
		}

		c.Data(http.StatusOK, "application/zip", buf.Bytes())
	}
}
