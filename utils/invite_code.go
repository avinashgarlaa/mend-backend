package utils

import (
	"net/smtp"

	"github.com/google/uuid"
)

// GeneratePartnerID returns a unique UUID for user/session/partner IDs
func GeneratePartnerID() string {
	return uuid.NewString()
}

func SendEmail(to, subject, body string) error {
	from := "your@email.com"
	password := "your_email_password" // App password if using Gmail

	// Set up authentication.
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	// Construct the message
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	// Send the email
	return smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, msg)
}
