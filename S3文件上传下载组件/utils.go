package utils

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func GeneratePresignedURL(client *s3.S3, bucket, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(client)
	
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	
	presignParams := &s3.PresignOptions{
		Expires: expiry,
	}
	
	result, err := presignClient.PresignGetObject(context.TODO(), input, presignParams)
	if err != nil {
		return "", parseS3Error(err)
	}
	return result.URL, nil
}

func parseS3Error(err error) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return &S3Error{
			Code:    apiErr.ErrorCode(),
			Message: apiErr.ErrorMessage(),
		}
	}
	return err
}

type S3Error struct {
	Code    string
	Message string
}

func (e *S3Error) Error() string {
	return "S3Error: " + e.Code + " - " + e.Message
}
