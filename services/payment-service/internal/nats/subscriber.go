package nats

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"payment-service/internal/domain"
	"payment-service/internal/repository"
)

// Subscriber listens for NATS events relevant to payments.
type Subscriber struct {
	conn *nats.Conn
	repo *repository.PaymentRepository
}

// NewSubscriber creates a new NATS subscriber.
func NewSubscriber(conn *nats.Conn, repo *repository.PaymentRepository) *Subscriber {
	return &Subscriber{conn: conn, repo: repo}
}

// OrderCreatedEvent represents the payload from order.created.
type OrderCreatedEvent struct {
	OrderID int64   `json:"order_id"`
	UserID  int64   `json:"user_id"`
	Amount  float64 `json:"amount"`
}

// OrderCancelledEvent represents the payload from order.cancelled.
type OrderCancelledEvent struct {
	OrderID int64 `json:"order_id"`
	UserID  int64 `json:"user_id"`
}

// Subscribe starts listening for order events.
func (s *Subscriber) Subscribe() error {
	// Subscribe to order.created — auto-create payment
	_, err := s.conn.Subscribe("order.created", func(msg *nats.Msg) {
		var event OrderCreatedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling order.created: %v", err)
			return
		}

		log.Printf("Received order.created: orderID=%d userID=%d amount=%.2f", event.OrderID, event.UserID, event.Amount)
		_, err := s.repo.Create(context.Background(), event.OrderID, event.UserID, event.Amount, "KZT", "card")
		if err != nil {
			log.Printf("Error creating payment for order %d: %v", event.OrderID, err)
		}
	})
	if err != nil {
		return err
	}
	log.Println("Subscribed to order.created")

	// Subscribe to order.cancelled — auto-refund payment
	_, err = s.conn.Subscribe("order.cancelled", func(msg *nats.Msg) {
		var event OrderCancelledEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling order.cancelled: %v", err)
			return
		}

		log.Printf("Received order.cancelled: orderID=%d — refunding payment", event.OrderID)
		payment, err := s.repo.GetByOrderIDForCancel(context.Background(), event.OrderID)
		if err != nil {
			log.Printf("No pending payment found for order %d: %v", event.OrderID, err)
			return
		}

		_, _, err = s.repo.Refund(context.Background(), payment.ID, 0)
		if err != nil {
			log.Printf("Error refunding payment %d: %v", payment.ID, err)
			return
		}

		// Also cancel completed payments
		s.repo.UpdateStatus(context.Background(), payment.ID, domain.PaymentCancelled, "cancel", "Order cancelled")
	})
	if err != nil {
		return err
	}
	log.Println("Subscribed to order.cancelled")

	return nil
}
