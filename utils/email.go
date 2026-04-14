package utils

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

func SendEmail(to, subject, resetCode string) error {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	username := strings.TrimSpace(os.Getenv("SMTP_USER"))
	password := os.Getenv("SMTP_PASS")

	if host == "" || port == "" || from == "" {
		return fmt.Errorf("smtp is not configured")
	}

	message := []byte(
		"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n\r\n" +
			"Your password reset code is: " + resetCode + "\r\n" +
			"This code will expire soon.\r\n",
	)

	var auth smtp.Auth
	if username != "" || password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}

	return smtp.SendMail(host+":"+port, auth, from, []string{to}, message)
}
