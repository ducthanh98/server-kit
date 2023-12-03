package app

import (
	"fmt"
	"github.com/ducthanh98/server-kit/apis/docs"
	_ "github.com/ducthanh98/server-kit/apis/docs"
	"github.com/ducthanh98/server-kit/kit/config"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/ducthanh98/server-kit/kit/middlewares"
	"github.com/ducthanh98/server-kit/kit/utils/string_utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"strings"
)

var (

	// LoggerTypeFile -- logger type FILE, log will be write to file
	LoggerTypeFile = "file"
	// LoggerTypeDefault -- logger type DEFAULT, log will be write to stdout
	LoggerTypeDefault = "default"
	// SupportedLoggerTypes -- list logger type supported
	SupportedLoggerTypes = []string{LoggerTypeFile, LoggerTypeDefault}
)

type App struct {
	Engine  *gin.Engine
	AppName string

	LoggerType string
}

func InitApp(appName string) *App {
	config.LoadConfig()
	gin.SetMode(viper.GetString("server.mode"))
	r := gin.Default()

	return &App{
		Engine:     r,
		AppName:    appName,
		LoggerType: viper.GetString("server.logger_type"),
	}
}

// InitLogger -- initializes logger
func (app *App) InitLogger() {
	loggerType := string_utils.StringTrimSpace(strings.ToLower(app.LoggerType))
	// if loggerType is empty, set default to `file` type`
	if string_utils.IsStringEmpty(loggerType) {
		loggerType = LoggerTypeFile
	}
	// check supported loggerType
	if !string_utils.IsStringSliceContains(SupportedLoggerTypes, loggerType) {
		logger.Log.Panicf("Not supported logger type: %v", loggerType)
	}

	switch loggerType {
	case LoggerTypeDefault:
		logger.InitLoggerDefault()
	case LoggerTypeFile:
		logger.InitLoggerFile()

	}

	logger.Log.Info("Logger loaded")
}

func (s *App) RunGlobalMiddleware() {
	s.Engine.Use(cors.Default())
	docs.SwaggerInfo.BasePath = "/api/v1"
	s.Engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.Engine.Use(middlewares.IdentifyRequest)
	s.Engine.Use(middlewares.LogRequestMiddleware)
}

func (s *App) AddRoutes(routes Routes) {
	defaultRouter := s.Engine.Group(DefaultBasePath)
	for _, route := range routes {
		if route.AuthInfo.Enable {
			defaultRouter.Handle(route.Method, route.Pattern, middlewares.AuthenticationVerify(route.AuthInfo.UserRoles), route.Handler.ServeHTTP)
		} else {
			defaultRouter.Handle(route.Method, route.Pattern, route.Handler.ServeHTTP)
		}
	}

}

func (s *App) RunApp() error {
	s.InitLogger()
	s.RunGlobalMiddleware()

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	logger.Log.Infof("App is running at port: %v", port)
	return s.Engine.Run(fmt.Sprintf(":%v", port))
}
