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

		c.Data(http.StatusOK, img.Mime, img.Contents)
	}
}

type batchController struct {
}

func (r *batchController) batch(workQueue chan resizeRequest) gin.HandlerFunc {
	return func(c *gin.Context) {
		//apiKey := c.Param("apiKey")
		var batchParams batchParams
		c.BindJSON(&batchParams)

		im, err := downloadImage(batchParams.URL)
		if err != nil {
			handleError(c, err)
			return
		}

		mu := &sync.Mutex{}
		buf := new(bytes.Buffer)
		w := zip.NewWriter(buf)
		defer w.Close()

		outChan := make(chan resizeRequestResult)
		defer close(outChan)

		numOfOps := len(batchParams.Operations)
		log.Print(numOfOps)
		errors := make([]error, 0)

		for _, operation := range batchParams.Operations {
			bp, err := parseBatchOperationParams(operation.RawOperationParams, im)
			if err != nil {
				errors = append(errors, err)
			}
			rR := resizeRequest{
				im,
				bp.ImageParams,
				outChan,
				operation,
			}
			workQueue <- rR
		}

		resultCount := 0
		for resizedRequest := range outChan {
			resultCount++
			log.Print(resultCount)
			mu.Lock()
			f, err := w.Create(resizedRequest.batchOperation.Filename)
			if err != nil {
				log.Fatal(err)
			}

			_, err = f.Write(resizedRequest.image.Contents)
			if err != nil {
				log.Fatal(err)
			}
			mu.Unlock()
			if resultCount >= numOfOps {
				return
			}
		}

		if len(errors) > 0 {
			handleErrors(c, errors)
			return
		}
		c.Data(http.StatusOK, "application/zip", buf.Bytes())
	}
}
