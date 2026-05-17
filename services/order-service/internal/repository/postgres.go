package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-service/internal/domain"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateWithItems(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}

	order := &domain.Order{}
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (user_id, status, total_price) VALUES ($1, $2, $3)
		 RETURNING id, user_id, status, total_price, created_at, updated_at`,
		userID, domain.StatusPending, total,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	for _, item := range items {
		var oi domain.OrderItem
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, product_name, quantity, price)
			 VALUES ($1, $2, $3, $4) RETURNING id, order_id, product_name, quantity, price, created_at`,
			order.ID, item.ProductName, item.Quantity, item.Price,
		).Scan(&oi.ID, &oi.OrderID, &oi.ProductName, &oi.Quantity, &oi.Price, &oi.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert order item: %w", err)
		}
		order.Items = append(order.Items, oi)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, status, total_price, created_at, updated_at FROM orders WHERE id = $1`, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	items, err := r.GetItems(ctx, id)
	if err == nil {
		order.Items = items
	}

	return order, nil
}

func (r *OrderRepository) List(ctx context.Context, offset, limit int) ([]*domain.Order, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, status, total_price, created_at, updated_at FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows, total)
}

func (r *OrderRepository) GetByUserID(ctx context.Context, userID int64, offset, limit int) ([]*domain.Order, int, error) {
	var total int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count user orders: %w", err)
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, status, total_price, created_at, updated_at FROM orders WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("list user orders: %w", err)
	}
	defer rows.Close()

	return scanOrders(rows, total)
}

func scanOrders(rows pgx.Rows, total int) ([]*domain.Order, int, error) {
	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(&o.ID, &o.UserID, &o.Status, &o.TotalPrice, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, total, nil
}

func (r *OrderRepository) Update(ctx context.Context, id int64, status string) (*domain.Order, error) {
	order := &domain.Order{}
	err := r.db.QueryRow(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
		 RETURNING id, user_id, status, total_price, created_at, updated_at`,
		status, id,
	).Scan(&order.ID, &order.UserID, &order.Status, &order.TotalPrice, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update order: %w", err)
	}
	return order, nil
}

func (r *OrderRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order not found")
	}
	return nil
}

func (r *OrderRepository) CancelUserOrders(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE user_id = $2 AND status NOT IN ($3, $4)`,
		domain.StatusCancelled, userID, domain.StatusCancelled, domain.StatusCompleted,
	)
	return err
}

func (r *OrderRepository) AddItem(ctx context.Context, orderID int64, productName string, quantity int32, price float64) (*domain.OrderItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	item := &domain.OrderItem{}
	err = tx.QueryRow(ctx,
		`INSERT INTO order_items (order_id, product_name, quantity, price) VALUES ($1, $2, $3, $4)
		 RETURNING id, order_id, product_name, quantity, price, created_at`,
		orderID, productName, quantity, price,
	).Scan(&item.ID, &item.OrderID, &item.ProductName, &item.Quantity, &item.Price, &item.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("add item: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders SET total_price = (SELECT COALESCE(SUM(price * quantity), 0) FROM order_items WHERE order_id = $1), updated_at = NOW() WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("update total: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return item, nil
}

func (r *OrderRepository) RemoveItem(ctx context.Context, itemID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var orderID int64
	err = tx.QueryRow(ctx, `DELETE FROM order_items WHERE id = $1 RETURNING order_id`, itemID).Scan(&orderID)
	if err != nil {
		return fmt.Errorf("remove item: %w", err)
	}

	_, err = tx.Exec(ctx,
		`UPDATE orders SET total_price = (SELECT COALESCE(SUM(price * quantity), 0) FROM order_items WHERE order_id = $1), updated_at = NOW() WHERE id = $1`,
		orderID,
	)
	if err != nil {
		return fmt.Errorf("update total: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) GetItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, order_id, product_name, quantity, price, created_at FROM order_items WHERE order_id = $1`,
		orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductName, &item.Quantity, &item.Price, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *OrderRepository) GetOrderTotal(ctx context.Context, orderID int64) (float64, int32, error) {
	var total float64
	var count int32
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(price * quantity), 0), COALESCE(COUNT(*), 0) FROM order_items WHERE order_id = $1`,
		orderID,
	).Scan(&total, &count)
	if err != nil {
		return 0, 0, fmt.Errorf("get total: %w", err)
	}
	return total, count, nil
}
