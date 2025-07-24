package wechat_template_message_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wechat_template_message"
)

func TestGetAccessToken(t *testing.T) {
	type testCase struct {
		name           string
		appID          string
		appSecret      string
		statusCode     int
		response       string
		wantToken      string
		wantErr        bool
		wantErrMsg     string
		partialErrMatch bool
	}

	testCases := []testCase{
		{
			name:       "成功获取",
			appID:      "valid_appid",
			appSecret:  "valid_secret",
			statusCode: http.StatusOK,
			response:   `{"access_token":"test_token_123","expires_in":7200}`,
			wantToken:  "test_token_123",
			wantErr:    false,
		},
		{
			name:       "微信接口错误",
			appID:      "invalid_appid",
			appSecret:  "invalid_secret",
			statusCode: http.StatusOK,
			response:   `{"errcode":40013,"errmsg":"invalid appid"}`,
			wantErr:    true,
			wantErrMsg: "微信接口错误[40013]: invalid appid",
		},
		{
			name:       "HTTP状态码异常",
			appID:      "test_appid",
			appSecret:  "test_secret",
			statusCode: http.StatusInternalServerError,
			response:   "Internal Server Error",
			wantErr:    true,
			wantErrMsg: "HTTP状态码异常: 500",
		},
		{
			name:           "JSON解析失败",
			appID:          "test_appid",
			appSecret:      "test_secret",
			statusCode:     http.StatusOK,
			response:       `{"access_token":"test_token", "expires_in":7200`, // 无效JSON
			wantErr:        true,
			wantErrMsg:     "JSON解析失败",
			partialErrMatch: true,
		},
		{
			name:       "AccessToken为空",
			appID:      "test_appid",
			appSecret:  "test_secret",
			statusCode: http.StatusOK,
			response:   `{"access_token":"","expires_in":7200}`,
			wantErr:    true,
			wantErrMsg: "获取的AccessToken为空",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(tc.response))
			}))
			defer ts.Close()

			oldClient := http.DefaultClient
			http.DefaultClient = ts.Client()
			defer func() { http.DefaultClient = oldClient }()

			token, err := wechat_template_message.GetAccessToken(tc.appID, tc.appSecret)

			if tc.wantErr {
				if err == nil {
					t.Fatal("预期返回错误，实际返回nil")
				}
				if tc.partialErrMatch {
					if !strings.Contains(err.Error(), tc.wantErrMsg) {
						t.Errorf("预期错误信息包含 %q, 实际得到 %q", tc.wantErrMsg, err.Error())
					}
				} else if err.Error() != tc.wantErrMsg {
					t.Errorf("预期错误信息 %q, 实际得到 %q", tc.wantErrMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("预期成功，实际返回错误: %v", err)
				}
				if token != tc.wantToken {
					t.Errorf("预期token %q, 实际得到 %q", tc.wantToken, token)
				}
			}
		})
	}
}

func TestTokenResponseUnmarshal(t *testing.T) {
	t.Run("空access_token处理", func(t *testing.T) {
		var resp wechat_template_message.tokenResponse
		err := json.Unmarshal([]byte(`{"access_token":"","expires_in":7200}`), &resp)
		if err != nil {
			t.Fatalf("JSON解析失败: %v", err)
		}
		if resp.AccessToken != "" {
			t.Errorf("预期空access_token，实际得到 %q", resp.AccessToken)
		}
	})

	t.Run("错误码解析", func(t *testing.T) {
		var resp wechat_template_message.tokenResponse
		err := json.Unmarshal([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`), &resp)
		if err != nil {
			t.Fatalf("JSON解析失败: %v", err)
		}
		if resp.ErrCode != 40013 || resp.ErrMsg != "invalid appid" {
			t.Errorf("预期错误码40013/invalid appid，实际得到 %d/%s", resp.ErrCode, resp.ErrMsg)
		}
	})
}
