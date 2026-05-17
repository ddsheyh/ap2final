package usecase

import (
	"context"
	"fmt"
	"log"

	"order-service/internal/domain"
	natspub "order-service/internal/nats"
	"order-service/internal/repository"
)

type OrderUsecase struct {
	repo      *repository.OrderRepository
	cache     *repository.RedisCache
	publisher *natspub.Publisher
}

func NewOrderUsecase(repo *repository.OrderRepository, cache *repository.RedisCache, publisher *natspub.Publisher) *OrderUsecase {
	return &OrderUsecase{repo: repo, cache: cache, publisher: publisher}
}

func (uc *OrderUsecase) CreateOrder(ctx context.Context, userID int64, items []domain.OrderItem) (*domain.Order, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}

	order, err := uc.repo.CreateWithItems(ctx, userID, items)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateUserOrders(ctx, userID)

	if uc.publisher != nil {
		if err := uc.publisher.PublishOrderCreated(order.ID, userID, order.TotalPrice); err != nil {
			log.Printf("Failed to publish order.created: %v", err)
		}
	}

	return order, nil
}

func (uc *OrderUsecase) GetOrder(ctx context.Context, id int64) (*domain.Order, error) {
	if cached, err := uc.cache.GetOrder(ctx, id); err == nil {
		return cached, nil
	}

	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.SetOrder(ctx, order)
	return order, nil
}

func (uc *OrderUsecase) ListOrders(ctx context.Context, page, pageSize int) ([]*domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return uc.repo.List(ctx, offset, pageSize)
}

func (uc *OrderUsecase) UpdateOrder(ctx context.Context, id int64, status string) (*domain.Order, error) {
	order, err := uc.repo.Update(ctx, id, status)
	if err != nil {
		return nil, err
	}
	_ = uc.cache.InvalidateOrder(ctx, id)
	_ = uc.cache.InvalidateUserOrders(ctx, order.UserID)
	return order, nil
}

func (uc *OrderUsecase) DeleteOrder(ctx context.Context, id int64) error {
	order, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}
	_ = uc.cache.InvalidateOrder(ctx, id)
	_ = uc.cache.InvalidateUserOrders(ctx, order.UserID)
	return nil
}

func (uc *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) (*domain.Order, error) {
	order, err := uc.repo.Update(ctx, orderID, newStatus)
	if err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateOrder(ctx, orderID)
	_ = uc.cache.InvalidateUserOrders(ctx, order.UserID)

	if uc.publisher != nil {
		switch newStatus {
		case domain.StatusCompleted:
			_ = uc.publisher.PublishOrderCompleted(orderID, order.UserID)
		case domain.StatusCancelled:
			_ = uc.publisher.PublishOrderCancelled(orderID, order.UserID)
		}
	}

	return order, nil
}

func (uc *OrderUsecase) CancelOrder(ctx context.Context, orderID int64) (*domain.Order, error) {
	return uc.UpdateOrderStatus(ctx, orderID, domain.StatusCancelled)
}

func (uc *OrderUsecase) GetOrdersByUser(ctx context.Context, userID int64, page, pageSize int) ([]*domain.Order, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return uc.repo.GetByUserID(ctx, userID, offset, pageSize)
}

func (uc *OrderUsecase) AddOrderItem(ctx context.Context, orderID int64, productName string, quantity int32, price float64) (*domain.OrderItem, error) {
	item, err := uc.repo.AddItem(ctx, orderID, productName, quantity, price)
	if err != nil {
		return nil, err
	}
	_ = uc.cache.InvalidateOrder(ctx, orderID)
	return item, nil
}

func (uc *OrderUsecase) RemoveOrderItem(ctx context.Context, itemID int64) error {
	if err := uc.repo.RemoveItem(ctx, itemID); err != nil {
		return err
	}
	return nil
}

func (uc *OrderUsecase) GetOrderItems(ctx context.Context, orderID int64) ([]domain.OrderItem, error) {
	return uc.repo.GetItems(ctx, orderID)
}

func (uc *OrderUsecase) GetOrderTotal(ctx context.Context, orderID int64) (float64, int32, error) {
	return uc.repo.GetOrderTotal(ctx, orderID)
}
