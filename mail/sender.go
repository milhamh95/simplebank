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
	subject     string
	content     string
	to          []string
	cc          []string
	bcc         []string
	attachFiles []string
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
	email.Subject = emailContent.subject
	email.HTML = []byte(emailContent.content)
	email.To = emailContent.to
	email.Cc = emailContent.cc
	email.Bcc = emailContent.bcc

	for _, f := range emailContent.attachFiles {
		_, err := email.AttachFile(f)
		if err != nil {
			return fmt.Errorf("failed to attach file %q: %w", f, err)
		}
	}

	smtpAuth := smtp.PlainAuth("", s.fromEmailAddress, s.fromEmailPassword, smtpAuthAddress)
	return email.Send(smtpServerAddress, smtpAuth)
}
