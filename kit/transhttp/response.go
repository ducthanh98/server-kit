package transhttp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ResponseStatusSuccess = "success"
	ResponseStatusError   = "error"
)

type response struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{
		Message:   ResponseStatusSuccess,
		Data:      data,
		RequestID: GetRequestID(c),
	})
}

func ResponseError(c *gin.Context, status int, data interface{}) {
	c.JSON(status, response{
		Message:   ResponseStatusError,
		Data:      data,
		RequestID: GetRequestID(c),
	})
}
