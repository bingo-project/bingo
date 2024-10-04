package bootstrap

import (
	"fmt"

	"bingo/internal/scheduler/facade"
	"bingo/pkg/mail"
)

func InitMail() {
	fmt.Println("init mail", facade.Config.Mail)
	facade.Mail = mail.NewMailer(facade.Config.Mail)
}
