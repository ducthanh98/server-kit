package middlewares

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func LogRequestMiddleware(c *gin.Context) {
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw

	c.Next()
	message := fmt.Sprintf("%v Go routines: %v  Uri: %v, Method: %v, Body: %v, Resp code: %v, Resp body: %v", time.Now().Format(time.RFC822),
		runtime.NumGoroutine(), c.Request.RequestURI, c.Request.Method, string(bodyBytes), c.Writer.Status(), blw.body.String())
	if c.Writer.Status() != http.StatusOK {
		log.Errorf(message)
	} else {
		log.Infof(message)
	}

}
