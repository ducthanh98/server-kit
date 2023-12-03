package app

// Route -- Defines a single route, e.g. a human readable name, HTTP method,
// pattern the function that will execute when the route is called.
type Route struct {
	Name     string
	Method   string
	BasePath string
	Pattern  string
	Handler  Handler
	AuthInfo AuthInfo
	Timeout  int64
}

// AuthInfo -- authentication and authorization for route
type AuthInfo struct {
	Enable    bool
	UserRoles map[string]bool
}

// Routes -- Defines the type Routes which is just an array (slice) of Route structs.
type Routes []Route
