package usecase

import (
	"context"
	"fmt"
	"log"

	"payment-service/internal/domain"
	"payment-service/internal/email"
	natspub "payment-service/internal/nats"
	"payment-service/internal/repository"
)

type PaymentUsecase struct {
	repo      *repository.PaymentRepository
	cache     *repository.RedisCache
	publisher *natspub.Publisher
	mailer    *email.SMTPSender
}

func NewPaymentUsecase(repo *repository.PaymentRepository, cache *repository.RedisCache, publisher *natspub.Publisher, mailer *email.SMTPSender) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, cache: cache, publisher: publisher, mailer: mailer}
}

func (uc *PaymentUsecase) CreatePayment(ctx context.Context, orderID, userID int64, amount float64, currency, method string) (*domain.Payment, error) {
	if currency == "" {
		currency = "KZT"
	}
	if method == "" {
		method = "card"
	}

	payment, err := uc.repo.Create(ctx, orderID, userID, amount, currency, method)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetPaymentStatus(ctx, orderID, payment.Status)
	return payment, nil
}

func (uc *PaymentUsecase) GetPayment(ctx context.Context, id int64) (*domain.Payment, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *PaymentUsecase) ListPayments(ctx context.Context, page, pageSize int) ([]*domain.Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.repo.List(ctx, (page-1)*pageSize, pageSize)
}

func (uc *PaymentUsecase) GetPaymentsByOrder(ctx context.Context, orderID int64) ([]*domain.Payment, error) {
	return uc.repo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUsecase) UpdatePaymentStatus(ctx context.Context, paymentID int64, newStatus string) (*domain.Payment, error) {
	txType := domain.TxTypeCharge
	description := fmt.Sprintf("Status changed to %s", newStatus)

	payment, err := uc.repo.UpdateStatus(ctx, paymentID, newStatus, txType, description)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetPaymentStatus(ctx, payment.OrderID, newStatus)

	if uc.publisher != nil {
		switch newStatus {
		case domain.PaymentCompleted:
			_ = uc.publisher.PublishPaymentCompleted(payment.ID, payment.OrderID, payment.UserID, payment.Amount)
		case domain.PaymentFailed:
			_ = uc.publisher.PublishPaymentFailed(payment.ID, payment.OrderID, payment.UserID)
		}
	}

	return payment, nil
}

func (uc *PaymentUsecase) CancelPayment(ctx context.Context, paymentID int64) (*domain.Payment, error) {
	payment, err := uc.repo.UpdateStatus(ctx, paymentID, domain.PaymentCancelled, "cancel", "Payment cancelled")
	if err != nil {
		return nil, err
	}
	_ = uc.cache.InvalidatePaymentStatus(ctx, payment.OrderID)
	return payment, nil
}

func (uc *PaymentUsecase) RefundPayment(ctx context.Context, paymentID int64, amount float64) (*domain.Payment, *domain.Transaction, error) {
	payment, tx, err := uc.repo.Refund(ctx, paymentID, amount)
	if err != nil {
		return nil, nil, err
	}

	_ = uc.cache.SetPaymentStatus(ctx, payment.OrderID, domain.PaymentRefunded)

	if uc.publisher != nil {
		_ = uc.publisher.PublishPaymentRefunded(payment.ID, payment.OrderID, payment.UserID, tx.Amount)
	}

	if uc.mailer != nil {
		go func() {
			if err := uc.mailer.SendRefundConfirmation("user@example.com", payment.OrderID, tx.Amount); err != nil {
				log.Printf("Failed to send refund email: %v", err)
			}
		}()
	}

	return payment, tx, nil
}

func (uc *PaymentUsecase) GetPaymentsByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Payment, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.repo.GetByUserID(ctx, userID, (page-1)*pageSize, pageSize)
}

func (uc *PaymentUsecase) ListTransactions(ctx context.Context, paymentID int64, page, pageSize int) ([]*domain.Transaction, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return uc.repo.ListTransactions(ctx, paymentID, (page-1)*pageSize, pageSize)
}

func (uc *PaymentUsecase) GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error) {
	return uc.repo.GetTransaction(ctx, id)
}

func (uc *PaymentUsecase) GetPaymentStats(ctx context.Context, userID int64) (*domain.PaymentStats, error) {
	return uc.repo.GetStats(ctx, userID)
}

func (uc *PaymentUsecase) RetryPayment(ctx context.Context, paymentID int64) (*domain.Payment, error) {
	payment, err := uc.repo.GetByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	if payment.Status != domain.PaymentFailed {
		return nil, fmt.Errorf("can only retry failed payments, current status: %s", payment.Status)
	}

	updated, err := uc.repo.UpdateStatus(ctx, paymentID, domain.PaymentPending, domain.TxTypeRetry, "Payment retry initiated")
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetPaymentStatus(ctx, payment.OrderID, domain.PaymentPending)
	return updated, nil
}
