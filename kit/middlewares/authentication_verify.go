package middlewares

import (
	"github.com/ducthanh98/server-kit/kit/constants"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/ducthanh98/server-kit/kit/transhttp"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/http"
)

func AuthenticationVerify(userRoles map[string]bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userToken, err := transhttp.ParseUserAccessToken(c)
		if err != nil {
			logger.Log.Error("cant parse user access token: ", err)
			transhttp.ResponseError(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		if userToken.UserID < 1 {
			logger.Log.Warnf("user token %v", constants.HttpErrorTokenInvalid.Error())
			transhttp.ResponseError(c, http.StatusUnauthorized, "Unauthorized", nil)
			return
		}

		// If user role change, need fix this
		if userRoles != nil && len(userRoles) > 0 && !userRoles[userToken.Role] {
			transhttp.ResponseError(c, http.StatusForbidden, "Forbidden", nil)
			return
		}

		// Assign data to request header
		c.Header("X-User-Id", cast.ToString(userToken.UserID))
		c.Header("X-User-Role", cast.ToString(userToken.Role))
		//logger.Log.Infow(
		//	fmt.Sprintf("Authencation info"),
		//	"X-User-Id", cast.ToString(userToken.UserID),
		//	"X-User-Role", cast.ToString(userToken.Role),
		//)
		c.Next()
	}
}
