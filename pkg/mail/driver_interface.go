package mail

type Driver interface {
	Send(to string, subject string, content string) error
}
