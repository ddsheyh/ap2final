package nats

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher struct {
	conn *nats.Conn
}

func NewPublisher(conn *nats.Conn) *Publisher {
	return &Publisher{conn: conn}
}

type PaymentEvent struct {
	PaymentID int64   `json:"payment_id"`
	OrderID   int64   `json:"order_id"`
	UserID    int64   `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

func (p *Publisher) PublishPaymentCompleted(paymentID, orderID, userID int64, amount float64) error {
	return p.publish("payment.completed", PaymentEvent{
		PaymentID: paymentID, OrderID: orderID, UserID: userID, Amount: amount, Status: "completed",
	})
}

func (p *Publisher) PublishPaymentFailed(paymentID, orderID, userID int64) error {
	return p.publish("payment.failed", PaymentEvent{
		PaymentID: paymentID, OrderID: orderID, UserID: userID, Status: "failed",
	})
}

func (p *Publisher) PublishPaymentRefunded(paymentID, orderID, userID int64, amount float64) error {
	return p.publish("payment.refunded", PaymentEvent{
		PaymentID: paymentID, OrderID: orderID, UserID: userID, Amount: amount, Status: "refunded",
	})
}

func (p *Publisher) publish(subject string, event PaymentEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.conn.Publish(subject, data); err != nil {
		return fmt.Errorf("publish %s: %w", subject, err)
	}
	log.Printf("Published %s: paymentID=%d orderID=%d", subject, event.PaymentID, event.OrderID)
	return nil
}
