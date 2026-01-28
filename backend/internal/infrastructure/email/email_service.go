package email

import (
	"fmt"
	"net/smtp"
)

// EmailService defines the interface for email operations
type EmailService interface {
	SendVerificationEmail(to, name, token string) error
	SendPasswordResetEmail(to, name, token string) error
}

type emailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	frontendURL  string
}

// NewEmailService creates a new email service
func NewEmailService(
	smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail, fromName, frontendURL string,
) EmailService {
	return &emailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
		frontendURL:  frontendURL,
	}
}

// SendVerificationEmail sends an email verification link to the user
func (s *emailService) SendVerificationEmail(to, name, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, token)
	
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Welcome to TkhanChat, %s!</h2>
			<p>Thank you for signing up. Please verify your email address by clicking the link below:</p>
			<p><a href="%s" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Verify Email</a></p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 24 hours.</p>
			<p>If you didn't create an account, please ignore this email.</p>
		</body>
		</html>
	`, name, verificationURL, verificationURL)

	return s.sendEmail(to, subject, body)
}

// SendPasswordResetEmail sends a password reset link to the user
func (s *emailService) SendPasswordResetEmail(to, name, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, token)
	
	subject := "Reset Your Password"
	body := fmt.Sprintf(`
		<html>
		<body>
			<h2>Password Reset Request</h2>
			<p>Hi %s,</p>
			<p>We received a request to reset your password. Click the link below to reset it:</p>
			<p><a href="%s" style="background-color: #2196F3; color: white; padding: 14px 20px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a></p>
			<p>Or copy and paste this link into your browser:</p>
			<p>%s</p>
			<p>This link will expire in 1 hour.</p>
			<p>If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
		</body>
		</html>
	`, name, resetURL, resetURL)

	return s.sendEmail(to, subject, body)
}

// sendEmail sends an email using SMTP
func (s *emailService) sendEmail(to, subject, body string) error {
	// Build email message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// SMTP authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// MockEmailService is a mock implementation for testing/development
type MockEmailService struct{}

// NewMockEmailService creates a mock email service
func NewMockEmailService() EmailService {
	return &MockEmailService{}
}

// SendVerificationEmail logs the verification email instead of sending
func (m *MockEmailService) SendVerificationEmail(to, name, token string) error {
	fmt.Printf("[MOCK EMAIL] Verification email to %s (%s)\nToken: %s\n", to, name, token)
	return nil
}

// SendPasswordResetEmail logs the password reset email instead of sending
func (m *MockEmailService) SendPasswordResetEmail(to, name, token string) error {
	fmt.Printf("[MOCK EMAIL] Password reset email to %s (%s)\nToken: %s\n", to, name, token)
	return nil
}
