package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type mockObjectService struct {
	PutFunc func(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error)
}

func (m *mockObjectService) Put(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error) {
	return m.PutFunc(ctx, key, r, opt)
}

func TestUploadFile_Success(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString("test content")
	tmpfile.Close()

	mockObj := &mockObjectService{
		PutFunc: func(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error) {
			return &cos.Response{Response: &http.Response{StatusCode: 200}}, nil
		},
	}

	client := &cos.Client{Object: mockObj}
	err = UploadFile(client, "test-bucket", "test/path", tmpfile.Name())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestUploadFile_FileOpenError(t *testing.T) {
	client := &cos.Client{}
	err := UploadFile(client, "test-bucket", "test/path", "non_existent_file")
	if !errors.Is(err, os.ErrInvalid) {
		t.Errorf("Expected os.ErrInvalid, got %v", err)
	}
}

func TestUploadFile_PutError(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	mockErr := errors.New("mock put error")
	mockObj := &mockObjectService{
		PutFunc: func(ctx context.Context, key string, r io.Reader, opt *cos.ObjectPutOptions) (*cos.Response, error) {
			return nil, mockErr
		},
	}

	client := &cos.Client{Object: mockObj}
	err = UploadFile(client, "test-bucket", "test/path", tmpfile.Name())
	if err != mockErr {
		t.Errorf("Expected %v, got %v", mockErr, err)
	}
}

func TestUploadFile_StatError(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	originalOpen := osOpen
	defer func() { osOpen = originalOpen }()
	osOpen = func(name string) (*os.File, error) {
		return os.OpenFile(name, os.O_RDWR, 0644)
	}

	originalStat := osStat
	defer func() { osStat = originalStat }()
	osStat = func(name string) (os.FileInfo, error) {
		return nil, errors.New("stat error")
	}

	client := &cos.Client{}
	err = UploadFile(client, "test-bucket", "test/path", tmpfile.Name())
	if !errors.Is(err, os.ErrInvalid) {
		t.Errorf("Expected os.ErrInvalid, got %v", err)
	}
}
