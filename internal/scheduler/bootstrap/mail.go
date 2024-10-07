package bootstrap

import (
	"bingo/internal/pkg/facade"
	"bingo/pkg/mail"
)

func InitMail() {
	facade.Mail = mail.NewMailer(facade.Config.Mail)
}
