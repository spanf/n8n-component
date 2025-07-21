package s3util_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/stretchr/testify/require"

	"yourmodulepath/s3util"
)

type mockS3Client struct {
	getObjectOutput *s3.GetObjectOutput
	getObjectError  error
	headObjectError error
}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return m.getObjectOutput, m.getObjectError
}

func (m *mockS3Client) HeadObject(ctx context.Context, input *s3.HeadObjectInput, opts ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	return nil, m.headObjectError
}

type mockWriterAt struct {
	buf     []byte
	writeAt func(p []byte, off int64) (n int, err error)
}

func (m *mockWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	return m.writeAt(p, off)
}

func TestDownloadFile_Success(t *testing.T) {
	client := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader("test content")),
		},
	}

	buf := new(bytes.Buffer)
	writer := &mockWriterAt{
		buf: make([]byte, 1024),
		writeAt: func(p []byte, off int64) (int, error) {
			return buf.Write(p)
		},
	}

	err := s3util.DownloadFile(client, "bucket", "key", writer)
	require.NoError(t, err)
	require.Equal(t, "test content", buf.String())
}

func TestDownloadFile_GetObjectError(t *testing.T) {
	client := &mockS3Client{
		getObjectError: errors.New("get object error"),
	}

	writer := &mockWriterAt{
		buf: make([]byte, 1024),
	}

	err := s3util.DownloadFile(client, "bucket", "key", writer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "get object error")
}

func TestDownloadFile_ReadError(t *testing.T) {
	client := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: io.NopCloser(&errorReader{err: errors.New("read error")}),
		},
	}

	writer := &mockWriterAt{
		buf: make([]byte, 1024),
	}

	err := s3util.DownloadFile(client, "bucket", "key", writer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "read error")
}

func TestDownloadFile_WriteError(t *testing.T) {
	client := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader("test")),
		},
	}

	writer := &mockWriterAt{
		writeAt: func(p []byte, off int64) (int, error) {
			return 0, errors.New("write error")
		},
	}

	err := s3util.DownloadFile(client, "bucket", "key", writer)
	require.Error(t, err)
	require.Contains(t, err.Error(), "write error")
}

func TestCheckObjectExists_Exists(t *testing.T) {
	client := &mockS3Client{
		headObjectError: nil,
	}

	exists, err := s3util.CheckObjectExists(client, "bucket", "key")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestCheckObjectExists_NotExists(t *testing.T) {
	apiErr := &smithy.GenericAPIError{Code: "NotFound"}
	client := &mockS3Client{
		headObjectError: apiErr,
	}

	exists, err := s3util.CheckObjectExists(client, "bucket", "key")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestCheckObjectExists_OtherError(t *testing.T) {
	client := &mockS3Client{
		headObjectError: errors.New("other error"),
	}

	_, err := s3util.CheckObjectExists(client, "bucket", "key")
	require.Error(t, err)
	require.Contains(t, err.Error(), "other error")
}

type errorReader struct {
	err error
}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, e.err
}
