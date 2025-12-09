package bootstrap

import (
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/pkg/mail"
)

func InitMail() {
	facade.Mail = mail.NewMailer(facade.Config.Mail)
}
