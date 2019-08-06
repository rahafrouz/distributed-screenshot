package common

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"sync"
)

func TokenGenerator(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

type ScreenshotHanlder interface {
	TakeScreenshot(url string, destination string, savetofile bool) ([]byte, bool)
}

type CloudStorageHandler interface {
	init()
	UploadToCloud(storagePath string, localPath string, data []byte)
}

type S3Storage struct {
	initialized sync.Once
	uploader    *s3manager.Uploader
	bucket      string
}

func (s3 *S3Storage) init() {
	//It would happen once. Init the S3 credentials
	//select Region to use.
	conf := aws.Config{
		Region:      aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
	}
	conf.WithCredentialsChainVerboseErrors(true)
	sess, err := session.NewSession(&conf)
	failOnError(err, "Problem in setting aws session")
	s3.uploader = s3manager.NewUploader(sess)

	s3.bucket = os.Getenv("BUCKET_NAME")

}

func (s3 *S3Storage) UploadFileToCloud(localFilePath string, fileNameInCloud string) (string, error) {
	panic("Not implemented yet")
}

func (s3 *S3Storage) UploadDataToCloud(filename string, data []byte) (string, error) {
	//Do the initialization only once (Just the first time)
	(*s3).initialized.Do(func() { // <-- atomic, does not allow repeating
		(*s3).init()
	})

	//Upload to s3
	fmt.Println("Uploading file to S3...")
	result, err := s3.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
	})

	failOnError(err, "Error Uploading to S3")
	//Everything was fine so far

	fmt.Println("Result is: " + result.Location)
	return result.Location, nil

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		os.Exit(88)
	}
}
