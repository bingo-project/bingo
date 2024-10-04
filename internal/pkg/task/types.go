package task

const (
	EmailVerificationCode = "email:verification"
)

type EmailVerificationCodePayload struct {
	To      string
	Subject string
	Content string
}
