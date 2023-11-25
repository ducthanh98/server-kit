package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func IdentifyRequest(c *gin.Context) {
	token := uuid.New().String()
	c.Request.Header.Add("X-Request-Id", token)

	c.Next()
}
