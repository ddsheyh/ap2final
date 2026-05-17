package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"order-service/internal/domain"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func userOrdersKey(userID int64) string {
	return fmt.Sprintf("user_orders:%d", userID)
}

func orderKey(orderID int64) string {
	return fmt.Sprintf("order:%d", orderID)
}

func (c *RedisCache) SetUserOrders(ctx context.Context, userID int64, orders []*domain.Order) error {
	data, err := json.Marshal(orders)
	if err != nil {
		return fmt.Errorf("marshal orders: %w", err)
	}
	return c.client.Set(ctx, userOrdersKey(userID), data, 10*time.Minute).Err()
}

func (c *RedisCache) GetUserOrders(ctx context.Context, userID int64) ([]*domain.Order, error) {
	data, err := c.client.Get(ctx, userOrdersKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}
	var orders []*domain.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, fmt.Errorf("unmarshal orders: %w", err)
	}
	return orders, nil
}

func (c *RedisCache) InvalidateUserOrders(ctx context.Context, userID int64) error {
	return c.client.Del(ctx, userOrdersKey(userID)).Err()
}

func (c *RedisCache) SetOrder(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("marshal order: %w", err)
	}
	return c.client.Set(ctx, orderKey(order.ID), data, 10*time.Minute).Err()
}

func (c *RedisCache) GetOrder(ctx context.Context, orderID int64) (*domain.Order, error) {
	data, err := c.client.Get(ctx, orderKey(orderID)).Bytes()
	if err != nil {
		return nil, err
	}
	var order domain.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("unmarshal order: %w", err)
	}
	return &order, nil
}

func (c *RedisCache) InvalidateOrder(ctx context.Context, orderID int64) error {
	return c.client.Del(ctx, orderKey(orderID)).Err()
}
