package utils

import (
	"os"

	"gopkg.in/gomail.v2"
)

func SendEmail(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_SENDER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		587,
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
}
