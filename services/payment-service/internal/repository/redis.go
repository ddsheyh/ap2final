package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func paymentStatusKey(orderID int64) string {
	return fmt.Sprintf("payment_status:%d", orderID)
}

func (c *RedisCache) SetPaymentStatus(ctx context.Context, orderID int64, status string) error {
	return c.client.Set(ctx, paymentStatusKey(orderID), status, 10*time.Minute).Err()
}

func (c *RedisCache) GetPaymentStatus(ctx context.Context, orderID int64) (string, error) {
	return c.client.Get(ctx, paymentStatusKey(orderID)).Result()
}

func (c *RedisCache) InvalidatePaymentStatus(ctx context.Context, orderID int64) error {
	return c.client.Del(ctx, paymentStatusKey(orderID)).Err()
}
