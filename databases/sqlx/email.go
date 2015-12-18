package sqlx

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"strconv"
	"text/template"
)

// Email holds data of email sender and recipient for easier handling in templates.
type Email struct {
	Sender    string
	Host      string
	Recipient RecipientStruct
}

// RecipientStruct holds data of email recipient for easier handling in templates.
type RecipientStruct struct {
	ID          string
	Name        string
	Address     string
	RecoveryKey string
}

var RecoveryTemplate = `Hello {{ .Recipient.Name }}

Somebody requested password recovery on this email.

You may reset your password through this link: {{ .Host }}/user/reset/{{ .Recipient.ID }}/{{ .Recipient.RecoveryKey }}`

// SendRecoveryEmail dispatches predefined recovery email to recipient defined in parameters.
// Makes use of https://gist.github.com/andelf/5004821
func (user User) SendRecoveryEmail() error {

	var email Email
	email.Sender = Settings.MailerLogin
	email.Host = Settings.Hostname
	email.Recipient.ID = strconv.Itoa(int(user.ID))
	email.Recipient.Name = user.Name
	email.Recipient.Address = user.Email
	email.Recipient.RecoveryKey = user.Recovery

	from := mail.Address{
		Name:    Settings.Name,
		Address: email.Sender,
	}
	to := mail.Address{
		Name:    email.Recipient.Name,
		Address: email.Recipient.Address,
	}
	title := "Password reset"

	t, err := template.New("mail").Parse(RecoveryTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, email)
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = title
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	var message string
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString(buf.Bytes())

	auth := smtp.PlainAuth(
		"",
		Settings.MailerLogin,
		Settings.MailerPassword,
		Settings.MailerHostname,
	)

	err = smtp.SendMail(
		fmt.Sprintf("%s:%d", Settings.MailerHostname, Settings.MailerPort),
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)
	if err != nil {
		return err
	}
	return nil
}
