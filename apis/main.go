package main

import (
	"github.com/ducthanh98/server-kit/apis/webserver"
	"github.com/ducthanh98/server-kit/kit/app"
	"github.com/ducthanh98/server-kit/kit/logger"
	"net/http"
)

// @title API
// @version 1.0
// @description This is a sample Gin API with swaggo documentation.
// @host localhost:8080
// @BasePath /
func main() {
	microApp := app.InitApp("teka-api")
	webserver.NewServer(microApp)

	err := microApp.RunApp()
	if err != nil {
		if err.Error() == http.ErrServerClosed.Error() {
			logger.Log.Info(http.ErrServerClosed.Error())
		} else {
			logger.Log.Errorf("HTTP server closed with error: %v", err)
		}
	}
	onClose(microApp)

}

func onClose(microApp *app.App) {
	webserver.OnClose()
}
