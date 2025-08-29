package mailer

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"time"

	gomail "gopkg.in/mail.v2"
)

type mailtrapClient struct {
	fromEmail string
	apiKey    string
}

var (
	ErrApiKeyIsRequired = errors.New("api key is required")
)

func NewMailTrapClient(apiKey, fromEmail string) (*mailtrapClient, error) {
	if apiKey == "" {
		return nil, ErrApiKeyIsRequired
	}

	return &mailtrapClient{
		fromEmail: fromEmail,
		apiKey:    apiKey,
	}, nil
}

func (m *mailtrapClient) Send(templateFile, username, email string, data any, isSandbox bool) error {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}

	body := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())

	message.AddAlternative("text/html", body.String())

	var retryErr error
	dialer := gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)

	for i := 0; i < maxRetries; i++ {
		if retryErr = dialer.DialAndSend(message); err != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		return nil
	}

	return fmt.Errorf("failed to send email after %d attempts, error: %v", maxRetries, retryErr)
}
