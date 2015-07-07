package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"

	. "github.com/9uuso/vertigo/settings"
)

type Email struct {
	Sender    string
	Host      string
	Recipient RecipientStruct
}

type RecipientStruct struct {
	ID          string
	Address     string
	RecoveryKey string
}

const templ = `
From: <{{ .Sender }}>
To: <{{ .Recipient.Address }}>
Subject: Password Reset <{{ .Sender }}>
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

Somebody requested password recovery on this email. You may reset your password through this link: {{ .Host }}/user/reset/{{ .Recipient.ID }}/{{ .Recipient.RecoveryKey }}
`

func SendRecoveryEmail(id, address, recovery string) error {

	t, err := template.New("mail").Parse(templ)
	if err != nil {
		return err
	}

	var email Email
	email.Sender = Settings.Mailer.Login
	email.Host = Settings.URL.String()
	email.Recipient.ID = id
	email.Recipient.Address = address
	email.Recipient.RecoveryKey = recovery

	var buf bytes.Buffer
	err = t.Execute(&buf, email)
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth(
		"",
		Settings.Mailer.Login,
		Settings.Mailer.Password,
		Settings.Mailer.Hostname,
	)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", Settings.Mailer.Hostname, Settings.Mailer.Port),
		auth,
		Settings.Mailer.Login,
		[]string{address},
		buf.Bytes(),
	)
	if err != nil {
		return err
	}
	return nil
}
