package transhttp

import (
	"github.com/ducthanh98/server-kit/kit/constants"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

type UserToken struct {
	UserID    int64
	Role      string
	ExpiredAt int64
	jwt.Claims
}

func GenerateUserAccessToken(tokenInstant *UserToken) (string, error) {
	expired := time.Now().Unix() + viper.GetInt64("auth.token_expired_time")
	tokenPassword := viper.GetString("auth.token_password")
	tokenInstant.ExpiredAt = expired
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenInstant)
	tokenString, err := token.SignedString([]byte(tokenPassword))

	return tokenString, err
}

func IsCorrectPassword(hashedPwd string, plainPwd string) bool {
	plainPwdBytes := []byte(plainPwd)
	byteHash := []byte(hashedPwd)

	err := bcrypt.CompareHashAndPassword(byteHash, plainPwdBytes)
	if err != nil {
		return false
	}

	return true
}

func GetUserIDFromRequest(r *http.Request) int64 {
	userID := cast.ToInt64(r.Header.Get("X-User-ID"))
	logger.Log.Infof("%v", r.Header)
	return userID
}

func GetUserRoleFromRequest(r *http.Request) string {
	role := cast.ToString(r.Header.Get("X-User-Role"))
	logger.Log.Infof("%v", r.Header)
	return role
}

func GetUserAccessToken(c *gin.Context) string {
	// Grab the token from the header
	accessToken := c.GetHeader(constants.TokenKeyUserType)
	if accessToken == "" {
		accessToken = c.Query(constants.TokenKeyUserType)
	}

	return accessToken
}

func ParseUserAccessToken(c *gin.Context) (userToken *UserToken, err error) {
	userAccessToken := GetUserAccessToken(c)
	return UserAccessToken(userAccessToken)
}

func UserAccessToken(userAccessToken string) (userToken *UserToken, err error) {
	userAccessToken = strings.Replace(userAccessToken, "Bearer ", "", 1)
	if userAccessToken == "" { //Token is missing, returns with error code 401 Unauthorized
		return nil, constants.HttpErrorTokenMissing
	}

	userToken = &UserToken{}

	tokenPassword := viper.GetString("auth.token_password")
	token, err := jwt.ParseWithClaims(userAccessToken, userToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenPassword), nil
	})

	//Malformed token, returns with http code 401 as usual
	if err != nil {
		return nil, constants.HttpErrorTokenInvalid
	}

	if !token.Valid { //Token is invalid, maybe not signed on this server
		return nil, constants.HttpErrorTokenInvalid
	}

	if userToken.ExpiredAt < time.Now().Unix() {
		return nil, constants.HttpErrorTokenExpired
	}

	return userToken, nil
}
