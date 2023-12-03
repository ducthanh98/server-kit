package handlers

import "github.com/gin-gonic/gin"

type SampleHandler struct {
}

// @BasePath /api/v1

// @BasePath /api/v1

// PingExample godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /example/helloworld [get]
func (s *SampleHandler) ServeHTTP(ctx *gin.Context) {

}
