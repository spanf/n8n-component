package sms

import (
	"testing"
)

func TestSMSClientInitialization(t *testing.T) {
	client := SMSClient{
		secretId:  "test_id",
		secretKey: "test_key",
	}

	if client.secretId != "test_id" {
		t.Errorf("Expected secretId 'test_id', got '%s'", client.secretId)
	}
	if client.secretKey != "test_key" {
		t.Errorf("Expected secretKey 'test_key', got '%s'", client.secretKey)
	}
}

func TestSendSMSResponseFields(t *testing.T) {
	resp := SendSMSResponse{
		RequestId: "req_123",
		Code:      "200",
		Message:   "Success",
	}

	if resp.RequestId != "req_123" {
		t.Errorf("Expected RequestId 'req_123', got '%s'", resp.RequestId)
	}
	if resp.Code != "200" {
		t.Errorf("Expected Code '200', got '%s'", resp.Code)
	}
	if resp.Message != "Success" {
		t.Errorf("Expected Message 'Success', got '%s'", resp.Message)
	}
}

func TestSendSMSResponseEmptyValues(t *testing.T) {
	resp := SendSMSResponse{}

	if resp.RequestId != "" {
		t.Errorf("Expected empty RequestId, got '%s'", resp.RequestId)
	}
	if resp.Code != "" {
		t.Errorf("Expected empty Code, got '%s'", resp.Code)
	}
	if resp.Message != "" {
		t.Errorf("Expected empty Message, got '%s'", resp.Message)
	}
}
