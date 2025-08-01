package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("mock read error")
}

func TestRecognizeVoice_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":0,"text":"测试语音内容"}`))
	}))
	defer ts.Close()

	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()
	http.DefaultTransport = redirectTransport(ts.URL)

	text, err := RecognizeVoice("media123", "token123", "zh_CN")
	if err != nil {
		t.Fatalf("预期成功，但返回错误: %v", err)
	}
	if text != "测试语音内容" {
		t.Errorf("预期文本'测试语音内容'，实际得到'%s'", text)
	}
}

func TestRecognizeVoice_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()
	http.DefaultTransport = redirectTransport(ts.URL)

	_, err := RecognizeVoice("media123", "token123", "zh_CN")
	if err == nil || !strings.Contains(err.Error(), "错误状态码: 500") {
		t.Errorf("预期HTTP错误，实际得到: %v", err)
	}
}

func TestRecognizeVoice_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"errcode":40001,"errmsg":"无效token"}`))
	}))
	defer ts.Close()

	oldTransport := http.DefaultTransport
	defer func() { http.DefaultTransport = oldTransport }()
	http.DefaultTransport = redirectTransport(ts.URL)

	_, err := RecognizeVoice("media123", "token123", "zh_CN")
	if err == nil || !strings.Contains(err.Error(), "无效token") {
		t.Errorf("预期API错误，实际得到: %v", err)
	}
}

func TestParseRecognitionResponse_Success(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"errcode":0,"text":"解析成功"}`)),
	}
	text, err := parseRecognitionResponse(resp)
	if err != nil {
		t.Fatalf("预期成功，但返回错误: %v", err)
	}
	if text != "解析成功" {
		t.Errorf("预期文本'解析成功'，实际得到'%s'", text)
	}
}

func TestParseRecognitionResponse_ReadError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(errorReader{}),
	}
	_, err := parseRecognitionResponse(resp)
	if err == nil || !strings.Contains(err.Error(), "读取响应失败") {
		t.Errorf("预期读取错误，实际得到: %v", err)
	}
}

func TestParseRecognitionResponse_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`invalid json`)),
	}
	_, err := parseRecognitionResponse(resp)
	if err == nil || !strings.Contains(err.Error(), "解析JSON失败") {
		t.Errorf("预期JSON解析错误，实际得到: %v", err)
	}
}

func TestBuildRecognitionRequest_Success(t *testing.T) {
	req, err := buildRecognitionRequest("media123", "token123", "en")
	if err != nil {
		t.Fatalf("构建请求失败: %v", err)
	}
	
	if req.Method != "POST" {
		t.Errorf("预期POST方法，实际得到%s", req.Method)
	}
	
	expectedURL := "https://api.weixin.qq.com/cgi-bin/media/voice/recognize?access_token=token123"
	if req.URL.String() != expectedURL {
		t.Errorf("URL不匹配，预期:%s 实际:%s", expectedURL, req.URL.String())
	}
	
	contentType := req.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type错误，预期application/json，实际%s", contentType)
	}
	
	var payload struct {
		MediaID string `json:"media_id"`
		Lang    string `json:"lang"`
	}
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		t.Fatalf("解析请求体失败: %v", err)
	}
	if payload.MediaID != "media123" || payload.Lang != "en" {
		t.Errorf("请求体错误，预期media_id:media123 lang:en，实际media_id:%s lang:%s", payload.MediaID, payload.Lang)
	}
}

func redirectTransport(targetURL string) http.RoundTripper {
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Host = strings.TrimPrefix(targetURL, "http://")
		req.URL.Scheme = "http"
		return http.DefaultTransport.RoundTrip(req)
	})
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
