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
	return &SMTPSender{host: host, port: port, user: user, password: password, from: from}
}

func (s *SMTPSender) SendPaymentConfirmation(to string, orderID int64, amount float64) error {
	subject := "Payment Confirmed — GoTicket"
	body := fmt.Sprintf("Your payment of %.2f for order #%d has been confirmed.\n\nThank you for your purchase!\n\nGoTicket Team", amount, orderID)
	return s.send(to, subject, body)
}

func (s *SMTPSender) SendPaymentFailure(to string, orderID int64, reason string) error {
	subject := "Payment Failed — GoTicket"
	body := fmt.Sprintf("Unfortunately, your payment for order #%d has failed.\nReason: %s\n\nPlease try again or contact support.\n\nGoTicket Team", orderID, reason)
	return s.send(to, subject, body)
}

func (s *SMTPSender) SendRefundConfirmation(to string, orderID int64, amount float64) error {
	subject := "Refund Processed — GoTicket"
	body := fmt.Sprintf("A refund of %.2f for order #%d has been processed.\n\nGoTicket Team", amount, orderID)
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
