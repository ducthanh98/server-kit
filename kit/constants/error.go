package constants

import "errors"

var (
	HttpErrorTokenExpired = errors.New("Token expired ")
	HttpErrorTokenMissing = errors.New("Missing authentication token ")
	HttpErrorForbidden    = errors.New("403 Forbidden – you don’t have permission to access on server ")
	HttpErrorTokenInvalid = errors.New("Token invalid ")
	HttpErrorNotFound     = errors.New("Not found! ")
	HttpErrorParseBody    = errors.New("error while parse request body")
)
