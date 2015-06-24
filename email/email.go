package email

import (
	"bytes"
	"net/smtp"
	"text/template"

	. "vertigo/misc"
)

type Email struct {
	Host      string
	Recipient RecipientStruct
}

type RecipientStruct struct {
	ID          string
	Address     string
	RecoveryKey string
}

const templ = `
From: <postmaster@{{ .Host }}>
To: <{{ .Recipient.Address }}>
Subject: Password Reset <postmaster@{{ .Host }}>
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

Somebody requested password recovery on this email. You may reset your password through this link: {{ .Host }}user/reset/{{ .Recipient.ID }}/{{ .Recipient.RecoveryKey }}
`

func SendRecoveryEmail(id, address, recovery string) error {

	t, err := template.New("mail").Parse(templ)
	if err != nil {
		return err
	}

	var email Email
	email.Host = UrlHost()
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
		"postmaster@example.com",
		"password",
		"smtp.example.com",
	)

	err = smtp.SendMail(
		"smtp.example.com:587",
		auth,
		"postmaster@example.com",
		[]string{address},
		buf.Bytes(),
	)
	if err != nil {
		return err
	}
	return nil
}
