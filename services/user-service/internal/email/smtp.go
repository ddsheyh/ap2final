package email

import (
	"fmt"
	"log"
	"net/smtp"
)

type SMTPSender struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewSMTPSender(host, port, user, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (s *SMTPSender) SendWelcomeEmail(to, name string) error {
	subject := "Welcome to GoTicket!"
	body := fmt.Sprintf("Hello %s,\n\nWelcome to GoTicket! Your account has been created successfully.\n\nBest regards,\nGoTicket Team", name)
	return s.send(to, subject, body)
}

func (s *SMTPSender) send(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		s.from, to, subject, body)

	auth := smtp.PlainAuth("", s.user, s.password, s.host)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		log.Printf("SMTP send error (to=%s): %v", to, err)
		return fmt.Errorf("send email: %w", err)
	}
	log.Printf("Email sent to %s: %s", to, subject)
	return nil
}
