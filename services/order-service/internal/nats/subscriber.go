package nats

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"

	"order-service/internal/repository"
)

type Subscriber struct {
	conn *nats.Conn
	repo *repository.OrderRepository
}

func NewSubscriber(conn *nats.Conn, repo *repository.OrderRepository) *Subscriber {
	return &Subscriber{conn: conn, repo: repo}
}

type UserDeletedEvent struct {
	UserID int64 `json:"user_id"`
}

func (s *Subscriber) Subscribe() error {
	_, err := s.conn.Subscribe("user.deleted", func(msg *nats.Msg) {
		var event UserDeletedEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Error unmarshaling user.deleted: %v", err)
			return
		}

		log.Printf("Received user.deleted: userID=%d — cancelling all orders", event.UserID)
		if err := s.repo.CancelUserOrders(context.Background(), event.UserID); err != nil {
			log.Printf("Error cancelling user orders: %v", err)
		}
	})
	if err != nil {
		return err
	}

	log.Println("Subscribed to user.deleted")
	return nil
}
