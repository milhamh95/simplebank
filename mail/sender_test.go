package mail

import (
	"github.com/milhamh95/simplebank/pkg/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test TestSendEmailWithGmail in short mode.")
	}

	cfg, err := config.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(cfg.EmailSenderName, cfg.EmailSenderAddress, cfg.EmailSenderPassword)
	subject := "A test email"
	content := `
		<h1>Hi there</h1>
		<p>This is a test email</p>
	`

	to := []string{"123@email.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(EmailContent{
		subject:     subject,
		content:     content,
		to:          to,
		attachFiles: attachFiles,
	})
	require.NoError(t, err)
}
