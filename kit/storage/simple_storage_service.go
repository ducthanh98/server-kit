package storage

import (
	"bytes"
	"io"
)

const (
	AwsType   = "aws"
	LocalType = "local"
)

// S3 --
type S3 interface {
	UploadFile(file io.Reader, filename string, bucket string, contentType string) error
	UploadChunkToS3(fileName, path, bucket, contentType string, header []string, data [][]string) error
	ReadFile(filename string, bucket string) (*bytes.Buffer, error)
	ReadFileForUpload(filename string, bucket string) (*bytes.Buffer, error, *string)
	Exist(url string, bucket string) (bool, error)
	Delete(listKey []string, bucket string) error
}

type S3Configuration struct {
	Type string
}
