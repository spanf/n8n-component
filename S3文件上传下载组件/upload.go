package upload

import (
	"io"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
)

func UploadFile(client *s3.S3, bucket, key string, file io.Reader) error {
	if err := validateBucketName(bucket); err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	}

	_, err := client.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return aerr
		}
		return err
	}
	return nil
}

func validateBucketName(bucket string) error {
	if bucket == "" {
		return awserr.New("InvalidBucketName", "Bucket name cannot be empty", nil)
	}

	if len(bucket) < 3 || len(bucket) > 63 {
		return awserr.New("InvalidBucketName", "Bucket name must be between 3 and 63 characters", nil)
	}

	validBucketRegex := regexp.MustCompile(`^[a-z0-9][a-z0-9.-]+[a-z0-9]$`)
	if !validBucketRegex.MatchString(bucket) {
		return awserr.New("InvalidBucketName", "Bucket name contains invalid characters or format", nil)
	}

	return nil
}
