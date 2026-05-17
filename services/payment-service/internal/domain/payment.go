package domain

import "time"

type Payment struct {
	ID            int64     `json:"id"`
	OrderID       int64     `json:"order_id"`
	UserID        int64     `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Transaction struct {
	ID          int64     `json:"id"`
	PaymentID   int64     `json:"payment_id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentStats struct {
	TotalPayments      int32   `json:"total_payments"`
	TotalAmount        float64 `json:"total_amount"`
	SuccessfulPayments int32   `json:"successful_payments"`
	FailedPayments     int32   `json:"failed_payments"`
	RefundedPayments   int32   `json:"refunded_payments"`
}

const (
	PaymentPending   = "pending"
	PaymentCompleted = "completed"
	PaymentFailed    = "failed"
	PaymentCancelled = "cancelled"
	PaymentRefunded  = "refunded"
)

const (
	TxTypeCharge = "charge"
	TxTypeRefund = "refund"
	TxTypeRetry  = "retry"
)
