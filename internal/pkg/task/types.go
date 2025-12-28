package task

const (
	EmailVerificationCode   = "email:verification"
	AnnouncementPublish     = "announcement:publish"
)

type EmailVerificationCodePayload struct {
	To      string
	Subject string
	Content string
}

type AnnouncementPublishPayload struct {
	AnnouncementID uint64 `json:"announcement_id"`
}
