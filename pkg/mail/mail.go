package mail

import (
	"sync"
)

type Mailer struct {
	Driver Driver
}

type Options struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"-" yaml:"password"`
	FromAddr string `mapstructure:"fromAddr" json:"fromAddr" yaml:"fromAddr"`
	FromName string `mapstructure:"fromName" json:"fromName" yaml:"fromName"`
}

var once sync.Once
var mailer *Mailer

func NewMailer(opts *Options) *Mailer {
	once.Do(func() {
		mailer = &Mailer{
			Driver: &SMTP{opts},
		}
	})

	return mailer
}

func (m *Mailer) Send(to string, subject string, content string) error {
	return m.Driver.Send(to, subject, content)
}
