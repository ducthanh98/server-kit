package transhttp

import "github.com/gin-gonic/gin"

func GetRequestID(c *gin.Context) string {
	return c.GetHeader("X-Request-Id")
}
