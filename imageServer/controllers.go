package imageServer

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type resizeController struct {
}

func (r *resizeController) resize() gin.HandlerFunc {
	return func(c *gin.Context) {
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

		img, err := resize(b, p)
		if err != nil {
			handleError(c, err)
			return
		}

		c.Data(http.StatusOK, "image/png", img)
	}
}
