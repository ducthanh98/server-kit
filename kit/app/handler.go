package app

import "github.com/gin-gonic/gin"

type Handler interface {
	ServeHTTP(ctx *gin.Context)
}
