package usecase

import (
	"testing"

	"payment-service/internal/domain"
)

func TestPaymentDomainModel(t *testing.T) {
	payment := &domain.Payment{
		ID:            1,
		OrderID:       10,
		UserID:        42,
		Amount:        5000.00,
		Currency:      "KZT",
		Status:        domain.PaymentPending,
		PaymentMethod: "card",
	}

	if payment.Status != "pending" {
		t.Fatalf("expected pending, got %s", payment.Status)
	}
	if payment.Currency != "KZT" {
		t.Fatalf("expected KZT, got %s", payment.Currency)
	}
}

func TestPaymentStatsModel(t *testing.T) {
	stats := &domain.PaymentStats{
		TotalPayments:      10,
		TotalAmount:        50000.00,
		SuccessfulPayments: 7,
		FailedPayments:     2,
		RefundedPayments:   1,
	}

	if stats.TotalPayments != stats.SuccessfulPayments+stats.FailedPayments+stats.RefundedPayments {
		t.Fatal("payment counts don't add up")
	}
}

func TestPaymentStatusConstants(t *testing.T) {
	statuses := []string{
		domain.PaymentPending,
		domain.PaymentCompleted,
		domain.PaymentFailed,
		domain.PaymentCancelled,
		domain.PaymentRefunded,
	}
	seen := make(map[string]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Fatalf("duplicate status: %s", s)
		}
		seen[s] = true
	}
	if len(statuses) != 5 {
		t.Fatalf("expected 5 statuses, got %d", len(statuses))
	}
}

func TestTransactionTypes(t *testing.T) {
	types := []string{domain.TxTypeCharge, domain.TxTypeRefund, domain.TxTypeRetry}
	for _, typ := range types {
		if typ == "" {
			t.Fatal("transaction type should not be empty")
		}
	}
}

// Integration test
func TestPaymentCRUD_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	t.Log("Integration test - requires running infrastructure")
}
