package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/resend/resend-go/v3"
)

func SendEmail(to, subject, resetCode string) error {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEYi"))
	from := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))

	if apiKey == "" || apiKey == "re_xxxxxxxxx" {
		return fmt.Errorf("resend api key is not configured; replace re_xxxxxxxxx with your real API key")
	}

	if from == "" {
		from = "onboarding@resend.dev"
	}

	client := resend.NewClient(apiKey)
	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    buildResetEmailHTML(resetCode),
	}

	if _, err := client.Emails.Send(params); err != nil {
		return fmt.Errorf("failed to send email with resend: %w", err)
	}

	return nil
}

func buildResetEmailHTML(resetCode string) string {
	return "<p>Your password reset code is: <strong>" + resetCode + "</strong></p>" +
		"<p>This code will expire soon.</p>"
}
