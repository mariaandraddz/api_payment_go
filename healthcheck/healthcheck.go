package healthcheck

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(c *gin.context) {
	path := c.FullPath()
	if path == "" {
		path = "/health"
	}

	observability.RecordHealthCheck(c.Request.Method, path)
	c.JSON(http.StatusOK, gin.H{
		"status": observability.StatusHealthCheck,
	})
}
