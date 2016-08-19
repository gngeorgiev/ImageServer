package imageServer

import (
	"net/http"

	"log"

	"github.com/gin-gonic/gin"
)

func handleErrors(c *gin.Context, err []error, status ...int) {
	var statusCode int
	if len(status) == 0 {
		statusCode = http.StatusBadRequest
	} else {
		statusCode = status[0]
	}

	log.Println(err)

	c.JSON(statusCode, map[string]interface{}{
		"errors": err,
	})
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

func Run() {
	r := gin.Default()

	rController := &resizeController{}
	bController := &batchController{}

	v1 := r.Group("/v1")
	{
		apiKey := v1.Group("/:apiKey")
		{
			apiKey.GET("/:params/*url", rController.resize())
			apiKey.POST("/batch", bController.batch())

		}
	}

	r.Run(":9001")
}
