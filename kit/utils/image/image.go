package image

import (
	"encoding/base64"
	"github.com/ducthanh98/server-kit/kit/logger"
	"io/ioutil"
	"net/http"
)

func ConvertImageToBase4(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Log.Errorf("Error when read %v : %v", path, err)
		return "", err
	}

	var base64Encoding string

	// Determine the content type of the image file
	mimeType := http.DetectContentType(bytes)

	// Prepend the appropriate URI scheme header depending
	// on the MIME type
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64,"
	case "image/png":
		base64Encoding += "data:image/png;base64,"
	}

	// Append the base64 encoded output
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)
	return base64Encoding, nil
}
