package mail

import (
	"reflect"
	"testing"
)

func TestConstants(t *testing.T) {
	if TEXT != "text" {
		t.Errorf("Expected TEXT constant to be 'text', got '%s'", TEXT)
	}
	if HTML != "html" {
		t.Errorf("Expected HTML constant to be 'html', got '%s'", HTML)
	}
}

func TestEmailRequest(t *testing.T) {
	req := EmailRequest{
		From:     "sender@example.com",
		To:       []string{"recv1@test.com", "recv2@test.com"},
		Subject:  "Test Subject",
		Body:     "Email content",
		BodyType: TEXT,
	}

	if req.From != "sender@example.com" {
		t.Errorf("Expected From 'sender@example.com', got '%s'", req.From)
	}

	expectedTo := []string{"recv1@test.com", "recv2@test.com"}
	if !reflect.DeepEqual(req.To, expectedTo) {
		t.Errorf("Expected To %v, got %v", expectedTo, req.To)
	}

	if req.Subject != "Test Subject" {
		t.Errorf("Expected Subject 'Test Subject', got '%s'", req.Subject)
	}

	if req.Body != "Email content" {
		t.Errorf("Expected Body 'Email content', got '%s'", req.Body)
	}

	if req.BodyType != TEXT {
		t.Errorf("Expected BodyType TEXT constant, got '%s'", req.BodyType)
	}
}

func TestEmailResponse(t *testing.T) {
	res := EmailResponse{
		RequestId: "req-12345",
		MessageId: "msg-67890",
	}

	if res.RequestId != "req-12345" {
		t.Errorf("Expected RequestId 'req-12345', got '%s'", res.RequestId)
	}

	if res.MessageId != "msg-67890" {
		t.Errorf("Expected MessageId 'msg-67890', got '%s'", res.MessageId)
	}
}

func TestErrorResponse(t *testing.T) {
	errRes := ErrorResponse{
		Code:    "ERR-100",
		Message: "Invalid recipient address",
	}

	if errRes.Code != "ERR-100" {
		t.Errorf("Expected Code 'ERR-100', got '%s'", errRes.Code)
	}

	if errRes.Message != "Invalid recipient address" {
		t.Errorf("Expected Message 'Invalid recipient address', got '%s'", errRes.Message)
	}
}
