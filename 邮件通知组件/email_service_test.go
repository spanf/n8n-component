package email_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"yourmodule/email"
)

// MockEmailClient 模拟EmailClient接口
type MockEmailClient struct {
	mock.Mock
}

func (m *MockEmailClient) SendRequest(req interface{}) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func TestSendEmail_Success(t *testing.T) {
	mockClient := new(MockEmailClient)
	service := email.EmailService{Client: mockClient}

	mockClient.On("SendRequest", mock.Anything).Return("msg123", nil)

	msgID, err := service.SendEmail(
		"test@example.com",
		[]string{"receiver@example.com"},
		"Subject",
		"Body content",
		"text",
	)

	assert.NoError(t, err)
	assert.Equal(t, "msg123", msgID)
	mockClient.AssertExpectations(t)
}

func TestSendEmail_ClientError(t *testing.T) {
	mockClient := new(MockEmailClient)
	service := email.EmailService{Client: mockClient}

	expectedErr := errors.New("connection failed")
	mockClient.On("SendRequest", mock.Anything).Return("", expectedErr)

	_, err := service.SendEmail(
		"valid@example.com",
		[]string{"valid@example.com"},
		"Subject",
		"Body",
		"html",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email client error")
	mockClient.AssertExpectations(t)
}

func TestSendEmail_ValidationErrors(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          []string
		subject     string
		body        string
		bodyType    string
		expectedErr string
	}{
		{
			name:        "Empty sender",
			from:        "",
			to:          []string{"valid@example.com"},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "sender email cannot be empty",
		},
		{
			name:        "Invalid sender format",
			from:        "invalid-email",
			to:          []string{"valid@example.com"},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "invalid sender email format",
		},
		{
			name:        "Empty recipients",
			from:        "valid@example.com",
			to:          []string{},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "recipient list cannot be empty",
		},
		{
			name:        "Empty recipient email",
			from:        "valid@example.com",
			to:          []string{""},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "recipient email cannot be empty",
		},
		{
			name:        "Invalid recipient format",
			from:        "valid@example.com",
			to:          []string{"invalid-email"},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "invalid recipient email",
		},
		{
			name:        "Empty subject",
			from:        "valid@example.com",
			to:          []string{"valid@example.com"},
			subject:     "",
			body:        "Body",
			bodyType:    "text",
			expectedErr: "email subject cannot be empty",
		},
		{
			name:        "Empty body",
			from:        "valid@example.com",
			to:          []string{"valid@example.com"},
			subject:     "Subject",
			body:        "",
			bodyType:    "text",
			expectedErr: "email body cannot be empty",
		},
		{
			name:        "Invalid body type",
			from:        "valid@example.com",
			to:          []string{"valid@example.com"},
			subject:     "Subject",
			body:        "Body",
			bodyType:    "invalid",
			expectedErr: "bodyType must be either 'text' or 'html'",
		},
	}

	service := email.EmailService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.SendEmail(
				tt.from,
				tt.to,
				tt.subject,
				tt.body,
				tt.bodyType,
			)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"test@example.com", true},
		{"user.name+tag@domain.co.uk", true},
		{"first.last@sub.domain.com", true},
		{"invalid", false},
		{"missing@tld", false},
		{"@missingusername.com", false},
		{"invalid@.com", false},
		{"invalid@domain.", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			assert.Equal(t, tt.valid, email.ValidateEmail(tt.email))
		})
	}
}
