package common

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
	"sync"
)

type S3Storage struct {
	initialized sync.Once
	uploader    *s3manager.Uploader
	bucket      string
}

//Initialize the S3 structures. It should be called once, and we use a singleton pattern here.
//Other details of S3 are taken from environment variables.
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

//To upload a file from disk to the cloud. To use gowitness, we should look for files in a directory, find the screenshot file and upload it.
//Removed the implementation, as gowitness had problem with concurrency. chromedp is the preffered option.
func (s3 *S3Storage) UploadFileToCloud(localFilePath string, fileNameInCloud string) (string, error) {
	panic("Not implemented yet")
}

//Upload an array of bytes to the cloud. No file-system involved. Everything in memory.
func (s3 *S3Storage) UploadDataToCloud(filename string, data []byte) (string, error) {
	//Do the initialization only once (Just the first time)
	(*s3).initialized.Do(func() { // <-- atomic, does not allow repeating
		(*s3).init()
	})

	//Upload to s3
	log.Println("Uploading file to S3...")
	result, err := s3.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3.bucket),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
		ACL:    aws.String("public-read"),
	})

	failOnError(err, "Error Uploading to S3")
	//Everything was fine so far

	log.Println("Result is: " + result.Location)
	return result.Location, nil

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		os.Exit(88)
	}
}
