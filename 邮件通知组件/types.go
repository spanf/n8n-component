package mail

const (
	TEXT = "text"
	HTML = "html"
)

type EmailRequest struct {
	From     string
	To       []string
	Subject  string
	Body     string
	BodyType string
}

type EmailResponse struct {
	RequestId string
	MessageId string
}

type ErrorResponse struct {
	Code    string
	Message string
}
