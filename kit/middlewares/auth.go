package middlewares

import (
	"bytes"
	"github.com/ducthanh98/server-kit/kit/transhttp"
	"github.com/ducthanh98/server-kit/kit/utils/error_messages"
	"github.com/ducthanh98/server-kit/kit/utils/hash"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func AuthRequired(c *gin.Context) {
	token := c.GetHeader("x-access-token")
	payload := ""
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodDelete {
		payload = c.Request.URL.String()
	} else {

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Errorf("Auth middleware: Parse body error: %v", err)
			transhttp.ResponseError(c, http.StatusInternalServerError, error_messages.MessageInternalServerError)
			return
		}

		payload = string(body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}

	hashResult := hash.GenerateSha256(payload)
	if hashResult == token {
		c.Next()
	} else {
		transhttp.ResponseError(c, http.StatusUnauthorized, error_messages.Unauthorized)
	}
}
