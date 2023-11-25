package storage

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

type AmazonS3 struct {
	Session *session.Session
	Domain  string
	Bucket  string
	Region  string
}

type AmazonS3Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Domain          string
	Bucket          string
	Region          string
}

func NewAmazonS3(opts *AmazonS3Config) *AmazonS3 {
	if opts == nil {
		opts = &AmazonS3Config{
			AccessKeyID:     viper.GetString("aws.access_key_id"),
			SecretAccessKey: viper.GetString("aws.secret_access_key"),
			Domain:          viper.GetString("aws.endpoint"),
			Region:          viper.GetString("aws.region"),
		}
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(opts.Region),
		Credentials: credentials.NewStaticCredentials(opts.AccessKeyID, opts.SecretAccessKey, ""),
	})

	if err != nil {
		log.Errorf("Get session AWS errors, %v", err)
		return nil
	}

	return &AmazonS3{
		Session: sess,
		Domain:  opts.Domain,
		Bucket:  opts.Bucket,
		Region:  opts.Region,
	}
}

func (s *AmazonS3) UploadFile(file io.Reader, filename string, bucket string, contentType string) error {
	bf, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if contentType == "" {
		contentType = http.DetectContentType(bf)
	}

	_, err = s3.New(s.Session).PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
		//ACL:                  aws.String("public-read"),
		Body:                 bytes.NewReader(bf),
		ContentLength:        aws.Int64(int64(len(bf))),
		ContentType:          aws.String(contentType),
		ContentDisposition:   aws.String("attachment"),
		ServerSideEncryption: aws.String("AES256"),
		ContentEncoding:      aws.String("utf-8"),
	})

	return err
}

func (s *AmazonS3) ReadFile(url string, bucket string) (*bytes.Buffer, error) {
	key := s.GetKeyFromUrl(url)
	results, err := s3.New(s.Session).GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return bytes.NewBuffer(nil), err
	}
	defer results.Body.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, results.Body); err != nil {
		return bytes.NewBuffer(nil), err
	}

	return buf, nil
}

func (s *AmazonS3) UploadChunkToS3(fileName, path, bucket, contentType string, header []string, data [][]string) error {
	return nil
}

func (s *AmazonS3) Exist(url string, bucket string) (bool, error) {
	key := s.GetKeyFromUrl(url)
	_, err := s3.New(s.Session).HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *AmazonS3) Delete(listKey []string, bucket string) error {
	objectsToDelete := make([]*s3.ObjectIdentifier, 0, 1000)

	for _, url := range listKey {
		key := s.GetKeyFromUrl(url)
		obj := s3.ObjectIdentifier{
			Key: aws.String(key),
		}
		objectsToDelete = append(objectsToDelete, &obj)
	}
	deleteArray := s3.Delete{Objects: objectsToDelete}
	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &deleteArray,
	}

	_, err := s3.New(s.Session).DeleteObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Errorf("Delete AWS errors, %v", err)
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Errorf("Delete AWS errors, %v", err)
		}
		return err
	}

	return nil
}

func (s *AmazonS3) GetKeyFromUrl(url string) string {
	key := url
	reg := regexp.MustCompile("^(" + s.Domain + ")")
	if s.Domain != "" && reg.MatchString(url) {
		key = reg.ReplaceAllString(key, "")
	} else {
		reg = regexp.MustCompile(`(?mi)^http(s)?:\/\/([a-zA-Z0-9-_.]+)\.amazonaws\.com\/([a-zA-Z0-9-_\.]+)`)
		key = reg.ReplaceAllString(key, "")
	}

	return key
}

func (s *AmazonS3) ReadFileForUpload(filename string, bucket string) (*bytes.Buffer, error, *string) {
	return nil, nil, nil
}
