package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"payment-service/internal/domain"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, orderID, userID int64, amount float64, currency, method string) (*domain.Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	p := &domain.Payment{}
	err = tx.QueryRow(ctx,
		`INSERT INTO payments (order_id, user_id, amount, currency, status, payment_method)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at`,
		orderID, userID, amount, currency, domain.PaymentPending, method,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert payment: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (payment_id, type, amount, status, description)
		 VALUES ($1, $2, $3, $4, $5)`,
		p.ID, domain.TxTypeCharge, amount, domain.PaymentPending, "Initial charge",
	)
	if err != nil {
		return nil, fmt.Errorf("insert transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return p, nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id int64) (*domain.Payment, error) {
	p := &domain.Payment{}
	err := r.db.QueryRow(ctx,
		`SELECT id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get payment: %w", err)
	}
	return p, nil
}

func (r *PaymentRepository) List(ctx context.Context, offset, limit int) ([]*domain.Payment, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM payments`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at
		 FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID int64) ([]*domain.Payment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at
		 FROM payments WHERE order_id = $1 ORDER BY created_at DESC`, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("get by order: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *PaymentRepository) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*domain.Payment, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at
		 FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		p := &domain.Payment{}
		if err := rows.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		payments = append(payments, p)
	}
	return payments, total, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id int64, newStatus, txType, description string) (*domain.Payment, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	p := &domain.Payment{}
	err = tx.QueryRow(ctx,
		`UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2
		 RETURNING id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at`,
		newStatus, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update status: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO transactions (payment_id, type, amount, status, description) VALUES ($1, $2, $3, $4, $5)`,
		id, txType, p.Amount, newStatus, description,
	)
	if err != nil {
		return nil, fmt.Errorf("insert tx record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return p, nil
}

func (r *PaymentRepository) Refund(ctx context.Context, id int64, amount float64) (*domain.Payment, *domain.Transaction, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	p := &domain.Payment{}
	err = tx.QueryRow(ctx,
		`UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2
		 RETURNING id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at`,
		domain.PaymentRefunded, id,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("update payment: %w", err)
	}

	refundAmount := amount
	if refundAmount <= 0 {
		refundAmount = p.Amount
	}

	t := &domain.Transaction{}
	err = tx.QueryRow(ctx,
		`INSERT INTO transactions (payment_id, type, amount, status, description)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, payment_id, type, amount, status, description, created_at`,
		id, domain.TxTypeRefund, refundAmount, domain.PaymentRefunded, "Refund processed",
	).Scan(&t.ID, &t.PaymentID, &t.Type, &t.Amount, &t.Status, &t.Description, &t.CreatedAt)
	if err != nil {
		return nil, nil, fmt.Errorf("insert refund tx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	return p, t, nil
}

func (r *PaymentRepository) ListTransactions(ctx context.Context, paymentID int64, offset, limit int) ([]*domain.Transaction, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE payment_id = $1`, paymentID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, payment_id, type, amount, status, description, created_at
		 FROM transactions WHERE payment_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		paymentID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list: %w", err)
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := rows.Scan(&t.ID, &t.PaymentID, &t.Type, &t.Amount, &t.Status, &t.Description, &t.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan: %w", err)
		}
		txs = append(txs, t)
	}
	return txs, total, nil
}

func (r *PaymentRepository) GetTransaction(ctx context.Context, id int64) (*domain.Transaction, error) {
	t := &domain.Transaction{}
	err := r.db.QueryRow(ctx,
		`SELECT id, payment_id, type, amount, status, description, created_at FROM transactions WHERE id = $1`, id,
	).Scan(&t.ID, &t.PaymentID, &t.Type, &t.Amount, &t.Status, &t.Description, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	return t, nil
}

func (r *PaymentRepository) GetStats(ctx context.Context, userID int64) (*domain.PaymentStats, error) {
	stats := &domain.PaymentStats{}
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(SUM(amount), 0),
			COUNT(*) FILTER (WHERE status = 'completed'),
			COUNT(*) FILTER (WHERE status = 'failed'),
			COUNT(*) FILTER (WHERE status = 'refunded')
		FROM payments WHERE user_id = $1`, userID,
	).Scan(&stats.TotalPayments, &stats.TotalAmount, &stats.SuccessfulPayments, &stats.FailedPayments, &stats.RefundedPayments)
	if err != nil {
		return nil, fmt.Errorf("get stats: %w", err)
	}
	return stats, nil
}

func (r *PaymentRepository) GetByOrderIDForCancel(ctx context.Context, orderID int64) (*domain.Payment, error) {
	p := &domain.Payment{}
	err := r.db.QueryRow(ctx,
		`SELECT id, order_id, user_id, amount, currency, status, payment_method, created_at, updated_at
		 FROM payments WHERE order_id = $1 AND status = $2 LIMIT 1`,
		orderID, domain.PaymentPending,
	).Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.PaymentMethod, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get pending payment: %w", err)
	}
	return p, nil
}
