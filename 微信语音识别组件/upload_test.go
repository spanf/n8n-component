package wechat_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestUploadVoiceFile(t *testing.T) {
	t.Run("成功上传文件", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		originalClient := http.DefaultClient
		http.DefaultClient = &http.Client{Transport: mockClient}
		defer func() { http.DefaultClient = originalClient }()

		tempFile, err := os.CreateTemp("", "test_voice.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		_, err = tempFile.WriteString("test audio content")
		require.NoError(t, err)
		tempFile.Close()

		expectedMediaID := "test_media_id"
		responseBody := fmt.Sprintf(`{"errcode":0, "errmsg":"ok", "media_id":"%s"}`, expectedMediaID)
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		}
		mockClient.On("Do", mock.Anything).Return(resp, nil)

		mediaID, err := UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.NoError(t, err)
		assert.Equal(t, expectedMediaID, mediaID)
	})

	t.Run("文件不存在", func(t *testing.T) {
		_, err := UploadVoiceFile("non_existent_file.amr", "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件不存在")
	})

	t.Run("文件是目录", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "test_dir")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		_, err = UploadVoiceFile(tempDir, "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "路径指向目录而非文件")
	})

	t.Run("空文件", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "empty_file.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		_, err = UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件内容为空")
	})

	t.Run("HTTP请求失败", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		originalClient := http.DefaultClient
		http.DefaultClient = &http.Client{Transport: mockClient}
		defer func() { http.DefaultClient = originalClient }()

		tempFile, err := os.CreateTemp("", "test_voice.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		mockClient.On("Do", mock.Anything).Return(
			&http.Response{},
			errors.New("connection error"),
		)

		_, err = UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "请求发送失败")
	})

	t.Run("微信服务器错误响应", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		originalClient := http.DefaultClient
		http.DefaultClient = &http.Client{Transport: mockClient}
		defer func() { http.DefaultClient = originalClient }()

		tempFile, err := os.CreateTemp("", "test_voice.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		resp := &http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       io.NopCloser(bytes.NewBufferString("server error")),
		}
		mockClient.On("Do", mock.Anything).Return(resp, nil)

		_, err = UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "微信服务器错误")
	})

	t.Run("微信API返回错误码", func(t *testing.T) {
		mockClient := new(MockHTTPClient)
		originalClient := http.DefaultClient
		http.DefaultClient = &http.Client{Transport: mockClient}
		defer func() { http.DefaultClient = originalClient }()

		tempFile, err := os.CreateTemp("", "test_voice.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		responseBody := `{"errcode":40001, "errmsg":"invalid credential"}`
		resp := &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		}
		mockClient.On("Do", mock.Anything).Return(resp, nil)

		_, err = UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "微信接口错误[40001]")
	})

	t.Run("构建multipart请求失败", func(t *testing.T) {
		// 使用无效的文件路径触发错误
		invalidPath := string([]byte{0x7f}) + "invalid_path"
		_, err := UploadVoiceFile(invalidPath, "test_token", "voice")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "构建请求失败")
	})
}

func TestValidateFile(t *testing.T) {
	t.Run("有效文件", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "valid_file.*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		_, err = tempFile.WriteString("test content")
		require.NoError(t, err)
		tempFile.Close()

		valid, err := validateFile(tempFile.Name())
		assert.True(t, valid)
		assert.NoError(t, err)
	})

	t.Run("文件不存在", func(t *testing.T) {
		valid, err := validateFile("non_existent_file.txt")
		assert.False(t, valid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件不存在")
	})

	t.Run("空文件", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "empty_file.*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		valid, err := validateFile(tempFile.Name())
		assert.False(t, valid)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件内容为空")
	})
}

func TestBuildUploadRequest(t *testing.T) {
	t.Run("成功构建请求", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "upload_file.*.txt")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		req, err := buildUploadRequest(tempFile.Name(), "test_token", "voice")
		require.NoError(t, err)
		assert.NotNil(t, req)

		// 验证请求头
		contentType := req.Header.Get("Content-Type")
		assert.Contains(t, contentType, "multipart/form-data; boundary=")

		// 验证请求体
		bodyBytes, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(bodyBytes), "Content-Disposition: form-data; name=\"media\";")
	})

	t.Run("文件打开失败", func(t *testing.T) {
		_, err := buildUploadRequest("non_existent_file.txt", "test_token", "voice")
		assert.Error(t, err)
	})
}

func TestUploadIntegration(t *testing.T) {
	t.Run("完整集成测试", func(t *testing.T) {
		// 创建测试服务器
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 验证请求内容类型
			contentType := r.Header.Get("Content-Type")
			if !bytes.HasPrefix([]byte(contentType), []byte("multipart/form-data")) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// 解析multipart表单
			err := r.ParseMultipartForm(10 << 20)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// 验证文件存在
			file, _, err := r.FormFile("media")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer file.Close()

			// 模拟成功响应
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"errcode":0, "errmsg":"ok", "media_id":"test_media_id"}`))
		}))
		defer server.Close()

		// 临时文件
		tempFile, err := os.CreateTemp("", "integration_test.*.amr")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())
		_, err = tempFile.WriteString("test audio content")
		require.NoError(t, err)
		tempFile.Close()

		// 替换URL
		originalBuildFunc := buildUploadRequest
		buildUploadRequest = func(filePath, accessToken, format string) (*http.Request, error) {
			req, err := originalBuildFunc(filePath, accessToken, format)
			if err != nil {
				return nil, err
			}
			req.URL, err = req.URL.Parse(server.URL + "?" + req.URL.RawQuery)
			if err != nil {
				return nil, err
			}
			return req, nil
		}
		defer func() { buildUploadRequest = originalBuildFunc }()

		mediaID, err := UploadVoiceFile(tempFile.Name(), "test_token", "voice")
		assert.NoError(t, err)
		assert.Equal(t, "test_media_id", mediaID)
	})
}
