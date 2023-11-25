package io

import (
	"io/ioutil"
	"os"
)

// ReadRawFile --
func ReadRawFile(file string) (string, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return "", err
	}

	b, err := ioutil.ReadFile(file)
	return string(b), err
}
