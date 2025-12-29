// ABOUTME: Asynq task type constants and payload structures.
// ABOUTME: Defines task names and data types for async job processing.

package task

const (
	EmailVerificationCode = "email:verification"
	AnnouncementPublish   = "announcement:publish"
)

type EmailVerificationCodePayload struct {
	To      string
	Subject string
	Content string
}

type AnnouncementPublishPayload struct {
	AnnouncementID uint64 `json:"announcement_id"`
}
