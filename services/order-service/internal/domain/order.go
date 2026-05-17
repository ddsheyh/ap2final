package domain

import "time"

type Order struct {
	ID         int64       `json:"id"`
	UserID     int64       `json:"user_id"`
	Status     string      `json:"status"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Items      []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID          int64     `json:"id"`
	OrderID     int64     `json:"order_id"`
	ProductName string    `json:"product_name"`
	Quantity    int32     `json:"quantity"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)
