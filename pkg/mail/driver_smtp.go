package mail

import (
	"crypto/tls"
	"fmt"
	"net/smtp"

	"github.com/jordan-wright/email"
)

type SMTP struct {
	*Options
}

func (s *SMTP) Send(to string, subject string, content string) error {
	// New instance
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.FromName, s.FromAddr)
	e.To = []string{to}
	e.Bcc = []string{}
	e.Cc = []string{}
	e.Subject = subject
	e.Text = []byte(content)

	// Addr & Auth
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	// Send Email
	return e.SendWithTLS(addr, auth, &tls.Config{ServerName: s.Host})
}
