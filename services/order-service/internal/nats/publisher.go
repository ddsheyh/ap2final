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

type OrderEvent struct {
	OrderID int64   `json:"order_id"`
	UserID  int64   `json:"user_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

func (p *Publisher) PublishOrderCreated(orderID, userID int64, amount float64) error {
	event := OrderEvent{OrderID: orderID, UserID: userID, Amount: amount, Status: "pending"}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.conn.Publish("order.created", data); err != nil {
		return fmt.Errorf("publish order.created: %w", err)
	}
	log.Printf("Published order.created: orderID=%d userID=%d amount=%.2f", orderID, userID, amount)
	return nil
}

func (p *Publisher) PublishOrderCancelled(orderID, userID int64) error {
	event := OrderEvent{OrderID: orderID, UserID: userID, Status: "cancelled"}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.conn.Publish("order.cancelled", data); err != nil {
		return fmt.Errorf("publish order.cancelled: %w", err)
	}
	log.Printf("Published order.cancelled: orderID=%d", orderID)
	return nil
}

func (p *Publisher) PublishOrderCompleted(orderID, userID int64) error {
	event := OrderEvent{OrderID: orderID, UserID: userID, Status: "completed"}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := p.conn.Publish("order.completed", data); err != nil {
		return fmt.Errorf("publish order.completed: %w", err)
	}
	log.Printf("Published order.completed: orderID=%d", orderID)
	return nil
}
