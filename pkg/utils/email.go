package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail string, otp string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	subject := "Your Unidate OTP Code"
	body := fmt.Sprintf("Your OTP is: %s. It is valid for 10 minutes.", otp)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	addr := smtpHost + ":" + smtpPort
	return smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
}
