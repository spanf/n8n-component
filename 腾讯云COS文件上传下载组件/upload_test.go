package cosutil

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// 模拟Transport用于拦截请求
type mockTransport struct {
	err      error
	resp     *http.Response
	putCount int
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.putCount++
	return m.resp, m.err
}

func createMockClient(t *testing.T, transport *mockTransport) *cos.Client {
	u, _ := url.Parse("https://testbucket-1250000000.cos.ap-guangzhou.myqcloud.com")
	b := &cos.BaseURL{BucketURL: u}
	return cos.NewClient(b, &http.Client{Transport: transport})
}

func TestUploadFile_Success(t *testing.T) {
	// 创建临时测试文件
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建模拟客户端
	transport := &mockTransport{
		resp: &http.Response{StatusCode: http.StatusOK},
	}
	client := createMockClient(t, transport)

	// 测试上传
	if err := UploadFile(client, filePath, "test.txt"); err != nil {
		t.Errorf("UploadFile failed: %v", err)
	}
	if transport.putCount != 1 {
		t.Error("Put method not called")
	}
}

func TestUploadFile_OpenError(t *testing.T) {
	// 创建模拟客户端
	transport := &mockTransport{}
	client := createMockClient(t, transport)

	// 测试打开不存在的文件
	err := UploadFile(client, "non_existent_file.txt", "test.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if transport.putCount != 0 {
		t.Error("Put should not be called on file open error")
	}
}

func TestUploadFile_StatError(t *testing.T) {
	// 创建临时测试文件
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建模拟客户端
	transport := &mockTransport{}
	client := createMockClient(t, transport)

	// 删除文件使其stat失败
	os.Remove(filePath)

	// 测试stat错误
	err := UploadFile(client, filePath, "test.txt")
	if err == nil {
		t.Error("Expected stat error")
	}
	if transport.putCount != 0 {
		t.Error("Put should not be called on stat error")
	}
}

func TestUploadFile_NetworkError(t *testing.T) {
	// 创建临时测试文件
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建模拟网络错误的客户端
	transport := &mockTransport{err: errors.New("network error")}
	client := createMockClient(t, transport)

	// 测试网络错误
	err := UploadFile(client, filePath, "test.txt")
	if err == nil {
		t.Error("Expected network error")
	}
	if transport.putCount != 1 {
		t.Error("Put method should be called")
	}
}

func TestUploadBytes_Success(t *testing.T) {
	// 创建模拟客户端
	transport := &mockTransport{
		resp: &http.Response{StatusCode: http.StatusOK},
	}
	client := createMockClient(t, transport)

	// 测试上传字节
	data := []byte("test content")
	if err := UploadBytes(client, data, "test.txt"); err != nil {
		t.Errorf("UploadBytes failed: %v", err)
	}
	if transport.putCount != 1 {
		t.Error("Put method not called")
	}
}

func TestUploadBytes_NetworkError(t *testing.T) {
	// 创建模拟网络错误的客户端
	transport := &mockTransport{err: errors.New("network error")}
	client := createMockClient(t, transport)

	// 测试网络错误
	data := []byte("test content")
	err := UploadBytes(client, data, "test.txt")
	if err == nil {
		t.Error("Expected network error")
	}
	if transport.putCount != 1 {
		t.Error("Put method should be called")
	}
}

func TestUploadBytes_ServerError(t *testing.T) {
	// 创建模拟服务端错误的客户端
	transport := &mockTransport{
		resp: &http.Response{StatusCode: http.StatusInternalServerError},
	}
	client := createMockClient(t, transport)

	// 测试服务端错误
	data := []byte("test content")
	err := UploadBytes(client, data, "test.txt")
	if err == nil {
		t.Error("Expected server error")
	}
	if transport.putCount != 1 {
		t.Error("Put method should be called")
	}
}
