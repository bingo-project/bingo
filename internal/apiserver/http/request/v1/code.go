package v1

type SendEmailRequest struct {
	Email string `json:"email" binding:"required,email" example:"peter@gmail.com"`
}
