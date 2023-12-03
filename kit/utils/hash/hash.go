package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"github.com/ducthanh98/server-kit/kit/logger"
	"github.com/spf13/viper"
)

func GenerateSha256(data string) string {
	secret := viper.GetString("media_server.media_secret")
	logger.Log.Infof("Secret: %v", secret)
	h := hmac.New(sha256.New, []byte(secret))

	// Write Data to it
	h.Write([]byte(data))

	// Get result and encode as hexadecimal string_utils
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
