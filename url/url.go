package url

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	// RequestTypeGet is the presigned request type to download a file.
	RequestTypeGet = "GET"
	// RequestTypePUT is the presigned request type to upload a file.
	RequestTypePut = "PUT"
)

// FindURL configures a client using the key, secret, and region provided in the env file
// while also taking in a filename and request from the user and returns a presigned
// url to upload a file or download a file from a DigitalOcean Space or a "failed" string if error.
func FindURL(filename string, req string, duration string) string {
	key := os.Getenv("SPACES_KEY")
	secret := os.Getenv("SPACES_SECRET")
	bucket := os.Getenv("BUCKET")
	region := os.Getenv("REGION")

	dur, err := time.ParseDuration(duration)
	if err != nil {
		return "failed"
	}
	if dur < 0 {
		return "failed"
	}

	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:    aws.String(fmt.Sprintf("%s.digitaloceanspaces.com:443", region)),
		Region:      aws.String(region),
	}
	sess := session.New(config)

	switch req {
	case RequestTypeGet:
		return downloadURL(sess, bucket, filename, dur)
	case RequestTypePut:
		return uploadURL(sess, bucket, filename, dur)
	default:
		return "failed"
	}
}

func uploadURL(sess *session.Session, bucket string, filename string, duration time.Duration) string {
	client := s3.New(sess)
	req, _ := client.PutObjectRequest(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	url, err := req.Presign(duration)
	if err != nil {
		return "failed"
	}
	return url
}

func downloadURL(sess *session.Session, bucket string, filename string, duration time.Duration) string {
	client := s3.New(sess)
	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	})
	url, err := req.Presign(duration)
	if err != nil {
		return "failed"
	}
	return url
}
