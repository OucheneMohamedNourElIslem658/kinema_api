package auth

import "net/smtp"

func SendEmailMessage(toEmail string, message string) error {
	// Authenticate:
	auth := smtp.PlainAuth(
		"",
		***,
		***,
		"smtp.gmail.com",
	)

	// Sending Email
	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		***,
		[]string{toEmail},
		[]byte(message),
	)

	return err
}
