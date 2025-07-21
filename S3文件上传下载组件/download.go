package s3util

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

func DownloadFile(client *s3.S3, bucket, key string, w io.WriterAt) error {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return err
	}
	defer result.Body.Close()

	buf := make([]byte, 1024*1024) // 1MB buffer
	var offset int64 = 0

	for {
		n, readErr := result.Body.Read(buf)
		if n > 0 {
			_, writeErr := w.WriteAt(buf[:n], offset)
			if writeErr != nil {
				return writeErr
			}
			offset += int64(n)
		}

		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			}
			break
		}
	}

	return nil
}

func checkObjectExists(client *s3.S3, bucket, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := client.HeadObject(context.TODO(), input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "NotFound", "NoSuchKey":
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}
