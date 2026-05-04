package email

import (
	"fmt"
	"net/smtp"
)

type Sender interface {
	SendOTP(toEmail string, otp string) error
}

type gmailSender struct {
	fromEmail string
	password  string
}

func NewGmailSender(from string, password string) Sender {
	return &gmailSender{
		fromEmail: from,
		password:  password,
	}
}

func (s *gmailSender) SendOTP(toEmail string, otp string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 1. Define your App's Display Name
	appName := "Dreams App"

	// 2. Build the Email Headers
	// The format MUST be: From: Display Name <email@domain.com>\r\n
	headerFrom := fmt.Sprintf("From: %s <%s>\r\n", appName, s.fromEmail)
	headerTo := fmt.Sprintf("To: %s\r\n", toEmail)
	headerSubject := "Subject: Dreams E-commerce Login OTP\r\n"
	
	// Add MIME headers so email clients parse the text cleanly
	headerMIME := "MIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n"

	// 3. Build the Email Body
	// Notice the two \r\n\r\n at the very start of the body. 
	// This blank line tells the email client "Headers are done, the body starts here!"
	body := fmt.Sprintf("\r\nHello!\r\n\r\nYour one-time password (OTP) is: %s\r\n\r\nThis code will expire in 5 minutes.\r\n\r\nDo not share this code with anyone.", otp)

	// 4. Combine Headers and Body into a single byte array
	message := []byte(headerFrom + headerTo + headerSubject + headerMIME + body)

	// 5. Authenticate and Send
	auth := smtp.PlainAuth("", s.fromEmail, s.password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, s.fromEmail, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}