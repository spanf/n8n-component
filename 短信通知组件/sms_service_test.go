package sms

import (
	"encoding/json"
	"errors"
	"testing"
)

type mockClient struct {
	SMSClient
	mockResponse []byte
	mockError    error
}

func (m *mockClient) sendRequest(params map[string]string) ([]byte, error) {
	return m.mockResponse, m.mockError
}

func TestSendSMS_Success(t *testing.T) {
	client := &mockClient{
		mockResponse: []byte(`{"Response":{"Error":{}}}`),
	}

	err := client.SendSMS("tpl123", "13800138000", "TestSign", []string{"param1", "param2"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestSendSMS_RequestError(t *testing.T) {
	expectedErr := errors.New("network error")
	client := &mockClient{
		mockError: expectedErr,
	}

	err := client.SendSMS("tpl123", "13800138000", "TestSign", nil)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestSendSMS_InvalidJSON(t *testing.T) {
	client := &mockClient{
		mockResponse: []byte("invalid json"),
	}

	err := client.SendSMS("tpl123", "13800138000", "TestSign", nil)
	if err == nil || err.Error() != "failed to parse response: invalid character 'i' looking for beginning of value" {
		t.Errorf("Expected JSON parse error, got '%v'", err)
	}
}

func TestSendSMS_APIError(t *testing.T) {
	resp := struct {
		Response struct {
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}{}
	resp.Response.Error.Code = "LimitExceeded"
	resp.Response.Error.Message = "Daily limit reached"
	
	jsonData, _ := json.Marshal(resp)
	client := &mockClient{
		mockResponse: jsonData,
	}

	err := client.SendSMS("tpl123", "13800138000", "TestSign", []string{"param"})
	if err == nil || err.Error() != "Daily limit reached" {
		t.Errorf("Expected API error 'Daily limit reached', got '%v'", err)
	}
}

func TestBuildSendSmsParams(t *testing.T) {
	params := buildSendSmsParams("TPL001", "13800138000", "MySign", []string{"val1", "val2"})
	
	testCases := []struct {
		key    string
		expect string
	}{
		{"TemplateId", "TPL001"},
		{"PhoneNumberSet.0", "13800138000"},
		{"SignName", "MySign"},
		{"TemplateParamSet.0", "val1"},
		{"TemplateParamSet.1", "val2"},
	}
	
	for _, tc := range testCases {
		if val, ok := params[tc.key]; !ok || val != tc.expect {
			t.Errorf("For key %s expected %s, got %v", tc.key, tc.expect, val)
		}
	}
	
	if len(params) != 5 {
		t.Errorf("Expected 5 parameters, got %d", len(params))
	}
}

func TestBuildSendSmsParams_NoTemplateParams(t *testing.T) {
	params := buildSendSmsParams("TPL001", "13800138000", "MySign", nil)
	
	if _, ok := params["TemplateParamSet.0"]; ok {
		t.Error("TemplateParamSet should not exist when no template params provided")
	}
	
	if len(params) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(params))
	}
}
