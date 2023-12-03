package webserver

import (
	"github.com/ducthanh98/server-kit/apis/webserver/handlers"
	"github.com/ducthanh98/server-kit/kit/app"
	"github.com/ducthanh98/server-kit/kit/constants"
	"gorm.io/gorm"
	"net/http"
)

var (
	sqlCon *gorm.DB
)

// NewServer return a new server
func NewServer(microApp *app.App) {
	sqlCon = microApp.CreateSqlConnection(nil, constants.PostgresqlDriver)
	s := &server{}

	microApp.AddRoutes(s.InitRoutes(""))
}

// OnClose -- on close
func OnClose() {
	if db, _ := sqlCon.DB(); db != nil {
		_ = db.Close()
	}
}

// Server implements test-api service
type server struct {
}

// InitRoutes -- Initialize our routes
func (s *server) InitRoutes(basePath string) app.Routes {
	// Initialize our routes
	return app.Routes{
		app.Route{
			Name:     "Get product bases",
			Method:   http.MethodGet,
			BasePath: basePath,
			Pattern:  "bases",
			Handler:  &handlers.SampleHandler{},
			AuthInfo: app.AuthInfo{
				Enable: true,
				UserRoles: map[string]bool{
					constants.UserRoleAdmin: true,
				},
			},
		},
	}
}
