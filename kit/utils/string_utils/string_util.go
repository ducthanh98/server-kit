package string_utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/ducthanh98/server-kit/kit/logger"
	"strings"
)

// IsStringSliceContains -- check slice contain string
func IsStringSliceContains(stringSlice []string, searchString string) bool {
	for _, value := range stringSlice {
		if value == searchString {
			return true
		}
	}
	return false
}

// StringTrimSpace -- trim space of string
func StringTrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// IsStringEmpty -- check if string is empty
func IsStringEmpty(s string) bool {
	return s == ""
}

func CensorString(str string) string {
	if len(str) <= 6 {
		return "***"
	}

	return str[:2] + "***" + str[len(str)-2:]
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

// ToJSONString -- convert data to json string
func ToJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		logger.Log.Warnf("Error When convert interface : %v to json string, detail: %v", v, err)
		return ""
	}
	return string(b)
}
