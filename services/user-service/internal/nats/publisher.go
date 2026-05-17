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

type UserRegisteredEvent struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

type UserDeletedEvent struct {
	UserID int64 `json:"user_id"`
}

func (p *Publisher) PublishUserRegistered(userID int64, email, name string) error {
	event := UserRegisteredEvent{UserID: userID, Email: email, Name: name}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := p.conn.Publish("user.registered", data); err != nil {
		return fmt.Errorf("publish user.registered: %w", err)
	}
	log.Printf("Published user.registered: userID=%d", userID)
	return nil
}

func (p *Publisher) PublishUserDeleted(userID int64) error {
	event := UserDeletedEvent{UserID: userID}
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}
	if err := p.conn.Publish("user.deleted", data); err != nil {
		return fmt.Errorf("publish user.deleted: %w", err)
	}
	log.Printf("Published user.deleted: userID=%d", userID)
	return nil
}
