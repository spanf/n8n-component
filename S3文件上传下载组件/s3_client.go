package s3client

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewS3Client(accessKey, secretKey, region string) (*s3.S3, error) {
	if err := validateCredentials(accessKey, secretKey); err != nil {
		return nil, err
	}

	creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
	config := &aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	}

	sess, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}

	return s3.New(sess), nil
}

func validateCredentials(accessKey, secretKey string) error {
	if accessKey == "" {
		return errors.New("accessKey cannot be empty")
	}
	if secretKey == "" {
		return errors.New("secretKey cannot be empty")
	}
	if len(accessKey) < 16 {
		return errors.New("accessKey format invalid: minimum length 16 characters")
	}
	if len(secretKey) < 16 {
		return errors.New("secretKey format invalid: minimum length 16 characters")
	}
	return nil
}
