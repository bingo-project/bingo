package bootstrap

import (
	"bingo/internal/apiserver/facade"
	"bingo/pkg/mail"
)

func InitMail() {
	facade.Mail = mail.NewMailer(facade.Config.Mail)
}
