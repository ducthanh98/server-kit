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
	Status    string      `json:"status"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, response{
		Status:    ResponseStatusSuccess,
		Data:      data,
		RequestID: GetRequestID(c),
	})
}

func ResponseError(c *gin.Context, status int, msg string, data interface{}) {
	c.JSON(status, response{
		Status:    ResponseStatusError,
		Message:   msg,
		Data:      data,
		RequestID: GetRequestID(c),
	})
}
