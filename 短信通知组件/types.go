package sms

type SMSClient struct {
	secretId  string
	secretKey string
}

type SendSMSResponse struct {
	RequestId string
	Code      string
	Message   string
}
