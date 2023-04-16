package mail

import (
	"fmt"
	emailpkg "github.com/jordan-wright/email"
	"net/smtp"
)

const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailContent struct {
	Subject     string
	Content     string
	To          []string
	CC          []string
	BCC         []string
	AttachFiles []string
}

type EmailSender interface {
	SendEmail(emailContent EmailContent) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) GmailSender {
	return GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (s GmailSender) SendEmail(emailContent EmailContent) error {
	email := emailpkg.NewEmail()
	email.From = fmt.Sprintf("%s <%s>", s.name, s.fromEmailAddress)
	email.Subject = emailContent.Subject
	email.HTML = []byte(emailContent.Content)
	email.To = emailContent.To
	email.Cc = emailContent.CC
	email.Bcc = emailContent.BCC

	for _, f := range emailContent.AttachFiles {
		_, err := email.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed To attach file %q: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", s.fromEmailAddress, s.fromEmailPassword, smtpAuthAddress)
	return email.Send(smtpServerAddress, smtpAuth)
}
