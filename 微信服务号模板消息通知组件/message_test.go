package wechat_template_message_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"wechat_template_message"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockHTTPClient struct {
	mock.Mock
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestSendTemplateMessage_Success(t *testing.T) {
	mockClient := new(mockHTTPClient)
	wechat_template_message.Client = mockClient

	expectedMsgID := int64(123456)
	respBody := fmt.Sprintf(`{"errcode":0,"errmsg":"ok","msgid":%d}`, expectedMsgID)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(respBody)),
	}

	mockClient.On("Do", mock.Anything).Return(response, nil)

	miniprogram := map[string]string{"appid": "test_app", "pagepath": "home"}
	data := map[string]interface{}{"key1": "value1"}
	msgID, err := wechat_template_message.SendTemplateMessage("token", "open123", "tpl123", data, "https://example.com", miniprogram)

	assert.NoError(t, err)
	assert.Equal(t, expectedMsgID, msgID)
	mockClient.AssertExpectations(t)
}

func TestSendTemplateMessage_HTTPError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	wechat_template_message.Client = mockClient

	response := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	mockClient.On("Do", mock.Anything).Return(response, nil)

	_, err := wechat_template_message.SendTemplateMessage("token", "open123", "tpl123", nil, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "http status: 500")
	mockClient.AssertExpectations(t)
}

func TestSendTemplateMessage_RequestError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	wechat_template_message.Client = mockClient

	mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("network error"))

	_, err := wechat_template_message.SendTemplateMessage("token", "open123", "tpl123", nil, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "send request failed")
	mockClient.AssertExpectations(t)
}

func TestSendTemplateMessage_InvalidResponse(t *testing.T) {
	mockClient := new(mockHTTPClient)
	wechat_template_message.Client = mockClient

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString("invalid json")),
	}
	mockClient.On("Do", mock.Anything).Return(response, nil)

	_, err := wechat_template_message.SendTemplateMessage("token", "open123", "tpl123", nil, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse response failed")
	mockClient.AssertExpectations(t)
}

func TestSendTemplateMessage_WechatAPIError(t *testing.T) {
	mockClient := new(mockHTTPClient)
	wechat_template_message.Client = mockClient

	respBody := `{"errcode":40001,"errmsg":"invalid credential"}`
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(respBody)),
	}
	mockClient.On("Do", mock.Anything).Return(response, nil)

	_, err := wechat_template_message.SendTemplateMessage("token", "open123", "tpl123", nil, "", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "40001")
	assert.Contains(t, err.Error(), "invalid credential")
	mockClient.AssertExpectations(t)
}

func TestBuildMessageBody(t *testing.T) {
	t.Run("WithMiniprogram", func(t *testing.T) {
		miniprogram := map[string]string{"appid": "app123", "pagepath": "home"}
		body, err := wechat_template_message.BuildMessageBody("open123", "tpl123", nil, "", miniprogram)
		assert.NoError(t, err)

		var msg wechat_template_message.MessageRequest
		err = json.Unmarshal(body, &msg)
		assert.NoError(t, err)
		assert.Equal(t, "app123", msg.MiniProgram.AppID)
		assert.Equal(t, "home", msg.MiniProgram.PagePath)
	})

	t.Run("WithURL", func(t *testing.T) {
		body, err := wechat_template_message.BuildMessageBody("open123", "tpl123", nil, "https://test.com", nil)
		assert.NoError(t, err)

		var msg wechat_template_message.MessageRequest
		err = json.Unmarshal(body, &msg)
		assert.NoError(t, err)
		assert.Equal(t, "https://test.com", msg.URL)
	})

	t.Run("WithData", func(t *testing.T) {
		data := map[string]interface{}{"name": "value"}
		body, err := wechat_template_message.BuildMessageBody("open123", "tpl123", data, "", nil)
		assert.NoError(t, err)

		var msg wechat_template_message.MessageRequest
		err = json.Unmarshal(body, &msg)
		assert.NoError(t, err)
		assert.Equal(t, data, msg.Data)
	})
}

func TestSendMessageRequest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("response data"))
		}))
		defer server.Close()

		apiURL := server.URL
		body := []byte("test body")
		respBody, err := wechat_template_message.SendMessageRequest(apiURL, body)

		assert.NoError(t, err)
		assert.Equal(t, []byte("response data"), respBody)
	})

	t.Run("HTTPError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		apiURL := server.URL
		body := []byte("test body")
		_, err := wechat_template_message.SendMessageRequest(apiURL, body)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http status: 400")
	})
}
