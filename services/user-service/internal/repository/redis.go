package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"user-service/internal/domain"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func userCacheKey(id int64) string {
	return fmt.Sprintf("user:%d", id)
}

func refreshTokenKey(token string) string {
	return fmt.Sprintf("refresh:%s", token)
}

func blacklistKey(token string) string {
	return fmt.Sprintf("blacklist:%s", token)
}

func (c *RedisCache) SetUser(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal user: %w", err)
	}
	return c.client.Set(ctx, userCacheKey(user.ID), data, 15*time.Minute).Err()
}

func (c *RedisCache) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	data, err := c.client.Get(ctx, userCacheKey(id)).Bytes()
	if err != nil {
		return nil, err
	}
	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("unmarshal user: %w", err)
	}
	return &user, nil
}

func (c *RedisCache) DeleteUser(ctx context.Context, id int64) error {
	return c.client.Del(ctx, userCacheKey(id)).Err()
}

func (c *RedisCache) StoreRefreshToken(ctx context.Context, token string, userID int64, ttl time.Duration) error {
	return c.client.Set(ctx, refreshTokenKey(token), userID, ttl).Err()
}

func (c *RedisCache) GetRefreshTokenUserID(ctx context.Context, token string) (int64, error) {
	val, err := c.client.Get(ctx, refreshTokenKey(token)).Int64()
	if err != nil {
		return 0, fmt.Errorf("get refresh token: %w", err)
	}
	return val, nil
}

func (c *RedisCache) DeleteRefreshToken(ctx context.Context, token string) error {
	return c.client.Del(ctx, refreshTokenKey(token)).Err()
}

func (c *RedisCache) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	return c.client.Set(ctx, blacklistKey(token), "1", ttl).Err()
}

func (c *RedisCache) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	exists, err := c.client.Exists(ctx, blacklistKey(token)).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
