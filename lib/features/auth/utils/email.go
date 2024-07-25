package auth

import "net/smtp"

func SendEmailMessage(toEmail string, message string) error {
	// Authenticate:
	auth := smtp.PlainAuth(
		"",
		"m_ouchene@estin.dz",
		"nmyg najm xgsw ggbu",
		"smtp.gmail.com",
	)

	// Sending Email
	err := smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		"m_ouchene@estin.dz",
		[]string{toEmail},
		[]byte(message),
	)

	return err
}