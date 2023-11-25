package app

import (
	"fmt"
	"github.com/ducthanh98/server-kit/kit/config"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/ducthanh98/server-kit/kit/middlewares"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type App struct {
	Engine  *gin.Engine
	AppName string
}

type Route struct {
	Method  string
	Path    string
	Handler Handler
}

type Routes []Route

func InitApp(appName string) *App {
	config.LoadConfig()
	gin.SetMode(viper.GetString("server.mode"))
	r := gin.Default()
	logger.RunCustomLogger(appName)

	return &App{
		Engine:  r,
		AppName: appName,
	}
}

func (s *App) RunAPI(routes Routes) {
	s.RunGlobalMiddleware()
	defaultRouter := s.Engine.Group(DefaultBasePath)

	for _, route := range routes {
		defaultRouter.Handle(route.Method, route.Path, route.Handler.ServeHTTP)
	}
}

func (s *App) RunGlobalMiddleware() {
	s.Engine.Use(middlewares.IdentifyRequest)
	s.Engine.Use(middlewares.LogRequestMiddleware)
}

func (s *App) RunApp() error {
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	log.Infof("App is running at port: %v", port)
	return s.Engine.Run(fmt.Sprintf(":%v", port))
}
