package email

import (
	"errors"
	"fmt"
	"regexp"
)

// EmailClient 接口定义实际邮件发送功能
type EmailClient interface {
	SendRequest(req interface{}) (string, error)
}

// EmailService 提供邮件发送服务
type EmailService struct {
	Client EmailClient
}

// emailRequest 内部邮件请求结构
type emailRequest struct {
	From     string
	To       []string
	Subject  string
	Body     string
	BodyType string
}

// SendEmail 发送邮件并返回消息ID
func (s *EmailService) SendEmail(from string, to []string, subject, body, bodyType string) (string, error) {
	if err := validateParameters(from, to, subject, body, bodyType); err != nil {
		return "", err
	}

	req := &emailRequest{
		From:     from,
		To:       to,
		Subject:  subject,
		Body:     body,
		BodyType: bodyType,
	}

	msgID, err := s.Client.SendRequest(req)
	if err != nil {
		return "", fmt.Errorf("email client error: %w", err)
	}
	return msgID, nil
}

// validateParameters 验证所有输入参数
func validateParameters(from string, to []string, subject, body, bodyType string) error {
	if from == "" {
		return errors.New("sender email cannot be empty")
	}
	if !validateEmail(from) {
		return errors.New("invalid sender email format")
	}

	if len(to) == 0 {
		return errors.New("recipient list cannot be empty")
	}
	for _, email := range to {
		if email == "" {
			return errors.New("recipient email cannot be empty")
		}
		if !validateEmail(email) {
			return fmt.Errorf("invalid recipient email: %s", email)
		}
	}

	if subject == "" {
		return errors.New("email subject cannot be empty")
	}
	if body == "" {
		return errors.New("email body cannot be empty")
	}
	if bodyType != "text" && bodyType != "html" {
		return errors.New("bodyType must be either 'text' or 'html'")
	}
	return nil
}

// validateEmail 验证邮箱格式
func validateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}
